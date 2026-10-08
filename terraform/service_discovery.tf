# ============================================================================
# SERVICE DISCOVERY (for ADOT Collector to discover ECS tasks)
# ============================================================================

resource "aws_service_discovery_private_dns_namespace" "main" {
  name        = var.project_name
  description = "Service discovery namespace for ${var.project_name}"
  vpc         = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-namespace"
  }
}

resource "aws_service_discovery_service" "internal" {
  name = "${var.project_name}-internal-service"

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
    Name = "${var.project_name}-internal-discovery"
  }
}

resource "aws_service_discovery_service" "external" {
  name = "${var.project_name}-external-service"

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
    Name = "${var.project_name}-external-discovery"
  }
}
