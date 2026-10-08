# ============================================================================
# APPLICATION LOAD BALANCER (Multi-AZ)
# ============================================================================

resource "aws_lb" "main" {
  name               = "${var.project_name}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = false
  enable_http2               = true

  tags = {
    Name = "${var.project_name}-alb"
  }
}

# ============================================================================
# TARGET GROUPS
# ============================================================================

resource "aws_lb_target_group" "internal" {
  name        = "${var.project_name}-internal-tg"
  port        = var.container_port
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    path                = "/health"
    matcher             = "200"
  }

  deregistration_delay = 30

  tags = {
    Name = "${var.project_name}-internal-tg"
  }
}

resource "aws_lb_target_group" "external" {
  name        = "${var.project_name}-external-tg"
  port        = var.container_port
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 3
    timeout             = 5
    interval            = 30
    path                = "/health"
    matcher             = "200"
  }

  deregistration_delay = 30

  tags = {
    Name = "${var.project_name}-external-tg"
  }
}

# ============================================================================
# ALB LISTENER & RULES
# ============================================================================

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.external.arn
  }
}

resource "aws_lb_listener_rule" "internal" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.internal.arn
  }

  condition {
    http_header {
      http_header_name = "X-User-Type"
      values           = ["staff"]
    }
  }

  tags = {
    Name = "${var.project_name}-internal-rule"
  }
}

resource "aws_lb_listener_rule" "external" {
  listener_arn = aws_lb_listener.http.arn
  priority     = 200

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.external.arn
  }

  condition {
    http_header {
      http_header_name = "X-User-Type"
      values           = ["public"]
    }
  }

  tags = {
    Name = "${var.project_name}-external-rule"
  }
}
