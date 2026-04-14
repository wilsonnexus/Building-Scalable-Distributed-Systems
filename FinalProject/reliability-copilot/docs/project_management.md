# Project Management

## Team
Wilson Neira — solo project

## How I moved from initial design to final state
I started with a very small local prototype to prove the core idea: take a failure event and return a category, recommendation, and risk score. From there, I expanded the system into a more realistic architecture with:
- asynchronous job handling
- a separate worker
- persistent job/result tracking
- experiment scripts
- more structured project management artifacts

## How I broke the problem down
1. Starter API and first local load test
2. Queue-backed event submission
3. Worker-based background processing
4. Persistent result lookup
5. Bottleneck experiment
6. Scaling experiment
7. Failure injection experiment
8. Final report, video, and lessons learned

## Problems encountered
- scope had to stay realistic for one person
- the starter project was too simple to support the final experiments
- I had to turn the original idea into something measurable instead of just descriptive
- I had to balance AI features with reliability and cost concerns

## Who worked on what
I worked on all parts:
- design
- coding
- experiments
- observability planning
- reporting
- presentation