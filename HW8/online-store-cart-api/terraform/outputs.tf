output "alb_dns_name" {
  value = aws_lb.main.dns_name
}

output "mysql_endpoint" {
  value = aws_db_instance.mysql.address
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.shopping_carts.name
}

output "ecs_cluster_name" {
  value = aws_ecs_cluster.main.name
}