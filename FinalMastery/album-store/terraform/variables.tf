variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "repo_url" {
  type = string
}

variable "instance_type" {
  type    = string
  default = "t3.small"
}

variable "ssh_cidr" {
  type    = string
  default = "0.0.0.0/0"
}

variable "key_name" {
  type    = string
  default = ""
}