# The Biggie

The BIG application for exercising HA and DR.  
Written in Golang.

---

## Table of Contents

- [For ALL APIs](#for-all-apis)
- [API Endpoints](#api-endpoints)
  - [Basic APIs](#basic-apis)
  - [Health & Metadata APIs](#health--metadata-apis)
  - [Stress Test APIs](#stress-test-apis)
  - [Heavy Database Activities](#heavy-database-activities)
    - [MySQL APIs](#mysql-apis)
    - [PostgreSQL APIs](#postgresql-apis)
    - [Redshift APIs](#redshift-apis)
    - [Redis APIs](#redis-apis)
    - [Kafka APIs](#kafka-apis)
  - [Error Injection APIs](#error-injection-apis)
  - [Service Management APIs](#service-management-apis)
  - [Concurrency & DDoS APIs](#concurrency--ddos-apis)
  - [System Metrics API](#system-metrics-api)

---

## For ALL APIs
All APIs are RESTful and most use JSON for request/response bodies unless noted as **[not JSON]**. Every response contains a `requested_at` field with an ISO 8601 timestamp (e.g., `2025-02-24T02:54:39.090Z`). For non-JSON APIs, this timestamp appears in plain text.

### Body Type
- **JSON:** All requests and responses use JSON except those explicitly marked as **[not JSON]**.
- **Timestamp:** Every response includes a `requested_at` field (ISO format with a 'T' delimiter).

### Optional Variables
- **Query Parameters:**  
  - Required: `?key=<value>`  
  - Optional: `?key=[value]`
- All query parameters, environment variables, and JSON body fields are "duck-typed". Biggie automatically converts values to the correct type; if conversion fails or illegal characters are detected, an error is returned.

### Random Variables
- Use the keyword `"RANDOM"` in any query parameter, environment variable, or JSON body field to select a random value per API request.
- When using `"RANDOM"`, the API response will include the chosen random value (either in a JSON field or as HTML text).
- For numeric fields, you can specify a range using the syntax: `RANDOM:<start>:<end>`.

### Standard Error Format
All JSON API errors follow this format:

```json
{
    "error": "UPPER_CASED_ERROR_TYPE",
    "message": "error reason in lower case",
    "request": {
        // Request details: header (with parsed cookie header), method, IP address, query parameters, and body information (length and payload).
    }
}
```

For non-JSON APIs, similar error information is returned in plain text.

### External Services
Some APIs require environment variables for external services. Variables are prioritized in the order listed; if a higher priority variable is provided, lower ones are ignored. Schemas and/or tables for testing are automatically created.

**Important:** All SSL/TLS certificates from external services are not verified on the Biggie side.

- **MySQL APIs:**  
  - `MYSQL_SECRET`, `AWS_REGION` (retrieves credentials from a secrets manager; format: `{"username":"a","password":"b","engine":"f","host":"c","port":"1","dbname":"d"}`)
  - `MYSQL_DBINFO` (credentials in JSON format; same format as above)  
  - Alternatively: `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USERNAME`, `MYSQL_PASSWORD`, `MYSQL_DBNAME`

- **PostgreSQL APIs:**  
  - `POSTGRES_SECRET`, `AWS_REGION` (retrieves credentials from a secrets manager; format: `{"username":"a","password":"b","engine":"f","host":"c","port":"1","dbname":"d"}`)
  - `POSTGRES_DBINFO` (credentials in JSON format; same format as above)  
  - Alternatively: `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USERNAME`, `POSTGRES_PASSWORD`, `POSTGRES_DBNAME`

- **Redshift APIs:**  
  - `REDSHIFT_SECRET`, `AWS_REGION` (retrieves credentials from a secrets manager; format: `{"username":"a","password":"b","engine":"f","host":"c","port":"1","dbname":"d"}`)
  - `REDSHIFT_DBINFO` (credentials in JSON format; same format as above)  
  - Alternatively: `REDSHIFT_HOST`, `REDSHIFT_PORT`, `REDSHIFT_USERNAME`, `REDSHIFT_PASSWORD`, `REDSHIFT_DBNAME`

- **Redis APIs:**  
  - `REDIS_HOST`, `REDIS_PORT`, `REDIS_TLS_ENABLED` (set to `true` or `false`)

- **Kafka APIs:**  
  - `KAFKA_SERVERS` (a comma-separated list of Kafka servers)
  - `KAFKA_TLS_ENABLED` (set to `true` or `false`)
  - `KAFKA_TOPIC`

---

## API Endpoints

### Basic APIs

#### Simple GET API
```
GET /simple
```
- Responds with the message `"ok"`.

#### Foo GET API
```
GET /simple/foo
```
- Responds with `"foo ok"`.
- The response includes request header information (with parsed cookies), method, IP address, and query parameters.

#### Bar POST API
```
POST /simple/bar
Content-Type: application/json

{ "any": "data" }
```
- Responds with `"bar ok"`.
- The response includes request header details (cookies parsed), method, IP address, query parameters, and body information (length and payload).

#### Random HTML API **[not JSON]**
```
GET /simple/color?color=[string]
```
- Returns an HTML file with a random background color.
- The color is initially selected when the application starts.
- The response includes request header information, method, IP address, and query parameters.
- You can override the default color using the `?color=` query parameter or the `RANDOM_HTML_API_COLOR` environment variable. Supports any CSS color format (e.g., hex code, rgb function).

#### Large Response API
```
GET /simple/large?length=<number>&sentence=[string]
```
- Generates a large JSON response by repeating a provided sentence or a random sentence.

---

### Health & Metadata APIs

#### Simple Health Check API
```
GET /healthcheck
```
- Returns `"ok"` as fast as possible.

#### Slow Health Check API
```
GET /healthcheck/slow?wait=[number]
```
- Waits for the number of seconds specified by `wait` (or a random duration) before returning `"ok"`.

#### Check External Service Health API
```
GET /healthcheck/external
```
- Tests the connection to all configured external services.

#### Fetch All Metadatas API
```
GET /metadata/all
```
- Retrieves metadata from EC2 Instance Metadata Service (v1 and v2), ECS Metadata Service, and EKS environment variables.

#### Visualize Revision HTML API **[not JSON]**
```
GET /metadata/revision_color
```
- Retrieves metadata from ECS Metadata Service, and EKS environment variables.
- Converts revision numbers (for EKS, the replicaSet from the pod name; for ECS, the task definition revision) to a CSS color string using a hash function.
- Displays different background colors based on revisions. The color is calculated at application startup.
- If ECS or EKS metadata is unavailable, displays a black background with an error message.

---

### Stress Test APIs

#### CPU Stress API
```
POST /stress/cpu
Content-Type: application/json

{ "cpu_percent": 30, "maintain_second": 30, "async": true }
```
- Maintains the specified `cpu_percent` for `maintain_second` seconds.
- If `async` is true, the API returns immediately while the stress test runs in the background.
- Memory usage is minimally affected.

#### Memory Stress API
```
POST /stress/memory
Content-Type: application/json

{ "memory_percent": 30, "maintain_second": 30, "async": true }
```
- Maintains the specified `memory_percent` for `maintain_second` seconds.
- If `async` is true, the API returns immediately while the stress test runs in the background.
- CPU usage is minimally affected.

#### Simulate Memory Leak API
```
POST /stress/memory_leak
Content-Type: application/json

{ "leak_size_mb": 50, "maintain_second": 30, "async": true }
```
- Gradually allocates a specified amount of memory (`leak_size_mb`) to simulate a memory leak, maintained for `maintain_second` seconds.
- Useful for testing the application's response to resource exhaustion.

#### Heavy File Write API
```
POST /stress/filesystem/write
Content-Type: application/json

{ "file_size": 1024, "file_count": 10, "maintain_second": 30, "async": true, "interval_second": 1 }
```
- Simulates heavy disk I/O by repeatedly writing multiple files (each of a specified size) for the duration defined by `maintain_second`.
- The `file_count` parameter controls how many files are written per interval.

#### Heavy File Read API
```
POST /stress/filesystem/read
Content-Type: application/json

{ "file_path": "/tmp/testfile", "maintain_second": 30, "async": true, "read_frequency": 10, "interval_second": 1 }
```
- Simulates high disk read load by repeatedly reading from a specified file.
- The `read_frequency` parameter determines how many read operations occur per interval.

#### Simulated Network Latency API
```
POST /stress/network/latency
Content-Type: application/json

{ "latency_ms": 200, "maintain_second": 30, "async": true }
```
- Introduces artificial latency in network communications by delaying responses by the specified number of milliseconds.
- Helps simulate slow or congested network conditions.

#### Simulated Packet Loss API
```
POST /stress/network/packet_loss
Content-Type: application/json

{ "loss_percentage": 20, "maintain_second": 30, "async": true }
```
- Simulates network instability by randomly dropping a percentage of packets during the test period.
- The `loss_percentage` parameter sets the drop rate.

---

### Heavy Database Activities

#### MySQL APIs

- **Heavy MySQL Query in Single Connection**
  ```
  POST /mysql/heavy
  Content-Type: application/json
  
  { "reads": true, "writes": true, "maintain_second": 30, "async": true, "query_per_interval": 10, "interval_second": 1 }
  ```
  - Performs heavy MySQL queries over a single connection with configurable read/write operations.
  - Queries run continuously for the duration specified by `maintain_second`.
  - Use `query_per_interval` and `interval_second` to control the query rate.
  - If `async` is enabled, the API returns immediately while processing in the background.

- **Heavy MySQL Query in Multiple Connections**
  ```
  POST /mysql/multi_heavy
  Content-Type: application/json
  
  { "reads": true, "writes": true, "maintain_second": 30, "async": true, "connection_counts": 10, "query_per_interval": 10, "interval_second": 1 }
  ```
  - Executes heavy MySQL queries across multiple concurrent connections.
  - Supports concurrent read/write operations with specified connection counts.
  - Maintains query load for `maintain_second` seconds with optional rate control.
  - Asynchronous mode returns immediately while running in the background.

- **Heavy MySQL Connections**
  ```
  POST /mysql/connection
  Content-Type: application/json
  
  { "maintain_second": 30, "async": true, "connection_counts": 100, "increase_per_interval": 10, "interval_second": 1 }
  ```
  - Simulates heavy connection loads by establishing multiple MySQL connections.
  - The total number is set by `connection_counts`, with connections ramping up based on `increase_per_interval` and `interval_second`.
  - Runs for `maintain_second` seconds and supports asynchronous execution.

#### PostgreSQL APIs

- **Heavy PostgreSQL Query in Single Connection**
  ```
  POST /postgres/heavy
  Content-Type: application/json
  
  { "reads": true, "writes": true, "maintain_second": 30, "async": true, "query_per_interval": 10, "interval_second": 1 }
  ```
  - Runs intensive PostgreSQL queries over a single connection with configurable read/write operations.
  - Queries execute continuously for `maintain_second` seconds.
  - Optional parameters control the query frequency.
  - Asynchronous mode returns immediately while the process runs in the background.

- **Heavy PostgreSQL Query in Multiple Connections**
  ```
  POST /postgres/multi_heavy
  Content-Type: application/json
  
  { "reads": true, "writes": true, "maintain_second": 30, "async": true, "connection_counts": 10, "query_per_interval": 10, "interval_second": 1 }
  ```
  - Executes heavy PostgreSQL queries using multiple connections.
  - Supports concurrent read/write operations with the number of connections specified by `connection_counts`.
  - The workload is maintained for `maintain_second` seconds with optional rate control.
  - Asynchronous execution enables background processing.

- **Heavy PostgreSQL Connections**
  ```
  POST /postgres/connection
  Content-Type: application/json
  
  { "maintain_second": 30, "async": true, "connection_counts": 100, "increase_per_interval": 10, "interval_second": 1 }
  ```
  - Simulates high connection loads for PostgreSQL by creating multiple connections.
  - The target is set by `connection_counts`, with a controlled ramp-up using `increase_per_interval` and `interval_second`.
  - Runs for `maintain_second` seconds and supports asynchronous processing.

#### Redshift APIs

- **Heavy Redshift Query in Single Connection**
  ```
  POST /redshift/heavy
  Content-Type: application/json
  
  { "reads": true, "writes": true, "maintain_second": 30, "async": true, "query_per_interval": 10, "interval_second": 1 }
  ```
  - Executes heavy Redshift queries over a single connection.
  - Supports both read and write operations.
  - Runs continuously for `maintain_second` seconds with rate control options.
  - Asynchronous mode returns an immediate response while processing continues in the background.

- **Heavy Redshift Query in Multiple Connections**
  ```
  POST /redshift/multi_heavy
  Content-Type: application/json
  
  { "reads": true, "writes": true, "maintain_second": 30, "async": true, "connection_counts": 10, "query_per_interval": 10, "interval_second": 1 }
  ```
  - Facilitates heavy query operations on Redshift using multiple connections.
  - The number of concurrent connections is defined by `connection_counts`, and queries run for `maintain_second` seconds with configurable frequency.
  - Supports asynchronous execution.

- **Heavy Redshift Connections**
  ```
  POST /redshift/connection
  Content-Type: application/json
  
  { "maintain_second": 30, "async": true, "connection_counts": 100, "increase_per_interval": 10, "interval_second": 1 }
  ```
  - Tests Redshift’s capacity by establishing many connections.
  - The total connections and ramp-up rate are controlled by `connection_counts`, `increase_per_interval`, and `interval_second`.
  - Runs for `maintain_second` seconds with asynchronous support.

#### Redis APIs

- **Heavy Redis Query in Single Connection**
  ```
  POST /redis/heavy
  Content-Type: application/json
  
  { "reads": true, "writes": true, "maintain_second": 30, "async": true, "query_per_interval": 10, "interval_second": 1 }
  ```
  - Triggers heavy Redis queries on a single connection, supporting both read and write operations.
  - Executes continuously for `maintain_second` seconds with optional query frequency control.
  - Asynchronous mode returns immediately while processing runs in the background.

- **Heavy Redis Query in Multiple Connections**
  ```
  POST /redis/multi_heavy
  Content-Type: application/json
  
  { "reads": true, "writes": true, "maintain_second": 30, "async": true, "connection_counts": 10, "query_per_interval": 10, "interval_second": 1 }
  ```
  - Executes heavy Redis queries across multiple connections.
  - The number of connections is set by `connection_counts`, and the query rate can be controlled with optional parameters.
  - Runs for `maintain_second` seconds with asynchronous processing.

- **Heavy Redis Connections**
  ```
  POST /redis/connection
  Content-Type: application/json
  
  { "maintain_second": 30, "async": true, "connection_counts": 100, "increase_per_interval": 10, "interval_second": 1 }
  ```
  - Simulates a high connection load for Redis by establishing multiple connections.
  - The total number is determined by `connection_counts`, and the ramp-up is controlled with `increase_per_interval` and `interval_second`.
  - Runs for `maintain_second` seconds with asynchronous execution.

#### Kafka APIs

- **Heavy Kafka Produce**
  ```
  POST /kafka/heavy
  Content-Type: application/json
  
  { "messages": "provided", "maintain_second": 30, "async": true, "produce_per_interval": 10, "interval_second": 1 }
  ```
  - Sends messages to a Kafka topic using a single producer connection.
  - Messages may be generated automatically or provided in the payload.
  - Production continues for `maintain_second` seconds with rate control via `produce_per_interval` and `interval_second`.
  - **Requires environment variables:**  
    - `KAFKA_SERVERS` (comma-separated list)  
    - `KAFKA_TLS_ENABLED` (true/false)  
    - `KAFKA_TOPIC`  
  - Asynchronous mode returns immediately while message production continues in the background.

- **Heavy Kafka Produce in Multiple Producers**
  ```
  POST /kafka/multi_heavy
  Content-Type: application/json
  
  { "messages": "provided", "maintain_second": 30, "async": true, "connection_counts": 10, "produce_per_interval": 10, "interval_second": 1 }
  ```
  - Performs heavy message production using multiple producer connections.
  - The number of producers is specified by `connection_counts`.
  - The process runs for `maintain_second` seconds with controlled message production frequency.
  - Requires the same Kafka environment variables as above.
  - Asynchronous mode returns promptly while processing continues in the background.

- **Heavy Kafka Connections**
  ```
  POST /kafka/connection
  Content-Type: application/json
  
  { "maintain_second": 30, "async": true, "connection_counts": 100, "increase_per_interval": 10, "interval_second": 1 }
  ```
  - Simulates a heavy load on Kafka by establishing many producer connections.
  - The total number is determined by `connection_counts`, with a controlled ramp-up using `increase_per_interval` and `interval_second`.
  - Requires the Kafka environment variables as described above.
  - Supports asynchronous execution.

---

### Error Injection APIs

#### Inject Random Error API
```
POST /stress/error_injection
Content-Type: application/json

{ "error_rate": 0.1, "maintain_second": 30, "async": true }
```
- Randomly injects errors into API responses at a defined rate (`error_rate`) to test application resilience and error handling under failure conditions.

#### Crash Simulation API
```
POST /stress/crash
Content-Type: application/json

{ "maintain_second": 10, "async": true }
```
- Simulates an unexpected service crash after a brief operational period.
- Useful for validating recovery procedures and failover mechanisms.

---

### Concurrency & DDoS APIs

#### Simulate Concurrent Flood
```
POST /stress/concurrent_flood
Content-Type: application/json

{ "target_endpoint": "/simple", "request_count": 1000, "maintain_second": 30, "async": true, "interval_second": 1 }
```
- Simulates a flood of concurrent requests to a specified target endpoint.
- The `request_count` parameter defines the number of requests generated per interval.
- The simulation runs for `maintain_second` seconds.
- With asynchronous mode enabled, the API returns immediately while the flood continues in the background.

#### Simulate Downtime
```
POST /stress/downtime
Content-Type: application/json

{ "downtime_second": 10, "async": true }
```
- Temporarily disables responses from Biggie for the duration specified by `downtime_second` to simulate service downtime.
- Useful for testing system resilience, failover mechanisms, and monitoring alerts.
- Asynchronous mode returns immediately while the downtime simulation is in progress.

#### Simulate External API Calls
```
POST /stress/third_party
Content-Type: application/json

{ "target_url": "https://api.example.com/data", "maintain_second": 30, "async": true, "call_rate": 10, "interval_second": 1, "simulate_errors": true }
```
- Simulates integration with a third-party API by sending continuous requests to the specified `target_url`.
- Operates for `maintain_second` seconds with a call frequency defined by `call_rate` and `interval_second`.
- When `simulate_errors` is enabled, random errors are injected into some calls to mimic an unstable external service.

#### Simulate DDoS Attack
```
POST /stress/ddos
Content-Type: application/json

{ "target_endpoint": "/simple", "attack_intensity": 1000, "maintain_second": 30, "async": true, "interval_second": 1 }
```
- Simulates a DDoS attack by generating a high volume of requests towards a specified target endpoint.
- The `attack_intensity` parameter defines the number of requests per interval.
- The simulation runs for the duration specified by `maintain_second`.
- With asynchronous mode enabled, the API returns immediately while the attack is executed in the background.

---

### System Metrics API

#### Fetch System Metrics
```
GET /metrics/system
```
- Provides aggregated system metrics such as CPU load, memory usage, network throughput, and details of ongoing stress tests.
- Useful for monitoring the overall performance and health of Biggie during various stress scenarios.
