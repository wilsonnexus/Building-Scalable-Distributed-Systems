$ACCOUNT_ID = "316009999564"
$REGION = "us-east-1"

$API_REPO = "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/reliability-copilot-api"
$WORKER_REPO = "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/reliability-copilot-worker"

aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com"

docker build -f Dockerfile.api -t reliability-copilot-api .
docker tag reliability-copilot-api:latest "${API_REPO}:latest"
docker push "${API_REPO}:latest"

docker build -f Dockerfile.worker -t reliability-copilot-worker .
docker tag reliability-copilot-worker:latest "${WORKER_REPO}:latest"
docker push "${WORKER_REPO}:latest"