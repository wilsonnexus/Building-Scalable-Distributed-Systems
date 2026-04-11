output "public_ip" {
  value = aws_instance.album_store.public_ip
}

output "public_dns" {
  value = aws_instance.album_store.public_dns
}

output "base_url" {
  value = "http://${aws_instance.album_store.public_dns}"
}