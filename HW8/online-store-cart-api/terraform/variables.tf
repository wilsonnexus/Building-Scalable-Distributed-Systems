variable "aws_region" {
  default = "us-east-1"
}

variable "project_name" {
  default = "cartapi"
}

variable "container_image" {
  type = string
}

variable "db_username" {
  default = "cartuser"
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "db_name" {
  default = "cartdb"
}