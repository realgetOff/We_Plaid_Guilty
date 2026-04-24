.PHONY: help bootstrap init init-upgrade fmt validate plan show output packer deploy deploy-ci deploy-debug ansible kubectl destroy destroy-full clean plan-all tf-vault

# ─── Colors ──────────────────────────────────────────────────────────────────
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
DOMAIN        := play-stupid.games
SECRETS_FILE  := $(ANSIBLE_DIR)/secrets.yml

# ─── Internal helpers ────────────────────────────────────────────────────────
define tf_output
$(shell cd $(TF_INFRA_DIR) && terraform output -raw $(1) 2>/dev/null)
endef

# Extract vault root token from secrets.yml (run from repo root)
define vault_token
$$(ansible-vault decrypt \
	--vault-password-file $(VAULT_FILE) \
	--output - $(SECRETS_FILE) \
	| grep vault_root_token \
	| awk '{print $$2}' \
	| tr -d '"')
endef

# Re-import Route53 zone if it exists in AWS but not in state
define import_route53
	@ZONE_ID=$$(aws route53 list-hosted-zones \
	  --query "HostedZones[?Name=='$(DOMAIN).'].Id" \
	  --output text 2>/dev/null | cut -d'/' -f3) && \
	if [ -n "$$ZONE_ID" ]; then \
	  echo "$(YELLOW)Re-importing Route53 zone $$ZONE_ID...$(RESET)" && \
	  cd $(TF_INFRA_DIR) && terraform import aws_route53_zone.main $$ZONE_ID 2>/dev/null || true; \
	fi
endef

# Uninstall K3s on all nodes to clean ENIs before destroy
define uninstall_k3s
	@echo "$(YELLOW)Cleaning K3s network interfaces...$(RESET)"
	@MASTER_IP=$$(cd $(TF_INFRA_DIR) && terraform output -raw master_ip 2>/dev/null) && \
	WORKER1_IP=$$(cd $(TF_INFRA_DIR) && terraform output -raw worker1_ip 2>/dev/null) && \
	WORKER2_IP=$$(cd $(TF_INFRA_DIR) && terraform output -raw worker2_ip 2>/dev/null) && \
	ssh -o StrictHostKeyChecking=no -i ~/.ssh/github_actions ec2-user@$$MASTER_IP "sudo k3s-uninstall.sh" 2>/dev/null || true && \
	ssh -o StrictHostKeyChecking=no -i ~/.ssh/github_actions ec2-user@$$WORKER1_IP "sudo k3s-agent-uninstall.sh" 2>/dev/null || true && \
	ssh -o StrictHostKeyChecking=no -i ~/.ssh/github_actions ec2-user@$$WORKER2_IP "sudo k3s-agent-uninstall.sh" 2>/dev/null || true
endef

# ─────────────────────────────────────────────────────────────────────────────

help: ## Display available commands
	@echo ""
	@echo "  $(CYAN)Transcendance — Makefile$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*##"}; {printf "  $(GREEN)%-18s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ─── Bootstrap ───────────────────────────────────────────────────────────────

bootstrap: ## Create S3 secrets bucket (run once)
	@echo "$(YELLOW)Creating S3 bucket...$(RESET)"
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
	@echo "$(GREEN)Bootstrap done — bucket $(BUCKET) ready!$(RESET)"

# ─── Terraform ───────────────────────────────────────────────────────────────

init: ## Initialize Terraform
	@echo "$(YELLOW)Initializing Terraform...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init
	cd $(TF_VAULT_DIR) && terraform init

init-upgrade: ## Upgrade Terraform providers
	@echo "$(YELLOW)Upgrading providers...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init -upgrade
	cd $(TF_VAULT_DIR) && terraform init -upgrade

fmt: ## Format Terraform files
	@echo "$(YELLOW)Formatting Terraform...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform fmt -recursive
	cd $(TF_VAULT_DIR) && terraform fmt -recursive

validate: ## Validate Terraform configuration
	@echo "$(YELLOW)Validating Terraform...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform validate
	cd $(TF_VAULT_DIR) && terraform validate

plan: ## Show Terraform plan
	@echo "$(YELLOW)Planning infra...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform plan
	@echo "$(YELLOW)Planning vault...$(RESET)"
	cd $(TF_VAULT_DIR) && terraform plan

show: ## Show Terraform state
	@echo "$(YELLOW)Terraform infra state...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform show
	@echo "$(YELLOW)Terraform vault state...$(RESET)"
	cd $(TF_VAULT_DIR) && terraform show

output: ## Show Terraform outputs
	@echo "$(YELLOW)Terraform infra outputs...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform output
	@echo "$(YELLOW)Terraform vault outputs...$(RESET)"
	cd $(TF_VAULT_DIR) && terraform output

tf-vault: ## Apply Terraform vault only (infra already up)
	@echo "$(YELLOW)Applying Terraform Vault...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(SECRETS_FILE) 2>/dev/null || true
	@if [ ! -f $(VAULT_FILE) ]; then \
	  read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE); \
	fi
	@cd $(TF_VAULT_DIR) && terraform init && \
	TF_VAR_vault_root_token=$(call vault_token) \
	terraform apply -auto-approve
	@echo "$(GREEN)Terraform Vault done!$(RESET)"

