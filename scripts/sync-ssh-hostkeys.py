#!/usr/bin/env python3
"""Enroll AWS-verified SSH host keys for disposable EC2 instances."""

import base64
import hashlib
import os
import re
import subprocess
import tempfile
import time
from pathlib import Path

import boto3

REGION = "eu-north-1"
PROFILE = "aws-learning"
PROJECT = "cloud-reliability"
ENVIRONMENT = "dev"
ROLES = {"application", "prometheus", "grafana"}

KNOWN_HOSTS = Path.home() / ".ssh/cloud-reliability/known_hosts"

FINGERPRINT_PATTERN = re.compile(r"SHA256:([A-Za-z0-9+/]+)" r"[^\n]*\(ED25519\)")


def fingerprint(public_key: str) -> str:
    """Calculate the OpenSSH SHA256 public-key fingerprint."""
    fields = public_key.split()

    if len(fields) < 3 or fields[1] != "ssh-ed25519":
        raise ValueError("Expected an ED25519 SSH host key")

    raw_key = base64.b64decode(fields[2], validate=True)

    digest = hashlib.sha256(raw_key).digest()
    encoded = base64.b64encode(digest).decode().rstrip("=")

    return f"SHA256:{encoded}"


def discover_instances(ec2):
    response = ec2.describe_instances(
        Filters=[
            {"Name": "tag:Project", "Values": [PROJECT]},
            {"Name": "tag:Environment", "Values": [ENVIRONMENT]},
            {"Name": "instance-state-name", "Values": ["running"]},
        ]
    )

    instances = {}

    for reservation in response["Reservations"]:
        for instance in reservation["Instances"]:
            tags = {tag["Key"]: tag["Value"] for tag in instance.get("Tags", [])}

            role = tags.get("Role")

            if role not in ROLES:
                continue

            if role in instances:
                raise RuntimeError(f"Duplicate instance role: {role}")

            instances[role] = {
                "id": instance["InstanceId"],
                "ip": instance["PublicIpAddress"],
            }

    if set(instances) != ROLES:
        raise RuntimeError(f"Expected {ROLES}, found {set(instances)}")

    return instances


def console_fingerprint(ec2, instance_id):
    for attempt in range(12):
        response = ec2.get_console_output(
            InstanceId=instance_id,
            Latest=True,
        )

        output = response.get("Output") or ""
        match = FINGERPRINT_PATTERN.search(output)

        if match:
            return f"SHA256:{match.group(1)}"

        if attempt < 11:
            time.sleep(10)

    raise RuntimeError(f"ED25519 fingerprint unavailable for {instance_id}")


def scan_host_key(ip):
    result = subprocess.run(
        ["ssh-keyscan", "-T", "5", "-t", "ed25519", ip],
        capture_output=True,
        text=True,
        timeout=15,
        check=True,
    )

    for line in result.stdout.splitlines():
        if line.startswith(f"{ip} ssh-ed25519 "):
            return line

    raise RuntimeError(f"No ED25519 host key received from {ip}")


def main():
    ec2 = boto3.Session(
        profile_name=PROFILE,
        region_name=REGION,
    ).client("ec2")

    instances = discover_instances(ec2)
    verified_keys = []

    for role, instance in sorted(instances.items()):
        instance_id = instance["id"]
        ip = instance["ip"]

        print(f"Verifying {role}: {instance_id} ({ip})")

        expected = console_fingerprint(ec2, instance_id)

        # Retry while the new VM finishes booting.
        for attempt in range(12):
            try:
                host_key = scan_host_key(ip)
                break
            except (RuntimeError, subprocess.SubprocessError):
                if attempt == 11:
                    raise
                time.sleep(5)

        actual = fingerprint(host_key)

        if actual != expected:
            raise RuntimeError(
                f"Host-key mismatch for {role}: " f"expected {expected}, got {actual}"
            )

        verified_keys.append(host_key)
        print(f"  VERIFIED: {actual}")

    KNOWN_HOSTS.parent.mkdir(
        parents=True,
        exist_ok=True,
    )

    # Write only after all three instances have passed verification.
    fd, temporary_path = tempfile.mkstemp(
        prefix=".known_hosts-",
        dir=KNOWN_HOSTS.parent,
    )

    try:
        with os.fdopen(fd, "w") as file:
            file.write("\n".join(verified_keys) + "\n")

        os.chmod(temporary_path, 0o600)
        os.replace(temporary_path, KNOWN_HOSTS)
    finally:
        if os.path.exists(temporary_path):
            os.unlink(temporary_path)

    print(f"Verified host keys saved to {KNOWN_HOSTS}")


if __name__ == "__main__":
    main()
