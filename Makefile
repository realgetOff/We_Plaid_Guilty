.PHONY: help bootstrap init init-upgrade fmt validate plan show output packer deploy deploy-ci deploy-debug ansible destroy destroy-full clean

# ─── Couleurs ────────────────────────────────────────────────────────────────
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
CYAN   := \033[0;36m
RESET  := \033[0m

# ─── Variables ───────────────────────────────────────────────────────────────
BUCKET        := transcendance-secrets-43783683331
REGION        := eu-north-1
TF_INFRA_DIR  := terraform/infra
TF_VAULT_DIR  := terraform/vault
ANSIBLE_DIR   := ansible
PACKER_DIR    := packer
PLAYBOOK      := deploy.yml
VAULT_FILE    := ~/.vault_pass
DEPLOY_USER   ?= manual

# ─── Helpers internes ────────────────────────────────────────────────────────
define tf_output
$(shell cd $(TF_INFRA_DIR) && terraform output -raw $(1) 2>/dev/null)
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
	cd $(TF_INFRA_DIR) && terraform init
	cd $(TF_VAULT_DIR) && terraform init

init-upgrade: ## Met à jour les providers Terraform
	@echo "$(YELLOW)Mise à jour des providers...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init -upgrade
	cd $(TF_VAULT_DIR) && terraform init -upgrade

fmt: ## Formate les fichiers Terraform
	@echo "$(YELLOW)Formatage Terraform...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform fmt -recursive
	cd $(TF_VAULT_DIR) && terraform fmt -recursive

validate: ## Valide la configuration Terraform
	@echo "$(YELLOW)Validation Terraform...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform validate
	cd $(TF_VAULT_DIR) && terraform validate

plan: ## Affiche le plan Terraform
	@echo "$(YELLOW)Plan infra...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform plan
	@echo "$(YELLOW)Plan vault...$(RESET)"
	cd $(TF_VAULT_DIR) && terraform plan

show: ## Affiche l'état Terraform
	@echo "$(YELLOW)État Terraform infra...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform show
	@echo "$(YELLOW)État Terraform vault...$(RESET)"
	cd $(TF_VAULT_DIR) && terraform show

output: ## Affiche les outputs Terraform
	@echo "$(YELLOW)Outputs Terraform infra...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform output
	@echo "$(YELLOW)Outputs Terraform vault...$(RESET)"
	cd $(TF_VAULT_DIR) && terraform output

# ─── Packer ──────────────────────────────────────────────────────────────────

packer: ## Build l'AMI de base avec Packer
	@echo "$(YELLOW)Build de l'AMI Packer...$(RESET)"
	@if [ ! -f $(VAULT_FILE) ]; then \
	  read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE); \
	fi
	ANSIBLE_VAULT_PASSWORD_FILE=$(VAULT_FILE) \
	  packer build $(PACKER_DIR)/alma.pkr.hcl
	@echo "$(GREEN)AMI buildée avec succès !$(RESET)"

# ─── Déploiement ─────────────────────────────────────────────────────────────

deploy: ## Déploie l'infra complète (interactif)
	@echo "$(YELLOW)Récupération des secrets S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(ANSIBLE_DIR)/secrets.yml 2>/dev/null || true
	@read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE)
	@echo "$(YELLOW)--- Étape 1/3 : Terraform infra AWS ---$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init && terraform apply -auto-approve
	@echo "$(YELLOW)Attente SSH (30s)...$(RESET)"
	@sleep 30
	@echo "$(YELLOW)--- Étape 2/3 : Ansible (services + Vault) ---$(RESET)"
	@cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	  -e "kms_key_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw kms_key_id)" \
	  -e "aws_account_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw aws_account_id)" \
	  -e "deploy_user=$(DEPLOY_USER)" \
	  --vault-password-file $(VAULT_FILE)
	@echo "$(YELLOW)Upload des secrets vers S3...$(RESET)"
	@aws s3 cp $(ANSIBLE_DIR)/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(YELLOW)--- Étape 3/3 : Terraform Vault ---$(RESET)"
	@VAULT_ROOT_TOKEN=$$(ansible-vault decrypt \
	  --vault-password-file $(VAULT_FILE) \
	  --output - $(ANSIBLE_DIR)/secrets.yml \
	  | grep vault_root_token \
	  | awk '{print $$2}' \
	  | tr -d '"') && \
	  cd $(TF_VAULT_DIR) && terraform init && \
	  TF_VAR_vault_root_token=$$VAULT_ROOT_TOKEN terraform apply -auto-approve
	@echo "$(GREEN)Déployé par $(DEPLOY_USER) — terminé !$(RESET)"

