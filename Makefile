.PHONY: help bootstrap init plan show output validate fmt deploy deploy-ci deploy-debug ansible destroy destroy-full clean

# ─── Couleurs ────────────────────────────────────────────────────────────────
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
CYAN   := \033[0;36m
RESET  := \033[0m

# ─── Variables ───────────────────────────────────────────────────────────────
BUCKET      := transcendance-secrets-43783683331
REGION      := eu-north-1
TF_DIR      := terraform
ANSIBLE_DIR := ansible
PLAYBOOK    := setup_alma.yml
VAULT_FILE  := ~/.vault_pass
DEPLOY_USER ?= manual

# ─── Helpers internes ────────────────────────────────────────────────────────
define tf_output
$(shell cd $(TF_DIR) && terraform output -raw $(1))
endef

define write_vault_pass
	echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && chmod 600 $(VAULT_FILE)
endef

define run_ansible
	cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	  -e "kms_key_id=$(call tf_output,kms_key_id)" \
	  -e "aws_account_id=$(call tf_output,aws_account_id)" \
	  -e "deploy_user=$(DEPLOY_USER)" \
	  --vault-password-file $(VAULT_FILE) \
	  $(ANSIBLE_EXTRA_ARGS)
endef

# ─────────────────────────────────────────────────────────────────────────────

help: ## Affiche l'aide
	@echo ""
	@echo "  $(CYAN)Transcendance — Makefile$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*##"}; {printf "  $(GREEN)%-18s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ─── Bootstrap ───────────────────────────────────────────────────────────────

bootstrap: ## Crée le bucket S3 de secrets (une seule fois)
	@echo "$(YELLOW)Création du bucket S3 permanent...$(RESET)"
	@aws s3 mb s3://$(BUCKET) --region $(REGION) || true
	@aws s3api put-bucket-versioning \
	  --bucket $(BUCKET) \
	  --versioning-configuration Status=Enabled || true
	@aws s3api put-bucket-encryption \
	  --bucket $(BUCKET) \
	  --server-side-encryption-configuration \
	    '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}' || true
	@aws s3api put-public-access-block \
	  --bucket $(BUCKET) \
	  --public-access-block-configuration \
	    "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true" || true
	@echo "$(GREEN)Bootstrap terminé — bucket $(BUCKET) prêt !$(RESET)"

# ─── Terraform ───────────────────────────────────────────────────────────────

init: ## Initialise Terraform
	@echo "$(YELLOW)Initialisation Terraform...$(RESET)"
	cd $(TF_DIR) && terraform init

init-upgrade: ## Met à jour les providers Terraform
	@echo "$(YELLOW)Mise à jour des providers...$(RESET)"
	cd $(TF_DIR) && terraform init -upgrade

fmt: ## Formate les fichiers Terraform
	@echo "$(YELLOW)Formatage Terraform...$(RESET)"
	cd $(TF_DIR) && terraform fmt -recursive

validate: ## Valide la configuration Terraform
	@echo "$(YELLOW)Validation Terraform...$(RESET)"
	cd $(TF_DIR) && terraform validate

plan: ## Affiche le plan Terraform
	@echo "$(YELLOW)Plan Terraform...$(RESET)"
	cd $(TF_DIR) && terraform plan

show: ## Affiche l'état Terraform
	@echo "$(YELLOW)État Terraform...$(RESET)"
	cd $(TF_DIR) && terraform show

output: ## Affiche les outputs Terraform
	@echo "$(YELLOW)Outputs Terraform...$(RESET)"
	cd $(TF_DIR) && terraform output

# ─── Déploiement ─────────────────────────────────────────────────────────────

deploy: init ## Déploie l'infra + lance Ansible (interactif)
	@echo "$(YELLOW)Application du plan Terraform...$(RESET)"
	cd $(TF_DIR) && terraform apply -auto-approve
	@echo "$(YELLOW)Récupération des secrets S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(ANSIBLE_DIR)/secrets.yml || true
	@echo "$(YELLOW)Lancement du playbook Ansible...$(RESET)"
	@read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE) && \
	  cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	    -e "kms_key_id=$$(cd ../$(TF_DIR) && terraform output -raw kms_key_id)" \
	    -e "aws_account_id=$$(cd ../$(TF_DIR) && terraform output -raw aws_account_id)" \
	    -e "deploy_user=$(DEPLOY_USER)" \
	    --vault-password-file $(VAULT_FILE)
	@echo "$(YELLOW)Upload des secrets vers S3...$(RESET)"
	@aws s3 cp $(ANSIBLE_DIR)/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(GREEN)Déployé par $(DEPLOY_USER) — terminé !$(RESET)"

deploy-ci: bootstrap init ## Déploie l'infra + lance Ansible (CI/CD, sans prompt)
	@echo "$(YELLOW)Déployé par : $(DEPLOY_USER)$(RESET)"
	@if [ -z "$$ANSIBLE_VAULT_PASSWORD" ]; then echo "$(RED)ANSIBLE_VAULT_PASSWORD non défini$(RESET)"; exit 1; fi
	@echo "$(YELLOW)Application du plan Terraform...$(RESET)"
	cd $(TF_DIR) && terraform apply -auto-approve
	@echo "$(YELLOW)Récupération des secrets S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(ANSIBLE_DIR)/secrets.yml || true
	@echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && chmod 600 $(VAULT_FILE)
	@echo "$(YELLOW)Lancement du playbook Ansible...$(RESET)"
	@cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	  -e "kms_key_id=$$(cd ../$(TF_DIR) && terraform output -raw kms_key_id)" \
	  -e "aws_account_id=$$(cd ../$(TF_DIR) && terraform output -raw aws_account_id)" \
	  -e "ansible_vault_password=$$ANSIBLE_VAULT_PASSWORD" \
	  -e "deploy_user=$$DEPLOY_USER" \
	  --vault-password-file $(VAULT_FILE)
	@echo "$(YELLOW)Upload des secrets vers S3...$(RESET)"
	@aws s3 cp $(ANSIBLE_DIR)/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(GREEN)Déployé par $$DEPLOY_USER — terminé !$(RESET)"

deploy-debug: ## Lance Ansible seul en mode verbeux (-vvv)
	@echo "$(YELLOW)Lancement du playbook Ansible (debug)...$(RESET)"
	@read -p "Ansible Vault password: " VAULT_PASS && \
	  echo "$$VAULT_PASS" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE) && \
	  cd $(ANSIBLE_DIR) && \
	  ANSIBLE_SSH_PIPELINING=0 \
	  ANSIBLE_SCP_IF_SSH=y \
	  ANSIBLE_SSH_ARGS="-o ControlMaster=no -o ControlPersist=0" \
	  ansible-playbook $(PLAYBOOK) -vvv \
	    -e "kms_key_id=$$(cd ../$(TF_DIR) && terraform output -raw kms_key_id)" \
	    -e "aws_account_id=$$(cd ../$(TF_DIR) && terraform output -raw aws_account_id)" \
	    -e "deploy_user=$(DEPLOY_USER)" \
	    --vault-password-file $(VAULT_FILE)

