# Build Docker image for Gospital API
resource "docker_image" "gospital" {
  name = "${aws_ecr_repository.gospital.repository_url}:latest"

  build {
    context    = "${path.module}/.."
    dockerfile = "Dockerfile"
    platform   = "linux/amd64"

    # Remove intermediate containers after build
    remove = true

    # Use build cache for faster builds
    no_cache = false
  }

  # Rebuild triggers
  triggers = {
    dockerfile_hash = filemd5("${path.module}/../Dockerfile")
    gomod_hash      = filemd5("${path.module}/../go.mod")
  }

  depends_on = [aws_ecr_repository.gospital]
}

# Push Docker image to ECR
resource "docker_registry_image" "gospital" {
  name = docker_image.gospital.name

  # Delete image from ECR when Terraform destroys this resource
  keep_remotely = false

  depends_on = [docker_image.gospital]
}
