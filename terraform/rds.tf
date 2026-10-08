# DB Subnet Group (Multi-AZ) - Using public subnets for migration access
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = aws_subnet.public[*].id

  tags = {
    Name = "${var.project_name}-db-subnet-group"
  }
}

# RDS MySQL Instance (Multi-AZ)
resource "aws_db_instance" "mysql" {
  identifier     = "${var.project_name}-mysql"
  engine         = "mysql"
  engine_version = "8.0"
  instance_class = var.db_instance_class

  allocated_storage     = 20
  max_allocated_storage = 100
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password

  multi_az               = false # Single AZ for free tier
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = true # Allow public access for migrations

  backup_retention_period = 1 # Reduced for free tier (was 7)
  backup_window           = "03:00-04:00"
  maintenance_window      = "mon:04:00-mon:05:00"

  skip_final_snapshot       = true
  final_snapshot_identifier = "${var.project_name}-mysql-final-snapshot"

  enabled_cloudwatch_logs_exports = ["error", "general", "slowquery"]

  tags = {
    Name = "${var.project_name}-mysql"
  }
}

# Null resource to initialize MySQL database with schema and seed data
resource "null_resource" "mysql_init" {
  triggers = {
    # Re-run when init.sql changes
    init_sql_hash = filemd5("${path.module}/../scripts/init.sql")
    # Re-run when RDS endpoint changes (new database)
    rds_endpoint = aws_db_instance.mysql.endpoint
  }

  provisioner "local-exec" {
    command = <<-EOF
      set -e
      
      echo "Waiting for RDS to be ready..."
      sleep 30
      
      echo "Running database initialization script..."
      mysql -h ${element(split(":", aws_db_instance.mysql.endpoint), 0)} \
            -P 3306 \
            -u ${var.db_username} \
            -p${var.db_password} \
            ${var.db_name} \
            < ${path.module}/../scripts/init.sql
      
      echo "Database initialized successfully!"
    EOF
  }

  depends_on = [aws_db_instance.mysql]
}
