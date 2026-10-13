TF_DIR := infrastructure/terraform/environments/dev
ANSIBLE_DIR := ansible
PYTHON := .venv/bin/python

.DEFAULT_GOAL := help

.PHONY: help bootstrap infra-plan infra-apply ssh-bootstrap ping configure configure-check destroy

help: ## Show available commands and what they do
	@printf 'Usage: make <command>\n\nCommands:\n'
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

bootstrap: ## Apply dev infrastructure, verify SSH host keys, and check Ansible connectivity
	$(MAKE) infra-apply
	$(MAKE) ssh-bootstrap
	$(MAKE) ping

infra-plan: ## Preview changes to the dev infrastructure with Terraform
	cd $(TF_DIR) && terraform plan

infra-apply: ## Create or update the dev infrastructure with Terraform
	cd $(TF_DIR) && terraform apply

ssh-bootstrap: ## Verify EC2 SSH host keys against AWS and save them to known_hosts
	$(PYTHON) scripts/sync-ssh-hostkeys.py

ping: ## Check SSH access and Python availability on all Ansible inventory hosts
	cd $(ANSIBLE_DIR) && ../.venv/bin/ansible all -m ansible.builtin.ping

configure: ## Configure inventory hosts using the Ansible site playbook
	cd $(ANSIBLE_DIR) && ../.venv/bin/ansible-playbook playbooks/site.yml

configure-check: ## Preview supported Ansible changes and show diffs using check mode
	cd $(ANSIBLE_DIR) && ../.venv/bin/ansible-playbook playbooks/site.yml --check --diff

destroy: ## Destroy all resources managed by the dev Terraform configuration
	cd $(TF_DIR) && terraform destroy
