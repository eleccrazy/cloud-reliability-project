# Cloud Reliability Project

A cloud reliability lab for infrastructure provisioning, configuration management,
Go service instrumentation, and SRE operations. The repository currently contains
the directory scaffold; application and deployment components are not implemented yet.

## Repository layout

- `app/`: Go backend. Unit tests live alongside source files as `*_test.go`.
- `infrastructure/terraform/`: AWS modules and development environment.
- `ansible/`: Inventory, playbooks, and roles for three Linux VMs.
- `observability/`: Canonical Prometheus rules, Grafana dashboards and data sources,
  SLI definitions, SLO targets, and error budget policy.
- `tests/`: Integration, load, and failure injection tests.
- `scripts/`: Deployment, verification, and cleanup orchestration.
- `docs/`: Architecture, decisions, runbooks, incidents, and cost analysis.

## Configuration ownership

Maintain observability configuration in `observability/`. Ansible roles deploy
those files rather than maintaining duplicate copies. Use Ansible templates for
host-specific configuration, such as Prometheus scrape targets using instance
private IP addresses and application systemd units.

Add `tasks/`, `handlers/`, `templates/`, `defaults/`, and `files/` to roles as their
implementation requires them. Empty directories are tracked with `.gitkeep`;
implementation files are created when they have meaningful contents. Generate
`app/go.sum` through Go tooling when dependencies require it.

## Development commands

The root Makefile reserves `build`, `test`, `infra-plan`, `infra-apply`,
`configure`, `verify`, and `destroy`. These commands currently report that their
implementation is pending and exit with an error.
