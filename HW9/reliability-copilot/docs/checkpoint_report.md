# Reliability Copilot Checkpoint Report

## 1. Problem, Team, and Overview of Experiments

My project is Reliability Copilot, a small but realistic AWS-based platform for analyzing simulated CI/CD and build/test failures. The system takes in a failure event and returns a failure category, a short recommended action, and a risk score for whether the same kind of failure is likely to happen again.

This problem matters because teams lose time when failures are noisy, unclear, or repetitive. A system like this could help engineers triage failures faster and make reliability problems easier to understand.

I am working solo on this project. I am not currently looking for more teammates because my schedule is heavy, so I am keeping the project focused and manageable for one person.

The main experiments I plan to run are:

1. bottleneck identification
2. scaling behavior under increased load
3. failure injection and recovery

The results I will evaluate include latency, throughput, queue depth, CPU and memory usage, and system behavior during failure conditions. AI will play the role of lightweight classification and recommendation support, while observability will come from logs, metrics, and later CloudWatch dashboards.

## 2. Project Plan and Recent Progress

Recent progress includes refining the architecture, defining the three core experiments, creating the starter repository, implementing an initial local API prototype, creating a Locust-based load test, and generating first charts for latency and throughput.

Since I am working solo, I am responsible for all parts of the project, including architecture, coding, testing, infrastructure, observability, reporting, poster creation, and presentation.

My plan from now until April 19 is:

- finish the checkpoint materials first
- add the queue and worker flow
- collect baseline bottleneck results
- test scaling behavior
- test failure injection and recovery
- finish the final report, poster, and presentation

I am using AI within the project in two ways. First, the system itself will include lightweight AI or LLM-assisted failure analysis. Second, AI helps me move faster in project development by assisting with drafting, iteration, and code planning. The benefit is faster prototyping and clearer recommendations. The cost is that I still need to validate outputs carefully, since the real value of the project comes from system design, measurements, and engineering judgment.

## 3. Objectives

My short-term objective is to build a working prototype that can ingest failure events, analyze them, and produce measurable system behavior under load. My longer-term objective is to turn this into a stronger portfolio project that shows how distributed systems, reliability engineering, and AI can work together in a realistic platform.

Observability is a core objective, not an extra feature. I want the final project to expose enough logs and metrics that stakeholders can actually understand what is happening when the system slows down or fails.

Because the project includes AI components, I also want to control performance, reliability, and cost. That means keeping the AI path lightweight, adding fallback behavior if inference becomes slow, and designing the system so that AI does not become a single point of failure.

## 4. Related Work

This project connects strongly to distributed systems ideas from the course readings. One important lesson is that distributed systems must be designed with the expectation of failure, rather than assuming that the network and components will always behave well. Another is that latency, availability, and consistency are shaped by real system constraints and tradeoffs.

The project also connects to common distributed systems goals such as scalability, fault tolerance, predictable performance, and recovery from failures. These ideas are directly relevant because Reliability Copilot is not only classifying events, but also trying to remain responsive and observable under changing conditions.

## 5. Methodology

My methodology is to build the project in layers. First, I will create a basic API that accepts failure events and produces a baseline classification and recommendation. Then I will add queue-based decoupling, a worker path, storage, and observability. After that, I will run experiments that stress the system and measure its behavior.

The three core experiments are:

1. bottleneck identification under increasing load
2. scaling behavior under higher worker capacity or scaling rules
3. failure injection and recovery under conditions such as slow AI, artificial storage delay, or downstream issues

I will use AI in a lightweight way, likely starting with rule-assisted classification or a small inference path before expanding to a fuller LLM-assisted recommendation step. I will support observability with request logs, latency measurements, throughput tracking, queue depth tracking, and infrastructure-level metrics once the AWS version is in place.

The main tradeoffs I am evaluating are:

- latency vs richer analysis
- simplicity vs resilience
- stronger processing guarantees vs faster responsiveness
- AI usefulness vs AI cost and slowness

## 6. Preliminary Results

So far, I have an initial local prototype and a first round of load testing results. These initial results are enough to produce starter throughput and latency charts, which show that the project already has a measurable behavior that I can improve and compare later.

What is still left is the more realistic distributed version with queue-based processing, cloud deployment, stronger observability, and the final experiment set. The worst-case workload for this project is a sudden burst of many failure events at once, especially if the AI path or storage path is slow. The base-case workload is a steady stream of modest event traffic with low contention.

## 7. Impact

I think people would care about this project because it connects several things that are useful in real engineering work: platform reliability, distributed systems design, observability, and practical AI assistance.

This could also be useful to other students if they want to help test the API or compare failure patterns. Even as a course project, it has value as a portfolio piece because it is not just another CRUD app or isolated AI demo. It is a system with measurable tradeoffs and stakeholder value.
