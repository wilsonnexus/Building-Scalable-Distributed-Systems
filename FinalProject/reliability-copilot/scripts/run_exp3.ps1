docker compose down

$env:WORKERS="2"
$env:INFERENCE_DELAY_MS="2000"
$env:DB_DELAY_MS="0"
$env:FAIL_RATE_PCT="20"
$env:PROTECTION_ENABLED="false"
docker compose up --build -d
locust -f .\tests\locustfile.py --host http://localhost:8080 --headless -u 50 -r 10 -t 1m --csv .\results\exp3\unprotected
docker compose down

$env:WORKERS="2"
$env:INFERENCE_DELAY_MS="2000"
$env:DB_DELAY_MS="0"
$env:FAIL_RATE_PCT="20"
$env:PROTECTION_ENABLED="true"
docker compose up --build -d
locust -f .\tests\locustfile.py --host http://localhost:8080 --headless -u 50 -r 10 -t 1m --csv .\results\exp3\protected
docker compose down

python .\tests\plot_results_exp3.py .\results\exp3
