# ============================================================================
# GRAFANA CONFIGURATION
# ============================================================================

locals {
  grafana_datasource_provisioning = <<-EOT
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    uid: prometheus
    url: http://${var.project_name}-prometheus.${var.project_name}:9090
    access: proxy
    isDefault: true
EOT

  grafana_dashboard_provisioning = <<-EOT
apiVersion: 1
providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    editable: true
    options:
      path: /var/lib/grafana/dashboards
EOT

  grafana_dashboard_json = file("${path.module}/../scripts/dashboard.json")
}

# ============================================================================
# GRAFANA ECS TASK DEFINITION
# ============================================================================

resource "aws_ecs_task_definition" "grafana" {
  family                   = "${var.project_name}-grafana"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name      = "grafana"
      image     = "grafana/grafana:10.4.0"
      essential = true

      portMappings = [
        {
          containerPort = 3000
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "GF_AUTH_ANONYMOUS_ENABLED"
          value = "true"
        },
        {
          name  = "GF_AUTH_ANONYMOUS_ORG_ROLE"
          value = "Admin"
        },
        {
          name  = "GF_SECURITY_ADMIN_PASSWORD"
          value = "admin"
        },
        {
          name  = "GF_SERVER_ROOT_URL"
          value = "%(protocol)s://%(domain)s/grafana/"
        },
        {
          name  = "GF_SERVER_SERVE_FROM_SUB_PATH"
          value = "true"
        },
        {
          name  = "GF_DASHBOARDS_DEFAULT_HOME_DASHBOARD_PATH"
          value = "/var/lib/grafana/dashboards/gospital-overview.json"
        },
        {
          name  = "DATASOURCE_PROVISIONING"
          value = base64encode(local.grafana_datasource_provisioning)
        },
        {
          name  = "DASHBOARD_PROVISIONING"
          value = base64encode(local.grafana_dashboard_provisioning)
        },
        {
          name  = "DASHBOARD_JSON"
          value = base64encode(local.grafana_dashboard_json)
        }
      ]

      entryPoint = ["/bin/sh", "-c"]
      command = [
        "mkdir -p /etc/grafana/provisioning/datasources /etc/grafana/provisioning/dashboards /var/lib/grafana/dashboards && echo $DATASOURCE_PROVISIONING | base64 -d > /etc/grafana/provisioning/datasources/prometheus.yml && echo $DASHBOARD_PROVISIONING | base64 -d > /etc/grafana/provisioning/dashboards/default.yml && echo $DASHBOARD_JSON | base64 -d > /var/lib/grafana/dashboards/gospital-overview.json && /run.sh"
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "grafana"
        }
      }
    }
  ])

  tags = {
    Name = "${var.project_name}-grafana-task"
  }
}

# ============================================================================
# GRAFANA ECS SERVICE
# ============================================================================

resource "aws_ecs_service" "grafana" {
  name            = "${var.project_name}-grafana"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.grafana.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private_ecs[*].id
    security_groups  = [aws_security_group.grafana.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.grafana.arn
    container_name   = "grafana"
    container_port   = 3000
  }

  depends_on = [
    aws_ecs_service.prometheus,
    aws_lb_listener.http
  ]

  tags = {
    Name = "${var.project_name}-grafana"
  }
}

# ============================================================================
# GRAFANA SECURITY GROUP
# ============================================================================

resource "aws_security_group" "grafana" {
  name        = "${var.project_name}-grafana-sg"
  description = "Security group for Grafana"
  vpc_id      = aws_vpc.main.id

  ingress {
    description     = "Grafana from ALB"
    from_port       = 3000
    to_port         = 3000
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-grafana-sg"
  }
}

# ============================================================================
# GRAFANA ALB TARGET GROUP
# ============================================================================

resource "aws_lb_target_group" "grafana" {
  name        = "${var.project_name}-grafana-tg"
  port        = 3000
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    path                = "/grafana/api/health"
    matcher             = "200"
  }

  deregistration_delay = 30

  tags = {
    Name = "${var.project_name}-grafana-tg"
  }
}

# ============================================================================
# GRAFANA ALB LISTENER RULE
# ============================================================================

resource "aws_lb_listener_rule" "grafana" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 50

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.grafana.arn
  }

  condition {
    path_pattern {
      values = ["/grafana", "/grafana/*"]
    }
  }

  tags = {
    Name = "${var.project_name}-grafana-rule"
  }
}
