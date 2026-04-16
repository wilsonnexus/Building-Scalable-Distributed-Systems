cd terraform
terraform apply -auto-approve `
  -var="container_image_api=316009999564.dkr.ecr.us-east-1.amazonaws.com/reliability-copilot-api:latest" `
  -var="container_image_worker=316009999564.dkr.ecr.us-east-1.amazonaws.com/reliability-copilot-worker:latest" `
  -var="worker_count=1" `
  -var="enable_worker_autoscaling=true"
cd ..

$BASE_URL="http://reliability-copilot-alb-2050495761.us-east-1.elb.amazonaws.com"

locust -f .\tests\locustfile.py --host $BASE_URL --headless -u 100 -r 20 -t 1m --csv .\results\exp2\autoscale_aws