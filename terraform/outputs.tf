# ============================================================================
# ALB OUTPUTS
# ============================================================================

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.main.dns_name
}

output "alb_url" {
  description = "URL to access the API"
  value       = "http://${aws_lb.main.dns_name}"
}

output "grafana_url" {
  description = "URL to access Grafana dashboard"
  value       = "http://${aws_lb.main.dns_name}/grafana"
}

# ============================================================================
# ECR OUTPUTS
# ============================================================================

output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = aws_ecr_repository.gospital.repository_url
}

# ============================================================================
# RDS OUTPUTS
# ============================================================================

output "rds_endpoint" {
  description = "RDS MySQL endpoint"
  value       = aws_db_instance.mysql.endpoint
}

output "rds_database_name" {
  description = "RDS database name"
  value       = aws_db_instance.mysql.db_name
}

# ============================================================================
# DYNAMODB OUTPUTS
# ============================================================================

output "dynamodb_medical_records_table" {
  description = "DynamoDB Medical Records table name"
  value       = aws_dynamodb_table.medical_records.name
}

output "dynamodb_invoices_table" {
  description = "DynamoDB Invoices table name"
  value       = aws_dynamodb_table.invoices.name
}

# ============================================================================
# ECS OUTPUTS
# ============================================================================

output "ecs_cluster_name" {
  description = "ECS Cluster name"
  value       = aws_ecs_cluster.main.name
}

output "ecs_internal_service_name" {
  description = "ECS Internal Service name"
  value       = aws_ecs_service.internal.name
}

output "ecs_external_service_name" {
  description = "ECS External Service name"
  value       = aws_ecs_service.external.name
}

# ============================================================================
# TARGET GROUP OUTPUTS
# ============================================================================

output "internal_target_group_arn" {
  description = "Internal service target group ARN"
  value       = aws_lb_target_group.internal.arn
}

output "external_target_group_arn" {
  description = "External service target group ARN"
  value       = aws_lb_target_group.external.arn
}

# ============================================================================
# VPC OUTPUTS
# ============================================================================

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "public_subnets" {
  description = "Public subnet IDs"
  value       = aws_subnet.public[*].id
}

output "private_ecs_subnets" {
  description = "Private ECS subnet IDs"
  value       = aws_subnet.private_ecs[*].id
}

output "private_rds_subnets" {
  description = "Private RDS subnet IDs"
  value       = aws_subnet.private_rds[*].id
}
