resource "aws_dynamodb_table" "shopping_carts" {
  name         = "shopping_carts"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "cart_id"

  attribute {
    name = "cart_id"
    type = "S"
  }
}