# ─── Ansible seul (infra déjà up) ────────────────────────────────────────────
# Usage : make ansible                          → playbook complet
#         make ansible role=docker_install      → un rôle spécifique
#         make ansible role=security_os,docker_install  → plusieurs rôles
ROLES ?=

ansible: ## Lance Ansible sans Terraform (make ansible [role=<nom>])
	@echo "$(YELLOW)Récupération des secrets S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(ANSIBLE_DIR)/secrets.yml || true
	@read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE) && \
	  cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	    -e "kms_key_id=$$(cd ../$(TF_DIR) && terraform output -raw kms_key_id)" \
	    -e "aws_account_id=$$(cd ../$(TF_DIR) && terraform output -raw aws_account_id)" \
	    -e "deploy_user=$(DEPLOY_USER)" \
	    --vault-password-file $(VAULT_FILE) \
	    $(if $(role),--tags "$(role)")
	@aws s3 cp $(ANSIBLE_DIR)/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(GREEN)Ansible terminé$(if $(role), [rôle(s) : $(role)])!$(RESET)"

# ─── Destruction ─────────────────────────────────────────────────────────────

destroy: ## Détruit l'infra EC2/SG/KeyPair (conserve KMS + S3)
	@echo "$(RED)Destruction de l'infra (KMS + S3 conservés)...$(RESET)"
	cd $(TF_DIR) && terraform destroy \
	  -target=aws_instance.my_alma_server \
	  -target=aws_security_group.ssh_access \
	  -target=aws_key_pair.admin_key \
	  -target=local_file.ansible_inventory \
	  -compact-warnings \
	  -auto-approve

destroy-full: ## Détruit TOUT (infra + vide et supprime le bucket S3)
	@echo "$(RED)Destruction COMPLÈTE — cette action est irréversible !$(RESET)"
	@read -p "Confirmer ? (yes/no): " CONFIRM && [ "$$CONFIRM" = "yes" ] || (echo "Annulé." && exit 1)
	cd $(TF_DIR) && terraform init
	cd $(TF_DIR) && terraform destroy -auto-approve
	@aws s3 rm s3://$(BUCKET) --recursive || true
	@aws s3api delete-objects \
	  --bucket $(BUCKET) \
	  --delete "$$(aws s3api list-object-versions \
	    --bucket $(BUCKET) \
	    --query '{Objects: Versions[].{Key:Key,VersionId:VersionId}}' \
	    --output json)" || true
	@aws s3api delete-objects \
	  --bucket $(BUCKET) \
	  --delete "$$(aws s3api list-object-versions \
	    --bucket $(BUCKET) \
	    --query '{Objects: DeleteMarkers[].{Key:Key,VersionId:VersionId}}' \
	    --output json)" || true
	@aws s3 rb s3://$(BUCKET) || true
	@rm -f $(ANSIBLE_DIR)/secrets.yml
	@echo "$(GREEN)Destruction complète terminée !$(RESET)"

# ─── Nettoyage local ─────────────────────────────────────────────────────────

clean: ## Supprime les fichiers temporaires locaux (vault_pass, inventory...)
	@echo "$(YELLOW)Nettoyage des fichiers temporaires...$(RESET)"
	@rm -f $(VAULT_FILE) $(ANSIBLE_DIR)/secrets.yml
	@echo "$(GREEN)Nettoyage terminé.$(RESET)"
