TF_DIR := infrastructure/terraform/environments/dev
ANSIBLE_DIR := ansible
PYTHON := .venv/bin/python

.PHONY: infra-plan infra-apply ssh-bootstrap ping destroy

infra-plan:
	cd $(TF_DIR) && terraform plan

infra-apply:
	cd $(TF_DIR) && terraform apply

ssh-bootstrap:
	$(PYTHON) scripts/sync-ssh-hostkeys.py

ping:
	cd $(ANSIBLE_DIR) && ../.venv/bin/ansible all -m ansible.builtin.ping

destroy:
	cd $(TF_DIR) && terraform destroy
