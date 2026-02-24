.PHONY: init plan show deploy deploy-ci destroy help

# Couleurs
GREEN  = \033[0;32m
YELLOW = \033[0;33m
RESET  = \033[0m

# Variables
BUCKET     := $(shell cd terraform && terraform output -raw secret_bucket 2>/dev/null)
DEPLOY_USER ?= manual

help:
	@echo "$(GREEN)Commandes disponibles :$(RESET)"
	@echo "  make init       - Initialise Terraform"
	@echo "  make plan       - Affiche le plan Terraform"
	@echo "  make show       - Affiche l'état Terraform"
	@echo "  make deploy     - Lance l'infra + Ansible (interactif)"
	@echo "  make deploy-ci  - Lance l'infra + Ansible (CI/CD)"
	@echo "  make destroy    - Détruit l'infra (garde KMS + S3)"

init:
	@echo "$(YELLOW)Initialisation Terraform...$(RESET)"
	cd terraform && terraform init

plan:
	@echo "$(YELLOW)Plan Terraform...$(RESET)"
	cd terraform && terraform plan

show:
	@echo "$(YELLOW)État Terraform...$(RESET)"
	cd terraform && terraform show

destroy:
	@echo "$(YELLOW)Destruction de l'infra (KMS + S3 conservés)...$(RESET)"
	cd terraform && terraform destroy \
	  -target=aws_instance.my_alma_server \
	  -target=aws_security_group.ssh_access \
	  -target=aws_key_pair.admin_key \
	  -target=local_file.ansible_inventory \
	  -compact-warnings \
	  -auto-approve

deploy:
	@echo "$(YELLOW)Déploiement de l'infra...$(RESET)"
	cd terraform && terraform apply -auto-approve
	@echo "$(YELLOW)Récupération des secrets...$(RESET)"
	aws s3 cp s3://$(BUCKET)/secrets.yml ansible/secrets.yml || true
	@echo "$(YELLOW)Lancement du playbook Ansible...$(RESET)"
	@read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	echo "$$ANSIBLE_VAULT_PASSWORD" > ~/.vault_pass && \
	chmod 600 ~/.vault_pass && \
	cd ansible && ansible-playbook setup_alma.yml \
	  -e "kms_key_id=$$(cd ../terraform && terraform output -raw kms_key_id)" \
	  -e "ansible_vault_password=$$ANSIBLE_VAULT_PASSWORD" \
	  -e "aws_account_id=$$(cd ../terraform && terraform output -raw aws_account_id)" \
	  -e "deploy_user=$(DEPLOY_USER)" \
	  --vault-password-file ~/.vault_pass
	@echo "$(YELLOW)Upload secrets...$(RESET)"
	aws s3 cp ansible/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(GREEN)Déployé par $(DEPLOY_USER) — terminé !$(RESET)"

deploy-ci:
	@echo "$(YELLOW)Déployé par : $(DEPLOY_USER)$(RESET)"
	cd terraform && terraform apply -auto-approve
	@echo "$(YELLOW)Récupération des secrets...$(RESET)"
	aws s3 cp s3://$(BUCKET)/secrets.yml ansible/secrets.yml || true
	cd ansible && ansible-playbook setup_alma.yml \
	  -e "kms_key_id=$$(cd ../terraform && terraform output -raw kms_key_id)" \
	  -e "ansible_vault_password=$$ANSIBLE_VAULT_PASSWORD" \
	  -e "aws_account_id=$$(cd ../terraform && terraform output -raw aws_account_id)" \
	  -e "deploy_user=$$DEPLOY_USER" \
	  --vault-password-file ~/.vault_pass
	@echo "$(YELLOW)Upload secrets...$(RESET)"
	aws s3 cp ansible/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(GREEN)Déployé par $$DEPLOY_USER — terminé !$(RESET)"
