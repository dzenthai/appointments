## **Quick start**

**Clone the repository**

   ```bash
   git clone https://github.com/dzenthai/appointments.git
   cd appointments
   ```

**Set up the environment for the API**
1. Copy the environment
   ```bash
   cp .envrc.example .envrc
   ```
2. Replace the placeholders with your own API key and sender
   ```
   export RESEND_API_KEY='api_key'
   export RESEND_SENDER='sender'
   ```
3. Load the environment automatically
   ```bash 
   direnv allow
   ```

**Docker compose**
   ```bash
   docker compose up -d 
   ```
   or run it using "make"
   ```bash
   make docker/compose
   ```

**Healthcheck**
   ```bash 
   curl localhost:4000/v1/healthcheck
   ```

---

## **Architecture**

### Layer responsibilities

**Middleware**

- panic recovery;
- request logging;
- rate limiting;
- authentication;
- CORS;

**Handler**

- writing JSON;
- reading JSON;
- input data validation;
- calling domain logic;
- converting an error into an HTTP response;

**Domain**

- business logic;
- constructors;
- models;

**Infrastructure**

- PostgreSQL;
- Resend (outbound email API);
- configuration;
- logging.

### Middleware chain architecture
1. recoverPanic - ensure the HTTP pipeline is fault-tolerant.
2. logRequest - collect information on each HTTP request.
3. rateLimiter - maximize the early rejection of excessive or harmful traffic.
4. enableCORS - ensure that the HTTP API interacts correctly with the security policies of browsers.
5. authenticate - establishing the identity of the user.

---

## **Solutions Logbook**

1. A "404 Not Found" response is returned instead of
   "403 Forbidden" to prevent resource enumeration.
2. A "400 Bad Request" response is used instead of
   "422 Unprocessable Content" to simplify client error handling and maintain consistent error responses.
3. Using interfaces in handlers to facilitate testing (interfaces are defined close to the consumer).
4. Using different byte lengths for tokens, the verify token is short-term and the auth token is long-term.
5. Checking access based on user roles instead of using a permissions table.
6. Separating audit and tidy Makefile targets.
7. Running make audit as part of the CI pipeline.