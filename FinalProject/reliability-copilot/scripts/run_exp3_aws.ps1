$BASE_URL="http://reliability-copilot-alb-2050495761.us-east-1.elb.amazonaws.com"

locust -f .\tests\locustfile.py --host $BASE_URL --headless -u 50 -r 10 -t 1m --csv .\results\exp3\unprotected_aws