# ─── Packer ──────────────────────────────────────────────────────────────────

packer: ## Build base AMI with Packer
	@echo "$(YELLOW)Building Packer AMI...$(RESET)"
	@if [ ! -f $(VAULT_FILE) ]; then \
	  read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE); \
	fi
	ANSIBLE_VAULT_PASSWORD_FILE=$(VAULT_FILE) \
	  packer build $(PACKER_DIR)/alma.pkr.hcl
	@echo "$(GREEN)AMI built successfully!$(RESET)"

# ─── Deployment ──────────────────────────────────────────────────────────────

deploy: ## Deploy full infrastructure (interactive)
	@echo "$(YELLOW)Fetching secrets from S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(SECRETS_FILE) 2>/dev/null || true
	@read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE)
	@echo "$(YELLOW)--- Step 1/3: Terraform AWS infra ---$(RESET)"
	@cd $(TF_INFRA_DIR) && terraform init
	$(call import_route53)
	@cd $(TF_INFRA_DIR) && terraform apply -auto-approve
	@echo "$(YELLOW)Waiting for SSH (30s)...$(RESET)"
	@sleep 30
	@echo "$(YELLOW)--- Step 2/3: Ansible (K3s + Vault init) ---$(RESET)"
	@cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	  -e "kms_key_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw kms_key_id)" \
	  -e "aws_account_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw aws_account_id)" \
	  -e "deploy_user=$(DEPLOY_USER)" \
	  --vault-password-file $(VAULT_FILE)
	@echo "$(YELLOW)Uploading secrets to S3...$(RESET)"
	@aws s3 cp $(SECRETS_FILE) s3://$(BUCKET)/secrets.yml
	@echo "$(YELLOW)--- Step 3/3: Terraform Vault (policies + roles) ---$(RESET)"
	@cd $(TF_VAULT_DIR) && terraform init && \
	TF_VAR_vault_root_token=$(call vault_token) \
	terraform apply -auto-approve
	@echo "$(GREEN)Deployed by $(DEPLOY_USER) — done!$(RESET)"

deploy-ci: bootstrap ## Deploy infrastructure (CI/CD, no prompt)
	@echo "$(YELLOW)Deployed by: $(DEPLOY_USER)$(RESET)"
	@if [ -z "$$ANSIBLE_VAULT_PASSWORD" ]; then \
	  echo "$(RED)ANSIBLE_VAULT_PASSWORD is not set$(RESET)"; exit 1; fi
	@echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && chmod 600 $(VAULT_FILE)
	@echo "$(YELLOW)Fetching secrets from S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(SECRETS_FILE) 2>/dev/null || true
	@echo "$(YELLOW)--- Step 1/3: Terraform AWS infra ---$(RESET)"
	@cd $(TF_INFRA_DIR) && terraform init
	$(call import_route53)
	@cd $(TF_INFRA_DIR) && terraform apply -auto-approve
	@echo "$(YELLOW)Waiting for SSH...$(RESET)"
	@MASTER_IP=$$(cd $(TF_INFRA_DIR) && terraform output -raw master_ip) && \
	for i in $$(seq 1 20); do \
		ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
		-i ~/.ssh/github_actions \
		ec2-user@$$MASTER_IP exit 2>/dev/null && break; \
		echo "Attempt $$i/20..."; \
		sleep 10; \
	done
	@echo "$(YELLOW)--- Step 2/3: Ansible (K3s + Vault init) ---$(RESET)"
	@cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	  -e "kms_key_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw kms_key_id)" \
	  -e "aws_account_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw aws_account_id)" \
	  -e "deploy_user=$$DEPLOY_USER" \
	  --vault-password-file $(VAULT_FILE)
	@echo "$(YELLOW)Uploading secrets to S3...$(RESET)"
	@aws s3 cp $(SECRETS_FILE) s3://$(BUCKET)/secrets.yml
	@echo "$(YELLOW)--- Step 3/3: Terraform Vault (policies + roles) ---$(RESET)"
	@aws s3 rm s3://$(BUCKET)/terraform-vault.tfstate || true
	@cd $(TF_VAULT_DIR) && terraform init && \
	TF_VAR_vault_root_token=$(call vault_token) \
	terraform apply -auto-approve
	@echo "$(GREEN)Deployed by $$DEPLOY_USER — done!$(RESET)"

