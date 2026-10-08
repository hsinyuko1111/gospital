# ============================================================================
# ECS CLUSTER
# ============================================================================

resource "aws_ecs_cluster" "main" {
  name = "${var.project_name}-cluster"

  setting {
    name  = "containerInsights"
    value = "enhanced"
  }

  tags = {
    Name = "${var.project_name}-cluster"
  }
}

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/${var.project_name}"
  retention_in_days = 7

  tags = {
    Name = "${var.project_name}-ecs-logs"
  }
}

# ============================================================================
# INTERNAL SERVICE (Staff Operations)
# ============================================================================

resource "aws_ecs_task_definition" "internal" {
  family                   = "${var.project_name}-internal"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.container_cpu
  memory                   = var.container_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([
    {
      name      = "api"
      image     = "${aws_ecr_repository.gospital.repository_url}:latest"
      essential = true

      portMappings = [
        {
          containerPort = var.container_port
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "MYSQL_DSN"
          value = "${var.db_username}:${var.db_password}@tcp(${aws_db_instance.mysql.endpoint})/${var.db_name}?parseTime=true"
        },
        {
          name  = "DYNAMODB_REGION"
          value = var.aws_region
        },
        {
          name  = "MEDICAL_RECORDS_TABLE"
          value = var.medical_records_table
        },
        {
          name  = "INVOICES_TABLE"
          value = var.invoices_table
        },
        {
          name  = "SERVER_PORT"
          value = tostring(var.container_port)
        },
        {
          name  = "SERVICE_TYPE"
          value = "internal"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "internal"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:${var.container_port}/health || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }
    }
  ])

  tags = {
    Name = "${var.project_name}-internal-task"
  }
}

resource "aws_ecs_service" "internal" {
  name            = "${var.project_name}-internal-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.internal.arn
  desired_count   = var.internal_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private_ecs[*].id
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.internal.arn
    container_name   = "api"
    container_port   = var.container_port
  }

  service_registries {
    registry_arn = aws_service_discovery_service.internal.arn
  }

  # FIXED: Increased grace period to account for container startPeriod (60s)
  # plus ALB health check interval (30s) plus buffer
  health_check_grace_period_seconds = 120

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100

  # FIXED: Added deployment circuit breaker for automatic rollback on failed deployments
  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  lifecycle {
    ignore_changes = [desired_count]
  }

  depends_on = [
    docker_registry_image.gospital,
    null_resource.mysql_init,
    aws_lb_listener.http,
    aws_db_instance.mysql,
    aws_dynamodb_table.medical_records,
    aws_dynamodb_table.invoices,
    aws_service_discovery_service.internal,
    aws_service_discovery_private_dns_namespace.main
  ]

  tags = {
    Name = "${var.project_name}-internal-service"
  }
}

resource "aws_appautoscaling_target" "internal" {
  max_capacity       = var.internal_max_tasks
  min_capacity       = var.internal_min_tasks
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.internal.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "internal_cpu" {
  name               = "${var.project_name}-internal-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.internal.resource_id
  scalable_dimension = aws_appautoscaling_target.internal.scalable_dimension
  service_namespace  = aws_appautoscaling_target.internal.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }

    target_value       = var.autoscaling_cpu_target
    scale_in_cooldown  = var.autoscaling_scale_in_cooldown
    scale_out_cooldown = 60
  }
}

resource "aws_appautoscaling_policy" "internal_memory" {
  name               = "${var.project_name}-internal-memory-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.internal.resource_id
  scalable_dimension = aws_appautoscaling_target.internal.scalable_dimension
  service_namespace  = aws_appautoscaling_target.internal.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }

    target_value       = var.autoscaling_memory_target
    scale_in_cooldown  = var.autoscaling_scale_in_cooldown
    scale_out_cooldown = 60
  }
}

# ============================================================================
# EXTERNAL SERVICE (Public/Patient Operations)
# ============================================================================

resource "aws_ecs_task_definition" "external" {
  family                   = "${var.project_name}-external"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.container_cpu
  memory                   = var.container_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([
    {
      name      = "api"
      image     = "${aws_ecr_repository.gospital.repository_url}:latest"
      essential = true

      portMappings = [
        {
          containerPort = var.container_port
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "MYSQL_DSN"
          value = "${var.db_username}:${var.db_password}@tcp(${aws_db_instance.mysql.endpoint})/${var.db_name}?parseTime=true"
        },
        {
          name  = "DYNAMODB_REGION"
          value = var.aws_region
        },
        {
          name  = "MEDICAL_RECORDS_TABLE"
          value = var.medical_records_table
        },
        {
          name  = "INVOICES_TABLE"
          value = var.invoices_table
        },
        {
          name  = "SERVER_PORT"
          value = tostring(var.container_port)
        },
        {
          name  = "SERVICE_TYPE"
          value = "external"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "external"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:${var.container_port}/health || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }
    }
  ])

  tags = {
    Name = "${var.project_name}-external-task"
  }
}

resource "aws_ecs_service" "external" {
  name            = "${var.project_name}-external-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.external.arn
  desired_count   = var.external_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private_ecs[*].id
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.external.arn
    container_name   = "api"
    container_port   = var.container_port
  }

  service_registries {
    registry_arn = aws_service_discovery_service.external.arn
  }

  # FIXED: Increased grace period to account for container startPeriod (60s)
  # plus ALB health check interval (30s) plus buffer
  health_check_grace_period_seconds = 120

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100

  # FIXED: Added deployment circuit breaker for automatic rollback on failed deployments
  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  lifecycle {
    ignore_changes = [desired_count]
  }

  depends_on = [
    docker_registry_image.gospital,
    null_resource.mysql_init,
    aws_lb_listener.http,
    aws_db_instance.mysql,
    aws_dynamodb_table.medical_records,
    aws_dynamodb_table.invoices,
    aws_service_discovery_service.external,
    aws_service_discovery_private_dns_namespace.main
  ]

  tags = {
    Name = "${var.project_name}-external-service"
  }
}

resource "aws_appautoscaling_target" "external" {
  max_capacity       = var.external_max_tasks
  min_capacity       = var.external_min_tasks
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.external.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "external_cpu" {
  name               = "${var.project_name}-external-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.external.resource_id
  scalable_dimension = aws_appautoscaling_target.external.scalable_dimension
  service_namespace  = aws_appautoscaling_target.external.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }

    target_value       = var.autoscaling_cpu_target
    scale_in_cooldown  = var.autoscaling_scale_in_cooldown
    scale_out_cooldown = 60
  }
}

resource "aws_appautoscaling_policy" "external_memory" {
  name               = "${var.project_name}-external-memory-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.external.resource_id
  scalable_dimension = aws_appautoscaling_target.external.scalable_dimension
  service_namespace  = aws_appautoscaling_target.external.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }

    target_value       = var.autoscaling_memory_target
    scale_in_cooldown  = var.autoscaling_scale_in_cooldown
    scale_out_cooldown = 60
  }
}
