variable "aws_region" {
  default = "us-east-1"
}

variable "project_name" {
  default = "ordersys"
}

variable "container_image" {
  type = string
}

variable "vpc_cidr" {
  default = "10.0.0.0/16"
}

variable "public_subnet_1" {
  default = "10.0.1.0/24"
}

variable "public_subnet_2" {
  default = "10.0.2.0/24"
}

variable "private_subnet_1" {
  default = "10.0.10.0/24"
}

variable "private_subnet_2" {
  default = "10.0.11.0/24"
}