deploy-debug: ## Run Ansible in verbose mode (debug)
	@echo "$(YELLOW)Running Ansible playbook (debug)...$(RESET)"
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

# ─── Ansible only (infra already up) ─────────────────────────────────────────

ansible: ## Run Ansible only (make ansible [role=<n>])
	@echo "$(YELLOW)Fetching secrets from S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(SECRETS_FILE) 2>/dev/null || true
	@read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE) && \
	  cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	    -e "kms_key_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw kms_key_id)" \
	    -e "aws_account_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw aws_account_id)" \
	    -e "deploy_user=$(DEPLOY_USER)" \
	    --vault-password-file $(VAULT_FILE) \
	    $(if $(role),--tags "$(role)")
	@aws s3 cp $(SECRETS_FILE) s3://$(BUCKET)/secrets.yml
	@echo "$(GREEN)Ansible done$(if $(role), [role(s): $(role)])!$(RESET)"

# ─── Kubectl ─────────────────────────────────────────────────────────────────

kubectl: ## Run kubectl on master (make kubectl cmd="get nodes")
	@MASTER_IP=$$(cd $(TF_INFRA_DIR) && terraform output -raw master_ip) && \
	  ssh -i ~/.ssh/github_actions ec2-user@$$MASTER_IP "k3s kubectl $(cmd)"

# ─── Destruction ─────────────────────────────────────────────────────────────

destroy: ## Destroy EC2 instances (keeps KMS, S3, Route53)
	@echo "$(RED)Destroying EC2 instances (KMS + S3 + Route53 preserved)...$(RESET)"
	$(call uninstall_k3s)
	@cd $(TF_INFRA_DIR) && terraform state rm aws_route53_zone.main || true
	cd $(TF_INFRA_DIR) && terraform destroy -auto-approve
	@echo "$(GREEN)EC2 instances destroyed!$(RESET)"

destroy-full: ## Destroy everything (Route53 zone preserved)
	@echo "$(RED)FULL DESTRUCTION — this action is irreversible!$(RESET)"
	@read -p "Confirm? (yes/no): " CONFIRM && [ "$$CONFIRM" = "yes" ] || (echo "Cancelled." && exit 1)
	$(call uninstall_k3s)
	@echo "$(YELLOW)Destroying Terraform Vault...$(RESET)"
	@cd $(TF_VAULT_DIR) && terraform init && terraform destroy -auto-approve || true
	@echo "$(YELLOW)Removing Route53 zone from state...$(RESET)"
	@cd $(TF_INFRA_DIR) && terraform state rm aws_route53_zone.main || true
	@echo "$(YELLOW)Destroying Terraform infra...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init && terraform destroy -auto-approve
	@echo "$(YELLOW)Cleaning KMS alias...$(RESET)"
	@aws kms delete-alias --alias-name alias/vault-unseal --region $(REGION) || true
	@echo "$(YELLOW)Cleaning S3 bucket...$(RESET)"
	@aws s3 rm s3://$(BUCKET) --recursive || true
	@aws s3api delete-objects \
	  --bucket $(BUCKET) \
	  --delete "$$(aws s3api list-object-versions \
	    --bucket $(BUCKET) \
	    --query '{Objects: Versions[].{Key:Key,VersionId:VersionId}}' \
	    --output json)" 2>/dev/null || true
	@aws s3api delete-objects \
	  --bucket $(BUCKET) \
	  --delete "$$(aws s3api list-object-versions \
	    --bucket $(BUCKET) \
	    --query '{Objects: DeleteMarkers[].{Key:Key,VersionId:VersionId}}' \
	    --output json)" 2>/dev/null || true
	@aws s3 rb s3://$(BUCKET) || true
	@rm -f $(SECRETS_FILE)
	@echo "$(GREEN)Full destruction complete! Route53 zone preserved.$(RESET)"

# ─── Cleanup ─────────────────────────────────────────────────────────────────

clean: ## Remove local temporary files
	@echo "$(YELLOW)Cleaning up temporary files...$(RESET)"
	@rm -f $(VAULT_FILE) $(SECRETS_FILE)
	@echo "$(GREEN)Cleanup done.$(RESET)"

plan-all: ## Plan all Terraform workspaces
	@echo "--- Planning Infra ---"
	cd $(TF_INFRA_DIR) && terraform plan
	@echo "--- Planning Vault (using remote state) ---"
	cd $(TF_VAULT_DIR) && terraform plan
