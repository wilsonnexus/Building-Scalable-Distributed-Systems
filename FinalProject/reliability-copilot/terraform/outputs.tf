output "alb_dns_name" {
  value = aws_lb.main.dns_name
}

output "sqs_queue_url" {
  value = aws_sqs_queue.jobs.id
}

output "ddb_table_name" {
  value = aws_dynamodb_table.events.name
}