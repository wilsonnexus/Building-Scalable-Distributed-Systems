$BASE_URL="http://reliability-copilot-alb-2050495761.us-east-1.elb.amazonaws.com"

locust -f .\tests\locustfile.py --host $BASE_URL --headless -u 20 -r 5 -t 1m --csv .\results\exp1\small_aws
locust -f .\tests\locustfile.py --host $BASE_URL --headless -u 100 -r 20 -t 1m --csv .\results\exp1\medium_aws
locust -f .\tests\locustfile.py --host $BASE_URL --headless -u 300 -r 50 -t 1m --csv .\results\exp1\heavy_aws