.PHONY: help deploy destroy start stop log shell clean status update

# Variables
TERRAFORM_DIR := terraform
AWS_REGION := us-west-2
PROJECT_NAME := gospital

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Gospital - Hospital Management API"
	@echo ""
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  make %-15s %s\n", $$1, $$2}'

deploy: ## Deploy infrastructure to AWS
	@echo "Deploying Gospital to AWS..."
	@cd $(TERRAFORM_DIR) && terraform init
	@cd $(TERRAFORM_DIR) && terraform apply -auto-approve
	@echo ""
	@echo "Deployment complete!"
	@echo "API URL:     $$(cd $(TERRAFORM_DIR) && terraform output -raw alb_url)"
	@echo "Grafana URL: $$(cd $(TERRAFORM_DIR) && terraform output -raw grafana_url)"

update: ## Rebuild Docker image and update both ECS services using Terraform
	@echo "Updating application code..."
	@echo "Forcing Terraform to rebuild Docker image..."
	@cd $(TERRAFORM_DIR) && terraform taint docker_image.gospital
	@cd $(TERRAFORM_DIR) && terraform taint docker_registry_image.gospital
	@echo "Rebuilding and pushing image via Terraform Docker provider..."
	@cd $(TERRAFORM_DIR) && terraform apply \
		-target=docker_image.gospital \
		-target=docker_registry_image.gospital \
		-auto-approve
	@echo "Forcing ECS services to redeploy with new image..."
	@CLUSTER=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_cluster_name) && \
	INTERNAL_SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_internal_service_name) && \
	EXTERNAL_SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_external_service_name) && \
	aws ecs update-service \
		--cluster $$CLUSTER \
		--service $$INTERNAL_SERVICE \
		--force-new-deployment \
		--region $(AWS_REGION) > /dev/null && \
	aws ecs update-service \
		--cluster $$CLUSTER \
		--service $$EXTERNAL_SERVICE \
		--force-new-deployment \
		--region $(AWS_REGION) > /dev/null && \
	echo "Update complete! New image deployed to both services."

destroy: ## Destroy all infrastructure
	@echo "WARNING: This will destroy all infrastructure!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		cd $(TERRAFORM_DIR) && terraform destroy -auto-approve; \
		echo "Removing local Docker images..."; \
		docker rmi $(PROJECT_NAME)-api:latest -f 2>/dev/null || true; \
		docker images | grep $(PROJECT_NAME) | awk '{print $$3}' | xargs -r docker rmi -f 2>/dev/null || true; \
		echo "Infrastructure destroyed."; \
	else \
		echo "Destroy cancelled."; \
	fi

start: ## Start both ECS services
	@echo "Starting both ECS services..."
	@CLUSTER=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_cluster_name) && \
	INTERNAL_SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_internal_service_name) && \
	EXTERNAL_SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_external_service_name) && \
	aws ecs update-service \
		--cluster $$CLUSTER \
		--service $$INTERNAL_SERVICE \
		--desired-count 1 \
		--region $(AWS_REGION) > /dev/null && \
	aws ecs update-service \
		--cluster $$CLUSTER \
		--service $$EXTERNAL_SERVICE \
		--desired-count 1 \
		--region $(AWS_REGION) > /dev/null && \
	echo "Both services started."

stop: ## Stop both ECS services
	@echo "Stopping both ECS services..."
	@CLUSTER=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_cluster_name) && \
	INTERNAL_SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_internal_service_name) && \
	EXTERNAL_SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_external_service_name) && \
	aws ecs update-service \
		--cluster $$CLUSTER \
		--service $$INTERNAL_SERVICE \
		--desired-count 0 \
		--region $(AWS_REGION) > /dev/null && \
	aws ecs update-service \
		--cluster $$CLUSTER \
		--service $$EXTERNAL_SERVICE \
		--desired-count 0 \
		--region $(AWS_REGION) > /dev/null && \
	echo "Both services stopped."

log: ## View ECS service logs
	@echo "Select service:"
	@echo "  1) Internal (staff)"
	@echo "  2) External (public)"
	@read -p "Choice [1-2]: " choice; \
	CLUSTER=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_cluster_name) && \
	if [ "$$choice" = "1" ]; then \
		SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_internal_service_name); \
		echo "Streaming internal service logs..."; \
	else \
		SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_external_service_name); \
		echo "Streaming external service logs..."; \
	fi && \
	TASK=$$(aws ecs list-tasks --cluster $$CLUSTER --service-name $$SERVICE --region $(AWS_REGION) --query 'taskArns[0]' --output text) && \
	if [ "$$TASK" = "None" ] || [ -z "$$TASK" ]; then \
		echo "No running tasks found."; \
	else \
		echo "Streaming logs from task: $$TASK"; \
		aws logs tail /ecs/$(PROJECT_NAME) --follow --region $(AWS_REGION); \
	fi

shell: ## Get shell access to running ECS container
	@echo "Select service:"
	@echo "  1) Internal (staff)"
	@echo "  2) External (public)"
	@read -p "Choice [1-2]: " choice; \
	CLUSTER=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_cluster_name) && \
	if [ "$$choice" = "1" ]; then \
		SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_internal_service_name); \
		echo "Connecting to internal service..."; \
	else \
		SERVICE=$$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_external_service_name); \
		echo "Connecting to external service..."; \
	fi && \
	TASK=$$(aws ecs list-tasks --cluster $$CLUSTER --service-name $$SERVICE --region $(AWS_REGION) --query 'taskArns[0]' --output text) && \
	if [ "$$TASK" = "None" ] || [ -z "$$TASK" ]; then \
		echo "No running tasks found. Start the service first with 'make start'"; \
	else \
		echo "Opening shell in task: $$TASK"; \
		aws ecs execute-command \
			--cluster $$CLUSTER \
			--task $$TASK \
			--container api \
			--interactive \
			--command "/bin/sh" \
			--region $(AWS_REGION); \
	fi

clean: ## Clean up local Terraform files and Docker artifacts
	@echo "Cleaning up Terraform files..."
	@rm -rf $(TERRAFORM_DIR)/.terraform
	@rm -rf $(TERRAFORM_DIR)/.terraform.lock.hcl
	@rm -rf $(TERRAFORM_DIR)/terraform.tfstate*
	@rm -rf docs/
	@echo "Cleaning up Docker artifacts..."
	@docker system prune -f
	@echo "Cleanup complete."

status: ## Show deployment status
	@echo "Gospital Deployment Status"
	@echo "=========================="
	@if [ -f "$(TERRAFORM_DIR)/terraform.tfstate" ]; then \
		echo "API URL:          $$(cd $(TERRAFORM_DIR) && terraform output -raw alb_url 2>/dev/null || echo 'Not deployed')"; \
		echo "Grafana URL:      $$(cd $(TERRAFORM_DIR) && terraform output -raw grafana_url 2>/dev/null || echo 'Not deployed')"; \
		echo "RDS Endpoint:     $$(cd $(TERRAFORM_DIR) && terraform output -raw rds_endpoint 2>/dev/null || echo 'Not deployed')"; \
		echo "ECS Cluster:      $$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_cluster_name 2>/dev/null || echo 'Not deployed')"; \
		echo "Internal Service: $$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_internal_service_name 2>/dev/null || echo 'Not deployed')"; \
		echo "External Service: $$(cd $(TERRAFORM_DIR) && terraform output -raw ecs_external_service_name 2>/dev/null || echo 'Not deployed')"; \
	else \
		echo "Infrastructure not deployed yet."; \
	fi