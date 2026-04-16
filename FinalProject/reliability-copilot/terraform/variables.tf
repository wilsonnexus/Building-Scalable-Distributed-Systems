variable "aws_region" {
  default = "us-east-1"
}

variable "project_name" {
  default = "reliability-copilot"
}

variable "container_image_api" {
  type = string
}

variable "container_image_worker" {
  type = string
}

variable "worker_count" {
  default = 1
}

variable "enable_worker_autoscaling" {
  default = false
}