deploy-ci: bootstrap ## Déploie l'infra (CI/CD, sans prompt)
	@echo "$(YELLOW)Déployé par : $(DEPLOY_USER)$(RESET)"
	@if [ -z "$$ANSIBLE_VAULT_PASSWORD" ]; then \
	  echo "$(RED)ANSIBLE_VAULT_PASSWORD non défini$(RESET)"; exit 1; fi
	@echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && chmod 600 $(VAULT_FILE)
	@echo "$(YELLOW)Récupération des secrets S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(ANSIBLE_DIR)/secrets.yml 2>/dev/null || true
	@echo "$(YELLOW)--- Étape 1/3 : Terraform infra AWS ---$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init && terraform apply -auto-approve
	@echo "$(YELLOW)Attente SSH...$(RESET)"
	@for i in $$(seq 1 20); do \
	  ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
	    -i ~/.ssh/github_actions \
	    ec2-user@$(call tf_output,app_ip) exit 2>/dev/null && break; \
	  echo "Tentative $$i/20..."; \
	  sleep 10; \
	done
	@echo "$(YELLOW)--- Étape 2/3 : Ansible (services + Vault) ---$(RESET)"
	@cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	  -e "kms_key_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw kms_key_id)" \
	  -e "aws_account_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw aws_account_id)" \
	  -e "deploy_user=$$DEPLOY_USER" \
	  --vault-password-file $(VAULT_FILE)
	@echo "$(YELLOW)Upload des secrets vers S3...$(RESET)"
	@aws s3 cp $(ANSIBLE_DIR)/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(YELLOW)--- Étape 3/3 : Terraform Vault ---$(RESET)"
	@VAULT_ROOT_TOKEN=$$(ansible-vault decrypt \
	  --vault-password-file $(VAULT_FILE) \
	  --output - $(ANSIBLE_DIR)/secrets.yml \
	  | grep vault_root_token \
	  | awk '{print $$2}' \
	  | tr -d '"') && \
	  cd $(TF_VAULT_DIR) && terraform init && \
	  TF_VAR_vault_root_token=$$VAULT_ROOT_TOKEN terraform apply -auto-approve
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
	    -e "kms_key_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw kms_key_id)" \
	    -e "aws_account_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw aws_account_id)" \
	    -e "deploy_user=$(DEPLOY_USER)" \
	    --vault-password-file $(VAULT_FILE)

# ─── Ansible seul (infra déjà up) ────────────────────────────────────────────

ansible: ## Lance Ansible sans Terraform (make ansible [role=<nom>])
	@echo "$(YELLOW)Récupération des secrets S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(ANSIBLE_DIR)/secrets.yml 2>/dev/null || true
	@read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE) && \
	  cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	    -e "kms_key_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw kms_key_id)" \
	    -e "aws_account_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw aws_account_id)" \
	    -e "deploy_user=$(DEPLOY_USER)" \
	    --vault-password-file $(VAULT_FILE) \
	    $(if $(role),--tags "$(role)")
	@aws s3 cp $(ANSIBLE_DIR)/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(GREEN)Ansible terminé$(if $(role), [rôle(s) : $(role)])!$(RESET)"

# ─── Destruction ─────────────────────────────────────────────────────────────

destroy: ## Détruit l'infra EC2/SG/KeyPair (conserve KMS + S3)
	@echo "$(RED)Destruction de l'infra (KMS + S3 conservés)...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform destroy \
	  -target=module.app.aws_instance.name \
	  -target=module.elk.aws_instance.name \
	  -target=module.grafana.aws_instance.name \
	  -target=aws_security_group.app_sg \
	  -target=aws_security_group.monitoring_sg \
	  -target=aws_key_pair.admin_key \
	  -target=local_file.ansible_inventory \
	  -compact-warnings \
	  -auto-approve

destroy-full: ## Détruit TOUT (infra + vide et supprime le bucket S3)
	@echo "$(RED)Destruction COMPLÈTE — cette action est irréversible !$(RESET)"
	@read -p "Confirmer ? (yes/no): " CONFIRM && [ "$$CONFIRM" = "yes" ] || (echo "Annulé." && exit 1)
	cd $(TF_VAULT_DIR) && terraform init && terraform destroy -auto-approve || true
	cd $(TF_INFRA_DIR) && terraform init && terraform destroy -auto-approve
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

clean: ## Supprime les fichiers temporaires locaux
	@echo "$(YELLOW)Nettoyage des fichiers temporaires...$(RESET)"
	@rm -f $(VAULT_FILE) $(ANSIBLE_DIR)/secrets.yml
	@echo "$(GREEN)Nettoyage terminé.$(RESET)"
