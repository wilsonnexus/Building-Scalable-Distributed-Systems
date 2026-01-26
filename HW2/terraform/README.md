# web-service-gin (Go + Gin REST API on AWS EC2)

Author: Wilson Neira
Course: CS 6650 – Building Scalable Distributed Systems
Week: 1 (AWS Learner Lab)

---

## OVERVIEW

This project is a small RESTful web service written in Go using the Gin framework.
It was originally developed locally (Homework 1a) and then deployed to a remote
AWS EC2 instance as part of the AWS Learner Lab.

The goal of this work was NOT to build a complex backend, but to learn:

- How to run a server on a remote virtual machine
- How networking differs between local and cloud environments
- How to cross-compile binaries across operating systems
- How SSH, key formats, and file transfer work in practice
- How AWS Security Groups affect connectivity

---

## API FUNCTIONALITY

The service exposes a simple in-memory “albums” API:

- GET /albums → list all albums
- GET /albums/:id → fetch an album by ID
- POST /albums → add a new album (with validation)

Data is stored in memory only. Restarting the server resets the data.

---

## LOCAL DEVELOPMENT (WINDOWS)

Prerequisites:

- Go (1.16+ recommended)
- Git Bash or PowerShell
- Internet access (for Go module downloads)

Run locally:
go get .
go run .

The server listens on:
http://localhost:8080

Test locally:
curl http://localhost:8080/albums
curl http://localhost:8080/albums/2

---

## IMPORTANT SERVER CHANGE (LOCAL → CLOUD)

Originally, the server was bound to:
localhost:8080

This only accepts connections from the same machine.
When running on EC2, this prevented access from my laptop.

The fix was to bind to:
0.0.0.0:8080

This allows the server to listen on all network interfaces,
making it reachable through the EC2 public IP.

---

## CROSS-COMPILING FOR AWS (WINDOWS → LINUX)

My development machine runs Windows, while EC2 runs Amazon Linux.
To run the server on EC2, I cross-compiled the binary:

    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o web-service-gin ./main.go

This produced a Linux ELF executable that can run directly on EC2.

---

## CONNECTING TO EC2 (SERVER SIDE)

From Windows using Git Bash + PuTTY/Plink:

    plink -i "C:\Users\Owner\Documents\Audacity\Northeastern University\CS 6650 Building Scalable Distributed Systems\Week 1\web-service-gin.ppk"       ec2-user@ec2-16-148-46-4.us-west-2.compute.amazonaws.com

On the EC2 instance:

    cd web-service-gin
    ./web-service-gin

If permission is denied:

    chmod 777 web-service-gin
    ls -l web-service-gin
    ./web-service-gin

Local test on EC2:

    curl localhost:8080/albums

---

## UPLOADING THE BINARY (FILE TRANSFER)

PuTTY uses .ppk keys, but scp requires OpenSSH-compatible .pem keys.
I converted my .ppk key to .pem to upload files.

Upload command from Windows:

    scp -i web-service-gin.pem ./web-service-gin       ec2-user@ec2-16-148-46-4.us-west-2.compute.amazonaws.com:/home/ec2-user/web-service-gin/

---

## CLIENT-SIDE TESTING (FROM MY MACHINE)

Once the server was running on EC2 and port 8080 was opened in the
Security Group, I tested remotely:

    curl http://16.148.46.4:8080/albums

This confirmed end-to-end connectivity from my laptop to EC2.

---

## AWS SECURITY GROUPS

AWS Security Groups act as a virtual firewall.

Inbound rules used:

- SSH (port 22) → allowed from My IP
- Custom TCP (port 8080) → allowed for HTTP testing

Without opening port 8080, the server would run but be unreachable.

---

## Notes: SCREENSHOTS & OBSERVATIONS

Screenshots included with submission show:

- Part1Connection.png: Successful SSH connection to EC2
- Part1Connection.png: The Gin server running and listening on 0.0.0.0:8080
- Part1Output.png: Successful curl responses remotely
- Part1Output.png: Cross-compile code and copying file to EC2
- Part1EC2Instance.png: AWS EC2 instance details (public IP, instance state)
- Part3LoadTest.png: Load test output
- Part3LoadTestDistribution.png: Response-time distribution

From these screenshots, it is clear that:

- The service was correctly deployed
- Networking and security rules were configured properly
- The same binary ran consistently across environments

---

## Notes: LOAD TESTING

I also ran a small Python-based load test from my local machine
to observe response times.

Observations Part3LoadTestDistribution.png:

- Most responses clustered around ~180–220 ms
- Occasional spikes (~500 ms) were visible
- Performance was stable for a single t2.micro instance
- Results are expected given no caching and limited resources
