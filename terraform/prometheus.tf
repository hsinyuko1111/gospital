# ============================================================================
# PROMETHEUS CONFIGURATION
# ============================================================================

locals {
  prometheus_config = <<-EOT
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'gospital-internal'
    dns_sd_configs:
      - names:
          - '${var.project_name}-internal-service.${var.project_name}'
        type: 'A'
        port: ${var.container_port}
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
      - target_label: service
        replacement: 'internal'

  - job_name: 'gospital-external'
    dns_sd_configs:
      - names:
          - '${var.project_name}-external-service.${var.project_name}'
        type: 'A'
        port: ${var.container_port}
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
      - target_label: service
        replacement: 'external'
EOT
}

# ============================================================================
# PROMETHEUS ECS TASK DEFINITION
# ============================================================================

resource "aws_ecs_task_definition" "prometheus" {
  family                   = "${var.project_name}-prometheus"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name      = "prometheus"
      image     = "prom/prometheus:v2.47.0"
      essential = true

      portMappings = [
        {
          containerPort = 9090
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "PROMETHEUS_CONFIG"
          value = base64encode(local.prometheus_config)
        }
      ]

      entryPoint = ["/bin/sh", "-c"]
      command = [
        "echo $PROMETHEUS_CONFIG | base64 -d > /etc/prometheus/prometheus.yml && /bin/prometheus --config.file=/etc/prometheus/prometheus.yml --storage.tsdb.path=/prometheus --storage.tsdb.retention.time=1h --web.enable-lifecycle"
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "prometheus"
        }
      }
    }
  ])

  tags = {
    Name = "${var.project_name}-prometheus-task"
  }
}

# ============================================================================
# PROMETHEUS ECS SERVICE
# ============================================================================

resource "aws_ecs_service" "prometheus" {
  name            = "${var.project_name}-prometheus"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.prometheus.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private_ecs[*].id
    security_groups  = [aws_security_group.prometheus.id]
    assign_public_ip = false
  }

  service_registries {
    registry_arn = aws_service_discovery_service.prometheus.arn
  }

  depends_on = [
    aws_service_discovery_private_dns_namespace.main,
    aws_ecs_service.internal,
    aws_ecs_service.external
  ]

  tags = {
    Name = "${var.project_name}-prometheus"
  }
}

# ============================================================================
# PROMETHEUS SERVICE DISCOVERY
# ============================================================================

resource "aws_service_discovery_service" "prometheus" {
  name = "${var.project_name}-prometheus"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }

  tags = {
    Name = "${var.project_name}-prometheus-discovery"
  }
}

# ============================================================================
# PROMETHEUS SECURITY GROUP
# ============================================================================

resource "aws_security_group" "prometheus" {
  name        = "${var.project_name}-prometheus-sg"
  description = "Security group for Prometheus"
  vpc_id      = aws_vpc.main.id

  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-prometheus-sg"
  }
}

resource "aws_security_group_rule" "prometheus_from_grafana" {
  type                     = "ingress"
  from_port                = 9090
  to_port                  = 9090
  protocol                 = "tcp"
  security_group_id        = aws_security_group.prometheus.id
  source_security_group_id = aws_security_group.grafana.id
  description              = "Prometheus from Grafana"
}

resource "aws_security_group_rule" "ecs_from_prometheus" {
  type                     = "ingress"
  from_port                = var.container_port
  to_port                  = var.container_port
  protocol                 = "tcp"
  security_group_id        = aws_security_group.ecs_tasks.id
  source_security_group_id = aws_security_group.prometheus.id
  description              = "Allow Prometheus to scrape metrics"
}
