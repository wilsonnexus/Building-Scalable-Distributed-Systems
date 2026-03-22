import json

with open("mysql_test_results.json", "r", encoding="utf-8") as f:
    mysql_data = json.load(f)

with open("dynamodb_test_results.json", "r", encoding="utf-8") as f:
    dynamodb_data = json.load(f)

assert len(mysql_data) == 150, f"MySQL count is {len(mysql_data)} not 150"
assert len(dynamodb_data) == 150, f"DynamoDB count is {len(dynamodb_data)} not 150"

combined = {
    "mysql": mysql_data,
    "dynamodb": dynamodb_data
}

with open("combined_results.json", "w", encoding="utf-8") as f:
    json.dump(combined, f, indent=2)

print("combined_results.json created")