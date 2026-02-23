.PHONY: init plan show deploy destroy help

# Couleurs
GREEN  = \033[0;32m
YELLOW = \033[0;33m
RESET  = \033[0m

help:
	@echo "$(GREEN)Commandes disponibles :$(RESET)"
	@echo "  make init     - Initialise Terraform"
	@echo "  make plan     - Affiche le plan Terraform"
	@echo "  make show     - Affiche l'état Terraform"
	@echo "  make deploy   - Lance l'infra + Ansible"
	@echo "  make destroy  - Détruit l'infra (garde KMS)"

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
	@echo "$(YELLOW)Destruction de l'infra (KMS conservé)...$(RESET)"
	cd terraform && terraform destroy \
	  -target=aws_instance.my_alma_server \
	  -target=aws_security_group.ssh_access \
	  -target=aws_key_pair.admin_key \
	  -target=local_file.ansible_inventory \
	  -auto-approve

deploy:
	@echo "$(YELLOW)Déploiement de l'infra...$(RESET)"
	cd terraform && terraform apply -auto-approve
	@echo "$(YELLOW)Lancement du playbook Ansible...$(RESET)"
	@read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	echo "$$ANSIBLE_VAULT_PASSWORD" > ~/.vault_pass && \
	chmod 600 ~/.vault_pass && \
	cd ansible && ansible-playbook setup_alma.yml \
	  -e "kms_key_id=$$(cd ../terraform && terraform output -raw kms_key_id)" \
	  -e "ansible_vault_password=$$ANSIBLE_VAULT_PASSWORD" \
	  -e "aws_account_id=$$(cd ../terraform && terraform output -raw aws_account_id)" \
	  --vault-password-file ~/.vault_pass
	@echo "$(GREEN)Déploiement terminé !$(RESET)"
