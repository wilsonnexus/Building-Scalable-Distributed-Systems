variable "aws_region" {
  type    = string
  default = "us-west-2"
}

variable "repo_url" {
  type = string
}

variable "instance_type" {
  type    = string
  default = "t3.large"
}

variable "ssh_cidr" {
  type    = string
  default = "0.0.0.0/0"
}

variable "key_name" {
  type    = string
  default = ""
}