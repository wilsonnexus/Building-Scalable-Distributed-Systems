# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "ecr_service"
}

variable "service_name" {
  type    = string
  default = "CS6650L2"
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "ecs_count" {
  type    = number
  default = 2
  description = "Desired Fargate task count"
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

variable "app_mode" {
  type = string
  default = "bad"
}

variable "image_tag" {
  type    = string
  default = "dev"
}


