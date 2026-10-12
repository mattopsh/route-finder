# Route Finder

Find routes from one source to multiple destinations. Results are sorted by duration, then distance.

## Prerequisites

Make sure you have the following installed:

- **Go** (version specified in `go.mod`)
- **make**
- **curl** (to test the API)

Then, download the project dependencies:

```bash
go mod download
```

## Run

**Start the server:**

```bash
make run
```

**Test the API:**

```bash
curl 'http://localhost:8080/routes?src=13.388860,52.517037&dst=13.397634,52.529407&dst=13.428555,52.523219'
```

## Development

```bash
make build # Build the project
make test  # Run tests
make vet   # Run static analysis
```

## Trade-offs

- **All or nothing:** Any failed OSRM request causes the entire API request to fail. Alternatives include retrying transient failures with exponential backoff or returning partial results with per-destination errors, which would require changes to the response contract.
- **Bounded fan-out and request size:** Destination lookups run concurrently, with concurrency limited by `OSRM_MAX_CONCURRENCY` per request. The API also limits each request to 50 destinations to bound response size, latency, and upstream load.
- **No repository layer:** Using a consumer-side interface keeps the OSRM client replaceable without introducing an additional abstraction. A repository layer could be added later as complexity grows.
- **No DI framework:** Dependencies are wired manually, which is sufficient at the current scale. A dependency injection framework could be considered as the application grows in complexity.

## Scalability

- **Stateless service:** Supports horizontal scaling behind a load balancer.
- **Upstream capacity:** OSRM latency and capacity may become bottlenecks as traffic increases.
- **No global concurrency limit:** Concurrent API requests and multiple service replicas can collectively generate excessive traffic to OSRM. A distributed rate limiter could enforce a shared upstream request budget across replicas, but it would introduce additional infrastructure and operational complexity.
- **Caching:** Caching could reduce upstream requests. However, approximate location matching might be needed to improve cache hit rates.

## System Resilience

- **Timeouts:** OSRM calls use `OSRM_TIMEOUT_MS` and request contexts to limit request duration.
- **Cancellation:** Context cancellation propagates to in-flight calls when the client cancels a request or the first error occurs.
- **Error classification:** No route found results in a `404`, upstream failures result in a `502`, and upstream timeouts result in a `504`.
- **Current limitations:** There are no retries, circuit breakers, or partial-result fallbacks.

## Full-Scale Implementation

Depending on traffic, requirements, and the deployment environment, I would consider the following improvements:

- **API and security:** Authentication or usage plan tracking with API keys, rate limiting, a CORS policy, OpenAPI documentation, and WAF/DDoS protection where appropriate.
- **Observability:** Structured logging (e.g. using slog), correlation IDs, metrics, distributed tracing, dashboards, alerts, clearly assigned on-call responsibilities, and incident response runbooks.
- **Reliability:** Bounded retries with exponential backoff, graceful shutdown, health checks, circuit breakers, fallbacks or partial-failures.
- **Testing:** The project currently has only unit tests. I would add integration tests (e.g., using Testcontainers), load/spike/stress testing (e.g., using Gatling or k6), and fault-injection testing (e.g., using AWS Fault Injection Service).
- **Deployment and infrastructure:** CI/CD, containerization, infrastructure as code, capacity planning, autoscaling, resource right-sizing, and no downtime deployment strategy e.g. blue-green.
- **Configuration and secrets:** Secrets manager if sensitive configuration is introduced.
- **Code quality:** Consistent formatting, linting, test coverage, pre-commit checks, and an `AGENTS.md` file containing specific guidelines for AI development.