# ============================================================================
# GENERAL CONFIGURATION
# ============================================================================

variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-west-2"
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "gospital"
}

# ============================================================================
# NETWORK CONFIGURATION
# ============================================================================

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones for multi-AZ deployment"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b"]
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to connect to RDS for migrations (your IP)"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# ============================================================================
# DATABASE CONFIGURATION (RDS MySQL)
# ============================================================================

variable "db_username" {
  description = "RDS master username"
  type        = string
  default     = "admin"
  sensitive   = true
}

variable "db_password" {
  description = "RDS master password"
  type        = string
  default     = "GospitalSecure123"
  sensitive   = true
}

variable "db_name" {
  description = "RDS database name"
  type        = string
  default     = "gospital"
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

# ============================================================================
# DYNAMODB CONFIGURATION
# ============================================================================

variable "medical_records_table" {
  description = "DynamoDB table name for medical records"
  type        = string
  default     = "medical_records"
}

variable "invoices_table" {
  description = "DynamoDB table name for invoices"
  type        = string
  default     = "invoices"
}

# ============================================================================
# ECS CONTAINER CONFIGURATION
# ============================================================================

variable "container_cpu" {
  description = "Fargate task CPU units (256 = 0.25 vCPU)"
  type        = number
  default     = 256
}

variable "container_memory" {
  description = "Fargate task memory in MB"
  type        = number
  default     = 512
}

variable "container_port" {
  description = "Container port for the application"
  type        = number
  default     = 8080
}

# ============================================================================
# ECS SERVICE CONFIGURATION (Internal Service - Staff)
# ============================================================================

variable "internal_desired_count" {
  description = "Desired number of tasks for internal service"
  type        = number
  default     = 1
}

variable "internal_min_tasks" {
  description = "Minimum number of tasks for internal service auto-scaling"
  type        = number
  default     = 1
}

variable "internal_max_tasks" {
  description = "Maximum number of tasks for internal service auto-scaling"
  type        = number
  default     = 10
}

# ============================================================================
# ECS SERVICE CONFIGURATION (External Service - Public)
# ============================================================================

variable "external_desired_count" {
  description = "Desired number of tasks for external service"
  type        = number
  default     = 1
}

variable "external_min_tasks" {
  description = "Minimum number of tasks for external service auto-scaling"
  type        = number
  default     = 1
}

variable "external_max_tasks" {
  description = "Maximum number of tasks for external service auto-scaling"
  type        = number
  default     = 10
}

# ============================================================================
# AUTO-SCALING CONFIGURATION
# ============================================================================

variable "autoscaling_cpu_target" {
  description = "Target CPU utilization percentage for auto-scaling"
  type        = number
  default     = 50.0
}

variable "autoscaling_memory_target" {
  description = "Target memory utilization percentage for auto-scaling"
  type        = number
  default     = 80.0
}

variable "autoscaling_scale_in_cooldown" {
  description = "Cooldown period (seconds) before scaling in (removing tasks)"
  type        = number
  default     = 180
}
