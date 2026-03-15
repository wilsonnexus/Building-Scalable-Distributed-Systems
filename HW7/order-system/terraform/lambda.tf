data "aws_iam_role" "lab_role_lambda" {
  name = "LabRole"
}

resource "aws_lambda_function" "order_processor" {
  function_name    = "ordersys-lambda-processor"
  role             = data.aws_iam_role.lab_role_lambda.arn
  handler          = "bootstrap"
  runtime          = "provided.al2"
  filename         = "${path.module}/../src/lambda.zip"
  source_code_hash = filebase64sha256("${path.module}/../src/lambda.zip")
  memory_size      = 512
  timeout          = 10
}

resource "aws_sns_topic_subscription" "orders_to_lambda" {
  topic_arn = aws_sns_topic.orders.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.order_processor.arn
}

resource "aws_lambda_permission" "allow_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.order_processor.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.orders.arn
}