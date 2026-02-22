Instructions
Prepare Credentials

Get your temporary credentials from Learner’s Lab, then configure AWS:

aws configure

After that, set your session token:

aws configure set aws_session_token <YOUR-TEMP-SESSION-TOKEN>

Apply Infrastructure (ECS + ALB + Auto Scaling)

Go into the terraform folder and deploy everything:

cd terraform
terraform init
terraform apply -auto-approve

This creates the ECS service, load balancer, and auto scaling setup.

Get Service Address

For Part II (single instance), you can use the public IP.

Get the public IP:

aws ec2 describe-network-interfaces \
--network-interface-ids $(
aws ecs describe-tasks \
 --cluster $(terraform output -raw ecs_cluster_name) \
 --tasks $(
aws ecs list-tasks \
 --cluster $(terraform output -raw ecs_cluster_name) \
 --service-name $(terraform output -raw ecs_service_name) \
 --query 'taskArns[0]' --output text
) \
 --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" \
 --output text
) \
--query 'NetworkInterfaces[0].Association.PublicIp' \
--output text

Example from my run:

3.229.117.16

Send a test request:

curl http://<PUBLIC-IP>:8080/products/search?q=electronics

This matches the testing step shown in my report (page 2).

CS6650HW6WNReport

Get ALB DNS (Horizontal Scaling — Part III)

For the load-balanced version, use the ALB address instead of the public IP.

terraform output -raw alb_dns_name

Example:

CS6650L2-alb-1187418907.us-east-1.elb.amazonaws.com

Send a test request:

curl http://<ALB-DNS>/products/search?q=electronics

This is how the system was tested in Part III.

CS6650HW6WNReport

Load Testing (Locust)

Run load tests using FastHttpUser.

Single instance test:

locust -f locustfile_fast.py --host http://<PUBLIC-IP>:8080

Load-balanced test:

locust -f locustfile_fast.py --host http://<ALB-DNS>

Examples used in the report:

1 user for verification

5 users for baseline

20 users for breaking point

CS6650HW6WNReport

Clean Up

Destroy all resources when finished:

terraform destroy -auto-approve
