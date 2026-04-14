docker compose down
docker compose up --build -d

locust -f .\tests\locustfile.py --host http://localhost:8080 --headless -u 20 -r 5 -t 1m --csv .\results\exp1\small
locust -f .\tests\locustfile.py --host http://localhost:8080 --headless -u 100 -r 20 -t 1m --csv .\results\exp1\medium
locust -f .\tests\locustfile.py --host http://localhost:8080 --headless -u 300 -r 50 -t 1m --csv .\results\exp1\heavy

python .\tests\plot_results.py .\results\exp1