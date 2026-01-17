Based on the provided PDF, here is the Engineering Requirements Document (ERD) for the Distributed URL Shortener.

---

# Engineering Requirements Document: Distributed URL Shortener

## 1. Introduction

The objective is to design and implement a URL shortening service that demonstrates the ability to balance architectural patterns, operational safety, and infrastructure design.

## 2. Functional Requirements

### 2.1 URL Creation

**Endpoint:** `POST /` (Implied)

* **Inputs:**
* 
`long_url` (Required): The original URL to be shortened.


* `ttl_seconds` (Optional): Time-to-live for the record. Default is 24 hours.




* **Validation:**
* Verify `long_url` is a valid HTTP/HTTPS URL.


* Enforce reasonable length constraints on inputs to mitigate abuse.




* **Processing:**
* Generate a unique short code.


* 
**Constraint:** Short codes must exclude `0`, `O`, `l`, and `1` to ensure readability.


* Persist the mapping of short code to long URL.



### 2.2 Redirection

**Endpoint:** `GET /s/{short_code}` 

* **Behavior:**
* 
**Active:** If the code exists and is not expired, return `302 Found` and redirect to the `long_url`.


* 
**Inactive:** If the code is nonexistent or expired, return `404 Not Found`.




* **Side Effects:**
* Update `click_count` and `last_accessed_at` in a thread-safe manner.





### 2.3 Observability Endpoint

**Endpoint:** `GET /stats/{short_code}` 

* **Output:** Return a JSON object containing:
* `long_url`
* `created_at`
* `expires_at`
* `click_count`
* 
`last_accessed_at`.





### 2.4 Expiration Management

* Implement a strategy (Lazy, Background, or Hybrid) to clean up expired records.


* Document the chosen mechanism and operational tradeoffs.



## 3. Non-Functional Requirements

### 3.1 Observability & Instrumentation

* 
**Custom Header:** All HTTP responses (Success, Redirect, Error) must include `X-Processing-Time-Micros` indicating internal execution duration in microseconds.



### 3.2 Privacy

* 
**PII Constraint:** Requester IP addresses must **not** be stored in the database or exposed in system logs.



### 3.3 Scalability & Performance

* 
**Readiness:** The architecture must be capable of scaling (via documentation plan) to 10,000 requests per second for redirects.


* 
**Capacity:** The system design should theoretically accommodate 100 million new URLs per month.



## 4. Technical Constraints & Architecture

### 4.1 Storage Abstraction

* 
**In-Memory Allowed:** The implementation may use in-memory storage for this exercise.


* **Interface Requirement:** Access to storage must be mediated through a strict abstraction layer (Interface or Abstract Class). This allows swapping the engine (e.g., to DynamoDB/Redis) without changing business logic.



### 4.2 Concurrency

* Use idiomatic concurrency primitives to prevent data races, specifically for `click_count` updates.



### 4.3 Containerization

* 
**Docker:** Provide a multi-stage Dockerfile.


* 
**Security:** The container must run as a non-root user.



### 4.4 Infrastructure as Code (IaC)

* 
**Tool:** Terraform (`main.tf`).


* 
**Platform:** AWS or GCP.


* **Resources:**
* Serverless compute resource.


* Managed storage resource.


* IAM roles following Least Privilege.





## 5. Quality Assurance & Delivery

### 5.1 Static Analysis (CI/CD)

A GitHub Actions pipeline must enforce the following:

* 
**Cyclomatic Complexity:** Max 10 per function.


* 
**Go:** `golangci-lint` (enable `gocyclo`, `revive`).





### 5.2 Testing Strategy

* 
**Concurrency Test:** Prove `click_count` safety under 100+ concurrent requests.


* 
**Deterministic Expiration:** Verify TTL logic using a mocked system clock (no `sleep` allowed).


* 
**Interface Verification:** Tests must target the storage abstraction interface, not the concrete implementation.



### 5.3 Documentation

* 
**Architecture Diagram:** Visual representation of data model and system.


* 
**Gap Analysis:** Why in-memory fails in the defined cloud env and how managed storage fixes it.


* 
**Capacity Planning:** Storage estimates for 12 months.


* 
**SLIs/SLOs:** Define two indicators and objectives, plus one on-call scenario.



### 5.4 Delivery Format

* Git repository with incremental commit history (no single-commit dumps).



---

**Would you like to start by defining the Storage Interface abstraction to satisfy the architectural requirements?**