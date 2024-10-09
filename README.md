# TaskTracker

## Summary

TaskTracker is a RESTful API that serves as the backend for a project management platform where users can create accounts, manage teams and assign tasks. It provides a robust role-based permission system for managing team members and tasks, and supports token-based authentication with password resets.

## Features

* User registration, email verification and password reset

* Team role-based access control

    There are 4 roles available, each having the permissions of the previous one. Only the owner can change member roles.

    - Regular: no special permissions

    - Leader: can assign tasks to other team members

    - Admin: can invite or remove team members

    - Owner: can edit team settings and manage member roles

* Task status

    - Open: by default tasks are created with open status

    - In-Progress: only open tasks can be marked as in-progress by the task assignee

    - Completed: only in-progress tasks can be makred as completed by the task creator

    - Cancelled: only in-progress tasks can be marked as cancelled by the task creator

* Task priority

    - Low

    - Medium

    - High

## API Documentation

You can explore the TaskTracker API using [this](https://elements.getpostman.com/redirect?entityId=38661095-56216de5-2baf-460a-a86f-8878121e4b12&entityType=collection) Postman collection. It includes all the available endpoints for user registration, team and task management.

## Running Locally

### Prerequisites

Ensure you have the following installed:

* Go 1.22: [Download Go](https://go.dev/dl/)

* PostgreSQL 16: [Download PostgreSQL](https://www.postgresql.org/download/)

* Go Migrate Tool: [Install Migrate](https://github.com/golang-migrate/migrate)

You will also need a PostgreSQL database with the citext extension enabled for case-insensitive email comparison:

```sql
CREATE EXTENSION IF NOT EXISTS citext;
```

### Installation

**1. Clone the repository**

```bash
git clone https://github.com/yourusername/tasktracker.git
cd tasktracker
```

**2. Run the migrations**

The Go Migrate tool must be installed locally in the project directory to run migrations.

```bash
migrate -path ./migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" up
```

### Running the Application

The application is located in the /cmd/api directory, and you must ensure the proper configurations (such as database connection, SMTP settings, etc.) are set via environment variables or command-line flags.

### Configuration

The application can be configured via command-line flags or environment variables. Command-line flags always take precedence over environment variables. If neither is provided or if the environment variable cannot be parsed, the application will fall back to the default value, if available.

For configurations that do not have default values (e.g., `pg-dsn`, `smtp-host`, etc.), the user must explicitly specify them either via a flag or an environment variable. Otherwise, the application will not function properly.

| Flag                     | Environment Variable      | Default Value         | Description                                                 |
|--------------------------|---------------------------|-----------------------|-------------------------------------------------------------|
| `-port`                  | `PORT`                    | `8080`                | Port to run the API server.                                  |
| `-environment`           | `ENVIRONMENT`             | `development`         | Set the environment (`development`, `staging`, `production`).|
| `-pg-dsn`                | `PG_DSN`                  | *None*                | PostgreSQL connection string (DSN). **Required**             |
| `-pg-max-open-conns`     | `PG_MAX_OPEN_CONNS`       | `25`                  | Maximum open PostgreSQL connections.                         |
| `-pg-max-idle-conns`     | `PG_MAX_IDLE_CONNS`       | `25`                  | Maximum idle PostgreSQL connections.                         |
| `-pg-conn-max-idle-time` | `PG_CONN_MAX_IDLE_TIME`   | `15m`                 | Maximum idle time for PostgreSQL connections.                |
| `-smtp-host`             | `SMTP_HOST`               | *None*                | SMTP host for sending emails. **Required**                   |
| `-smtp-port`             | `SMTP_PORT`               | `2525`                | SMTP port.                                                   |
| `-smtp-username`         | `SMTP_USERNAME`           | *None*                | SMTP username. **Required**                                  |
| `-smtp-password`         | `SMTP_PASSWORD`           | *None*                | SMTP password. **Required**                                  |
| `-smtp-sender`           | `SMTP_SENDER`             | *None*                | Email sender address for outgoing emails (e.g., password resets). **Required** |
| `-limiter-rps`           | `LIMITER_RPS`             | `2`                   | Rate limiter requests per second.                            |
| `-limiter-burst`         | `LIMITER_BURST`           | `4`                   | Rate limiter burst.                                          |
| `-limiter-enabled`       | `LIMITER_ENABLED`         | `true`                | Enable or disable the rate limiter.                          |
| `-cors-trusted-origins`  | `CORS_TRUSTED_ORIGINS`    | *None*                | Comma-separated list of trusted CORS origins.                |
