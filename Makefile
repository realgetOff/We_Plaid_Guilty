.PHONY: help bootstrap init init-upgrade fmt validate plan show output packer deploy deploy-ci deploy-debug ansible kubectl destroy destroy-full clean plan-all

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

# ─── Internal helpers ────────────────────────────────────────────────────────
define tf_output
$(shell cd $(TF_INFRA_DIR) && terraform output -raw $(1) 2>/dev/null)
endef

# ─────────────────────────────────────────────────────────────────────────────

# Display available commands with their descriptions
help:
	@echo ""
	@echo "  $(CYAN)Transcendance — Makefile$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*##"}; {printf "  $(GREEN)%-18s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ─── Bootstrap ───────────────────────────────────────────────────────────────

# Create the S3 bucket used to store secrets and Terraform state (run once)
bootstrap: ## Create S3 secrets bucket
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

# Initialize Terraform backends and providers for both infra and vault workspaces
init: ## Initialize Terraform
	@echo "$(YELLOW)Initializing Terraform...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init
	cd $(TF_VAULT_DIR) && terraform init

# Upgrade Terraform providers to latest allowed versions
init-upgrade: ## Upgrade Terraform providers
	@echo "$(YELLOW)Upgrading providers...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init -upgrade
	cd $(TF_VAULT_DIR) && terraform init -upgrade

# Format all Terraform files recursively
fmt: ## Format Terraform files
	@echo "$(YELLOW)Formatting Terraform...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform fmt -recursive
	cd $(TF_VAULT_DIR) && terraform fmt -recursive

# Validate Terraform configuration syntax
validate: ## Validate Terraform configuration
	@echo "$(YELLOW)Validating Terraform...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform validate
	cd $(TF_VAULT_DIR) && terraform validate

# Show planned changes for both infra and vault workspaces
plan: ## Show Terraform plan
	@echo "$(YELLOW)Planning infra...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform plan
	@echo "$(YELLOW)Planning vault...$(RESET)"
	cd $(TF_VAULT_DIR) && terraform plan

# Display current Terraform state for both workspaces
show: ## Show Terraform state
	@echo "$(YELLOW)Terraform infra state...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform show
	@echo "$(YELLOW)Terraform vault state...$(RESET)"
	cd $(TF_VAULT_DIR) && terraform show

# Display all Terraform outputs (IPs, IDs, etc.)
output: ## Show Terraform outputs
	@echo "$(YELLOW)Terraform infra outputs...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform output
	@echo "$(YELLOW)Terraform vault outputs...$(RESET)"
	cd $(TF_VAULT_DIR) && terraform output

# ─── Packer ──────────────────────────────────────────────────────────────────

# Build the base AMI with Packer (AlmaLinux 9 ARM64 + security_os + docker)
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

# Full interactive deployment: Terraform infra → Ansible (Vault + K3s) → Terraform vault
deploy: ## Deploy full infrastructure (interactive)
	@echo "$(YELLOW)Fetching secrets from S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(ANSIBLE_DIR)/secrets.yml 2>/dev/null || true
	@read -p "Ansible Vault password: " ANSIBLE_VAULT_PASSWORD && \
	  echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && \
	  chmod 600 $(VAULT_FILE)
	@echo "$(YELLOW)--- Step 1/3: Terraform AWS infra ---$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init && terraform apply -auto-approve
	@echo "$(YELLOW)Waiting for SSH (30s)...$(RESET)"
	@sleep 30
	@echo "$(YELLOW)--- Step 2/3: Ansible (Vault + K3s) ---$(RESET)"
	@cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	  -e "kms_key_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw kms_key_id)" \
	  -e "aws_account_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw aws_account_id)" \
	  -e "deploy_user=$(DEPLOY_USER)" \
	  --vault-password-file $(VAULT_FILE)
	@echo "$(YELLOW)Uploading secrets to S3...$(RESET)"
	@aws s3 cp $(ANSIBLE_DIR)/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(YELLOW)--- Step 3/3: Terraform Vault (policies + roles) ---$(RESET)"
	@cd $(TF_VAULT_DIR) && terraform init && \
	TF_VAR_vault_root_token=$$(ansible-vault decrypt \
		--vault-password-file $(VAULT_FILE) \
		--output - ../../$(ANSIBLE_DIR)/secrets.yml \
		| grep vault_root_token \
		| awk '{print $$2}' \
		| tr -d '"') \
	terraform apply -auto-approve
	@echo "$(GREEN)Deployed by $(DEPLOY_USER) — done!$(RESET)"

# CI/CD deployment without prompts — requires ANSIBLE_VAULT_PASSWORD env var
deploy-ci: bootstrap ## Deploy infrastructure (CI/CD, no prompt)
	@echo "$(YELLOW)Deployed by: $(DEPLOY_USER)$(RESET)"
	@if [ -z "$$ANSIBLE_VAULT_PASSWORD" ]; then \
	  echo "$(RED)ANSIBLE_VAULT_PASSWORD is not set$(RESET)"; exit 1; fi
	@echo "$$ANSIBLE_VAULT_PASSWORD" > $(VAULT_FILE) && chmod 600 $(VAULT_FILE)
	@echo "$(YELLOW)Fetching secrets from S3...$(RESET)"
	@aws s3 cp s3://$(BUCKET)/secrets.yml $(ANSIBLE_DIR)/secrets.yml 2>/dev/null || true
	@echo "$(YELLOW)--- Step 1/3: Terraform AWS infra ---$(RESET)"
	cd $(TF_INFRA_DIR) && terraform init && terraform apply -auto-approve
	@echo "$(YELLOW)Waiting for SSH...$(RESET)"
	@MASTER_IP=$$(cd $(TF_INFRA_DIR) && terraform output -raw master_ip) && \
	for i in $$(seq 1 20); do \
		ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
		-i ~/.ssh/github_actions \
		ec2-user@$$MASTER_IP exit 2>/dev/null && break; \
		echo "Attempt $$i/20..."; \
		sleep 10; \
	done
	@echo "$(YELLOW)--- Step 2/3: Ansible (Vault + K3s) ---$(RESET)"
	@cd $(ANSIBLE_DIR) && ansible-playbook $(PLAYBOOK) \
	  -e "kms_key_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw kms_key_id)" \
	  -e "aws_account_id=$$(cd ../$(TF_INFRA_DIR) && terraform output -raw aws_account_id)" \
	  -e "deploy_user=$$DEPLOY_USER" \
	  --vault-password-file $(VAULT_FILE)
	@echo "$(YELLOW)Uploading secrets to S3...$(RESET)"
	@aws s3 cp $(ANSIBLE_DIR)/secrets.yml s3://$(BUCKET)/secrets.yml
	@echo "$(YELLOW)--- Step 3/3: Terraform Vault (policies + roles) ---$(RESET)"
	@cd $(TF_VAULT_DIR) && terraform init && \
	TF_VAR_vault_root_token=$$(ansible-vault decrypt \
		--vault-password-file $(VAULT_FILE) \
		--output - ../../$(ANSIBLE_DIR)/secrets.yml \
		| grep vault_root_token \
		| awk '{print $$2}' \
		| tr -d '"') \
	terraform apply -auto-approve
	@echo "$(GREEN)Deployed by $$DEPLOY_USER — done!$(RESET)"

# Run Ansible playbook only in verbose mode (-vvv) for debugging
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

# Run Ansible playbook without Terraform — optionally filter by role tag
# Usage: make ansible [role=<name>]
ansible: ## Run Ansible only (make ansible [role=<name>])
	@echo "$(YELLOW)Fetching secrets from S3...$(RESET)"
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
	@echo "$(GREEN)Ansible done$(if $(role), [role(s): $(role)])!$(RESET)"

# ─── Kubectl ─────────────────────────────────────────────────────────────────

# Run kubectl commands on the K3s master node via SSH
# Usage: make kubectl cmd="get nodes"
kubectl: ## Run kubectl on master (make kubectl cmd="get nodes")
	@MASTER_IP=$$(cd $(TF_INFRA_DIR) && terraform output -raw master_ip) && \
	  ssh -i ~/.ssh/github_actions ec2-user@$$MASTER_IP "k3s kubectl $(cmd)"

# ─── Destruction ─────────────────────────────────────────────────────────────

# Destroy EC2 instances and security groups — preserves KMS and S3
destroy: ## Destroy EC2 instances (keeps KMS + S3)
	@echo "$(RED)Destroying infra (KMS + S3 preserved)...$(RESET)"
	cd $(TF_INFRA_DIR) && terraform destroy \
	  -target=module.master.aws_instance.name \
	  -target=module.worker1.aws_instance.name \
	  -target=module.worker2.aws_instance.name \
	  -target=aws_security_group.app_sg \
	  -target=aws_security_group.worker_sg \
	  -target=aws_key_pair.admin_key \
	  -target=local_file.ansible_inventory \
	  -compact-warnings \
	  -auto-approve

# Destroy everything including S3 bucket and all its contents
destroy-full: ## Destroy everything (infra + S3 bucket)
	@echo "$(RED)FULL DESTRUCTION — this action is irreversible!$(RESET)"
	@read -p "Confirm? (yes/no): " CONFIRM && [ "$$CONFIRM" = "yes" ] || (echo "Cancelled." && exit 1)
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
	@echo "$(GREEN)Full destruction complete!$(RESET)"

# ─── Cleanup ─────────────────────────────────────────────────────────────────

# Remove local temporary files (vault password + secrets)
clean: ## Remove local temporary files
	@echo "$(YELLOW)Cleaning up temporary files...$(RESET)"
	@rm -f $(VAULT_FILE) $(ANSIBLE_DIR)/secrets.yml
	@echo "$(GREEN)Cleanup done.$(RESET)"

plan-all:
	@echo "--- Planning Infra ---"
	cd $(TF_INFRA_DIR) && terraform plan -out=infra.tfplan
	@echo "--- Planning Vault (using remote state) ---"
	cd $(TF_VAULT_DIR) && terraform plan
