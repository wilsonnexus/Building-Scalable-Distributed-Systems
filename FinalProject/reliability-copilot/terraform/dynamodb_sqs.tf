resource "aws_dynamodb_table" "events" {
  name         = "${var.project_name}-events"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"
  range_key    = "sk"

  attribute {
    name = "pk"
    type = "S"
  }

  attribute {
    name = "sk"
    type = "S"
  }
}

resource "aws_sqs_queue" "jobs" {
  name                       = "${var.project_name}-jobs"
  visibility_timeout_seconds = 30
  receive_wait_time_seconds  = 20
}