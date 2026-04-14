docker compose down

$env:WORKERS="1"
docker compose up --build -d
locust -f .\tests\locustfile.py --host http://localhost:8080 --headless -u 100 -r 20 -t 1m --csv .\results\exp2\workers1
docker compose down

$env:WORKERS="4"
docker compose up --build -d
locust -f .\tests\locustfile.py --host http://localhost:8080 --headless -u 100 -r 20 -t 1m --csv .\results\exp2\workers4
docker compose down

$env:WORKERS="8"
docker compose up --build -d
locust -f .\tests\locustfile.py --host http://localhost:8080 --headless -u 100 -r 20 -t 1m --csv .\results\exp2\workers8
docker compose down

python .\tests\plot_results_exp2.py .\results\exp2
