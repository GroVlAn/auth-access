# Access service

Authorization and access control service for the Auth Platform.

The service is responsible for managing roles and permissions and determining which operations are available to a user.

## Overview

auth-access is a backend microservice responsible for authorization and Role-Based Access Control (RBAC).

The service manages the relationship between users, roles, and permissions.

It does not authenticate users and does not issue JWT tokens.

## Responsibilities

The service is responsible for:

- Role management
- Permission management
- Assigning roles to users
- Assigning permissions to roles
- Retrieving user roles
- Retrieving role permissions
- Retrieving all permissions available to a user
- Providing authorization information to the API Gateway

## Repository Structure

```text
auth-access/
├── cmd/
│   └── main.go                     # Application entry point
├── configs/
│   ├── config-local.yml            # Local development configuration
│   └── config-prod.yml             # Production configuration
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration parsing and loading
│   ├── domain/
│   │   ├── access.go               # Domain models (Role, Permission, etc.)
│   │   ├── request.go              # Request DTOs
│   │   └── response.go             # Response DTOs
│   ├── handler/
│   │   ├── http-handler/
│   │   │   ├── http-handler.go     # HTTP endpoints (if any)
│   │   │   └── tools.go            # HTTP utilities
│   │   └── grpc-handler/
│   │       ├── grpc-handle.go      # gRPC server registration
│   │       └── access.go           # gRPC service implementation
│   ├── infrastructure/
│   │   ├── database/
│   │   │   └── postgresql.go       # PostgreSQL connection and queries
│   │   ├── preloader/
│   │   │   └── preloader.go        # Seeds default roles/permissions on startup
│   │   ├── secrets/
│   │   │   ├── provider.go         # Secret provider interface
│   │   │   └── secrets.go          # Vault-based secret retrieval
│   │   └── vault-client/
│   │       └── vault.go            # Vault client wrapper
│   ├── repository/
│   │   ├── repository.go           # Data access layer (CRUD for roles/permissions)
│   │   └── tools.go                # Repository helpers
│   ├── server/
│   │   ├── grpc-server/
│   │   │   └── grpc-server.go      # gRPC server setup
│   │   └── http-server/
│   │       └── http-server.go      # HTTP server setup
│   └── service/
│       ├── service.go              # Business logic (RBAC operations)
│       ├── service_test.go         # Unit tests
│       └── mocks_test.go           # Mock implementations for testing
├── migrations/
│   ├── 001_create_access_table.up.sql   # Initial schema
│   └── 001_create_access_table.down.sql # Rollback schema
├── role-config.json                # Default roles and permissions (seeded on first run)
├── .mockery.yml                    # Mockery configuration for generating mocks
├── Dockerfile                      # Docker build instructions
├── go.mod
└── go.sum
```

## Configuration

The service uses a YAML configuration file and environment variables. Two presets are provided:

- `configs/config-local.yml` – for local development (HTTP `9083`, gRPC `9013`).
- `configs/config-prod.yml` – for production (HTTP `8083`, gRPC `8013`).

You can copy and edit the appropriate file. The configuration path can be passed via the `-config` flag or the `CONFIG_PATH` environment variable.

### Environment Variables

The following variables **must** be set:

| Variable             | Description                                                      |
| -------------------- | ---------------------------------------------------------------- |
| `VAULT_SECRET_TOKEN` | Vault authentication token.                                      |
| `VAULT_ADDRESS`      | Vault server address (e.g., `http://vault:8200`).                |
| `VAULT_MOUNT`        | Vault mount path where secrets are stored (e.g., `secret`).      |
| `POSTGRES_PATH`      | Vault path for PostgreSQL credentials (e.g., `access/postgres`). |

Database credentials (username, password, database name) are fetched from Vault using the `POSTGRES_PATH`. The service expects the following keys under that path: `host`, `port`, `username`, `password`, `dbname`, `sslmode`.

Example configuration structure in Vault:

```json
{
	"host": "access_db",
	"port": "5435",
	"username": "admin",
	"password": "securepassword",
	"dbname": "auth_access",
	"sslmode": "disable"
}
```

## Initial Roles and Permissions

On first startup, the service seeds the database with default roles and permissions defined in `role-config.json`.

The file contains an array of roles, each with a name, description, `is_default` flag, and a list of permissions:

```json
[
  {
    "name": "admin",
    "is_default": false,
    "description": "administrator",
    "permissions": [
      {"name": "watch", "description": "can only watch content"},
      {"name": "create", "description": "can create new records"},
      ...
    ]
  }
]
```

## Running the Service

### Local Development

Build and run:

```bash
go run cmd/main.go -config=configs/config-local.yml
```

## Database Migrations

Schema changes are managed using [golang-migrate](https://github.com/golang-migrate/migrate). Migrations are stored in `migrations/` and are applied automatically on startup via the `migration_access` container (see `docker-compose.yaml`).

To run migrations manually:

```bash
migrate -path ./migrations -database "postgres://user:pass@host:5432/auth_access?sslmode=disable" up
```

## Authorization Model

The service uses Role-Based Access Control (RBAC).

The basic relationship is

![Relationship](architecture/relationship.png)

A user does not need to be assigned individual permissions directly.

Instead:

A user is assigned one or more roles.
A role is assigned one or more permissions.
The user's effective permissions are calculated from their roles.

For example:

user -> - user.watch - user.create - user.update

## Permission Model

Permissions represent individual operations that can be performed within the platform.

A permission follows the following naming convention:

resource.action

For example:

user.read
user.write
user.delete

role.read
role.write
role.delete

This naming convention makes permissions explicit and easy to evaluate.

## User Permissions

The service can calculate the complete set of permissions available to a user.

For example:

User: 123

Roles:

- admin
- user

Permissions:

- user.watch
- user.create
- user.delete
- admin.watch
- admin.create

The resulting permission set is returned to the API Gateway, which uses it to determine whether a request should be allowed.

## API Gateway Integration

**auth-access** is primarily consumed by the API Gateway.

The Gateway determines which permission is required for an incoming request and requests the permissions available to the authenticated user.

For example:

GET /users

Required permission:

user.watch

The Gateway then evaluates:

user.watch ∈ userPermissions

If the permission exists, the request is forwarded.

Otherwise, the Gateway returns:

403 Forbidden

## Security

The service follows several security principles:

- Authorization data is isolated from authentication data.
- The database is accessed only by auth-access.
- Other services communicate through gRPC.
- Secrets are managed through HashiCorp Vault.
- Permissions are explicitly defined.
- Access decisions are based on the authenticated user's identity.
- Database queries return distinct effective permissions for a user.

## License

This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.

You are free to use, modify, distribute, and sublicense the code for both commercial and non‑commercial purposes, provided that the original copyright notice and permission notice are included in all copies or substantial portions of the software.

For more information, see the full [MIT License](https://opensource.org/licenses/MIT).

```

```
