# The Biggie

The BIG application for exercising HA and DR.  
Written in Golang.  

## Features
* Simulate slow startup, network delays, and packet drops.
* Make application crash and restarts randomly.
* Perform CPU, memory, and memory leak stress tests.
* Generate high disk I/O loads with file read/write operations.
* Flood the system with concurrent requests and simulate DDoS attacks.
* Inject random errors and simulate service crashes for HA/DR testing.
* Retrieve detailed system metrics including CPU load, memory usage, and network throughput.
* Stress-test RDBMS (MySQL, PostgreSQL, Redshift) with massive queries and connection flooding.
* Challenge Redis and Kafka systems with high-volume operations and producer connection loads.
* Customize logging output with environment. or make it very RANDOM
* Consistent JSON API responses and standardized error handling across all endpoints.

---

## Table of Contents

- [The Biggie](#the-biggie)
  - [Features](#features)
  - [Table of Contents](#table-of-contents)
  - [For ALL APIs](#for-all-apis)
    - [Body Type](#body-type)
    - [Optional Variables](#optional-variables)
    - [Random Variables](#random-variables)
    - [Standard Error Format](#standard-error-format)
    - [External Services](#external-services)
    - [LOG\_FORMAT Environment Variable](#log_format-environment-variable)
      - [Predefined Formats](#predefined-formats)
      - [Custom Formats](#custom-formats)
      - [RANDOM Format](#random-format)
      - [Examples](#examples)
    - [STARTUP\_DELAY\_SECOND Environment Variable](#startup_delay_second-environment-variable)
  - [API Endpoints](#api-endpoints)
    - [Basic APIs](#basic-apis)
      - [Simple GET API](#simple-get-api)
      - [Foo GET API](#foo-get-api)
      - [Bar POST API](#bar-post-api)
      - [Random HTML API **\[not JSON\]**](#random-html-api-not-json)
      - [Large Response API](#large-response-api)
    - [Health \& Metadata APIs](#health--metadata-apis)
      - [Simple Health Check API](#simple-health-check-api)
      - [Slow Health Check API](#slow-health-check-api)
      - [Check External Service Health API](#check-external-service-health-api)
      - [Run HTTP request](#run-http-request)
      - [Fetch All Metadatas API](#fetch-all-metadatas-api)
      - [Visualize Revision HTML API **\[not JSON\]**](#visualize-revision-html-api-not-json)
    - [Stress Test APIs](#stress-test-apis)
      - [CPU Stress API](#cpu-stress-api)
      - [Memory Stress API](#memory-stress-api)
      - [Simulate Memory Leak API](#simulate-memory-leak-api)
      - [Heavy File Write API](#heavy-file-write-api)
      - [Heavy File Read API](#heavy-file-read-api)
      - [Simulated Network Latency API](#simulated-network-latency-api)
      - [Simulated Packet Loss API](#simulated-packet-loss-api)
    - [Heavy Database Activities](#heavy-database-activities)
      - [MySQL APIs](#mysql-apis)
      - [PostgreSQL APIs](#postgresql-apis)
      - [Redshift APIs](#redshift-apis)
      - [Redis APIs](#redis-apis)
      - [Kafka APIs](#kafka-apis)
    - [Error Injection APIs](#error-injection-apis)
      - [Inject Random Error API](#inject-random-error-api)
      - [Crash Simulation API](#crash-simulation-api)
    - [Concurrency \& DDoS APIs](#concurrency--ddos-apis)
      - [Simulate Concurrent Flood](#simulate-concurrent-flood)
      - [Simulate Downtime](#simulate-downtime)
      - [Simulate External API Calls](#simulate-external-api-calls)
      - [Simulate DDoS Attack](#simulate-ddos-attack)
    - [System Metrics API](#system-metrics-api)
      - [Fetch System Metrics](#fetch-system-metrics)
    - [Fake Log Generation API](#fake-log-generation-api)
      - [Generate Logs](#generate-logs)

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

### LOG_FORMAT Environment Variable

The `LOG_FORMAT` environment variable controls the log output format for the application. It accepts either predefined format names, a custom format string with placeholders, or a special value `"RANDOM"` which instructs the system to generate a random log format according to a defined algorithm.

#### Predefined Formats
- **apache**: Uses the Apache common log format.
  - Example:  
    `{client_ip} - - {time:%d/%m/%Y:%H:%M:%S} {method} {path} {status_code} -`
- **nginx**: Uses a common Nginx log format.
  - Example:  
    `{client_ip} - {time:%d/%b/%Y:%H:%M:%S} {method} {path} {status_code} {latency:ms}`
- **full**: Outputs logs in common format, including all supported placeholders.
  - Example:  
    `{time} {status_code} {method} {path} {client_ip} {latency} "{user_agent}" {protocol} {request_size} {response_size}`

#### Custom Formats
> We serve simple webpage that generates random Log formats\
> https://biggie-logformat.pmh.codes

You can specify a custom log format string using placeholders. Placeholders are defined using curly braces in the following forms:
- **Basic placeholder**:  
  `{placeholder_name}`
- **Placeholder with unit**:  
  `{placeholder_name:unit}`

Supported placeholder names (case-insensitive):
- **time** – The current timestamp.  
  *Optionally*, you can specify a strftime(3)-like format (e.g., `{time:%Y-%m-%dT%H:%M:%S}`).
- **status_code** – The HTTP response status code.
- **method** – The HTTP request method.
- **path** – The requested URL path.
- **client_ip** – The client's IP address.
- **latency** – The time taken to process the request.  
  Supported units: `s`, `ms`, `mcs` (microseconds), `ns` (nanoseconds).  
  If omitted, a human-readable value with unit label is provided (e.g., `10.001s`).
- **user_agent** – The User-Agent header value.
- **protocol** – The HTTP protocol (e.g., `HTTP/1.1`, `HTTP/2`).
- **request_size** – The size of the HTTP request body.  
  Supported units: `b` (bytes), `kb`, `mb`, `gb` (using 1024-based conversion).  
  If omitted, a human-readable value with unit label is provided (e.g., `10.001kb`).
- **response_size** – The size of the HTTP response body.  
  Same unit rules as for request_size.

#### RANDOM Format
If you set `LOG_FORMAT` to `"RANDOM"`, the application will generate a random log format at startup according to these rules:
- **Mandatory Fields**:  
  The generated format will always include these required placeholders:  
  `time`, `status_code`, `method`, `path`, and `client_ip`.
- **Optional Fields**:  
  Up to 2 optional placeholders (from `latency`, `user_agent`, `protocol`, `request_size`, `response_size`) may be added, with a total of no more than 7 placeholders.
- **Random Ordering**:  
  The placeholders are randomly ordered.
- **Unit Specifiers**:  
  For placeholders that support units (such as `time`, `latency`, `request_size`, and `response_size`), a random unit specifier is chosen.  
  For example, the `time` placeholder might be assigned a random strftime format (ensuring it includes year, month, day, hour, minute, and second).
- **Random Quoting and Delimiters**:  
  Each placeholder is independently wrapped (or not) with random quotes (`" "`, `' '`) or square brackets (`[ ]`).  
  Additionally, random " - " tokens may be inserted between placeholders.
- **Consistency**:  
  The generated random format is created once at application startup and used consistently for all logging during that run.
- **Display**:  
  The generated random format is printed at startup so you know exactly which format is being used.

#### Examples
- **Predefined Format (apache)**:  
  `{client_ip} - - {time:%d/%m/%Y:%H:%M:%S} {method} {path} {status_code} -`
- **Predefined Format (nginx)**:  
  `{client_ip} - {time:%d/%b/%Y:%H:%M:%S} {method} {path} {status_code} {latency:ms}`
- **Predefined Format (json)**:  
  `{time} {status_code} {method} {path} {client_ip} {latency} {user_agent} {protocol} {request_size} {response_size}`
- **RANDOM Format (example)**:  
  `"[{time:%Y/%m/%dT%H:%M:%S}]" - {client_ip} {method} "{path}" {status_code} {latency:ms} - {user_agent}`

Each time the application starts with `LOG_FORMAT` set to `"RANDOM"`, a new random log format is generated following the rules above.

This feature gives you flexible control over your log output, allowing you to use standard log formats, customize the output, or experiment with randomly generated log formats.

### STARTUP_DELAY_SECOND Environment Variable

The `STARTUP_DELAY_SECOND` environment variable allows you to introduce an intentional delay at application startup. This is useful for simulating service initialization delays, orchestrating startup order among dependent services, or testing how your application behaves when there is a delay before it starts handling requests.

**Example Usage:**
- Fixed delay:  
  `STARTUP_DELAY_SECOND=3`
- Random delay within a range:  
  `STARTUP_DELAY_SECOND=RANDOM:1:5`

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

#### Run HTTP request
```
POST /healthcheck/hops
Content-Type: application/json

{
  "url": "http://a.com/simple/bar",
  "method": "POST",
  "headers": {
     "Content-Type": "application"
  },
  "body": "{\"hello\":\"world\"}"
}
```
- For simulate micro service architecture this API calls another APIs for you

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

---

### Fake Log Generation API

#### Generate Logs
\`\`\`
POST /stress/logs
Content-Type: application/json

{ "maintain_second": 30, "log_count_per_interval": "RANDOM:5:15", "line_per_log": 3, "interval_seconds": 1, "async": true }
\`\`\`

- Generates log messages over time with random content.
- The `maintain_second` parameter defines the total duration (in seconds) during which logs are generated.
- The `log_count_per_interval` parameter specifies the number of log messages to generate per interval. This field supports the RANDOM syntax (e.g., `"RANDOM:5:15"`).
- The `line_per_log` parameter indicates the number of lines in each generated log message.
- The `interval_seconds` parameter defines the time interval (in seconds) between each log generation cycle.
- If `async` is true, the API returns immediately while log generation continues in the background.
- Each log message is generated using random values for common placeholders (such as time, status code, method, path, client IP, latency, and cookies) according to the current LOG_FORMAT configuration.
