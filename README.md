# GoRestExample API

A production-ready REST API built with Go, featuring JWT authentication and OAuth2.0 integration.

## Features

- **JWT Authentication**: Secure token-based authentication with access and refresh tokens
- **OAuth2.0 Integration**: Support for Google, GitHub, Microsoft, and Facebook login
- **Role-Based Access Control**: Admin and user roles with middleware-based authorization
- **User Management**: Complete user CRUD operations with proper authorization
- **Clean Architecture**: Well-structured codebase following Go best practices
- **PostgreSQL Database**: Using GORM ORM with automatic migrations
- **Environment Configuration**: Secure configuration management with environment variables

## Tech Stack

- **Web Framework**: Gin (v1.9.1)
- **Database ORM**: GORM (v1.25.5)
- **Database**: PostgreSQL
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **OAuth2.0**: golang.org/x/oauth2
- **Password Hashing**: bcrypt

## Project Structure

```
GoRestExample/
├── config/          # Configuration management
├── handlers/        # HTTP request handlers
├── middleware/      # Authentication and authorization middleware
├── models/          # Database models and schema
├── utils/           # Utility functions (JWT, password hashing)
├── main.go          # Application entry point
├── .env.example     # Environment variables template
└── README.md        # This file
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- (Optional) OAuth2.0 credentials from providers

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd GoRestExample
```

2. Copy the environment template and configure:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Install dependencies:
```bash
go mod download
```

4. Set up PostgreSQL database and update `.env` with your credentials

5. Run the application:
```bash
go run main.go
```

The server will start on `http://localhost:8080` (or the port specified in `.env`)

## Configuration

### Environment Variables

Create a `.env` file based on `.env.example`:

```bash
# Server
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_USER=gorm
DB_PASSWORD=gorm
DB_NAME=gorm
DB_PORT=9920

# JWT (CHANGE THESE IN PRODUCTION!)
JWT_SECRET=your-super-secret-jwt-key
JWT_REFRESH_SECRET=your-super-secret-refresh-key
JWT_ACCESS_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_HOURS=168

# OAuth2.0 (Optional)
OAUTH_REDIRECT_BASE_URL=http://localhost:8080
GOOGLE_OAUTH_ENABLED=true
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
# ... (similar for GitHub, Microsoft, Facebook)
```

### OAuth2.0 Setup

To enable OAuth2.0 providers:

1. **Google**: https://console.developers.google.com/
2. **GitHub**: https://github.com/settings/developers
3. **Microsoft**: https://portal.azure.com/
4. **Facebook**: https://developers.facebook.com/

Set the redirect URL to: `{OAUTH_REDIRECT_BASE_URL}/auth/oauth/{provider}/callback`

## API Endpoints

### Authentication

#### Register (Username/Password)
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user",
    "oauth_provider": "local"
  }
}
```

#### Login (Username/Password)
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

#### OAuth2.0 Login
```http
GET /api/v1/auth/oauth/{provider}/login
```
Supported providers: `google`, `github`, `microsoft`, `facebook`

**Response:**
```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/auth?...",
  "state": "google-1234567890"
}
```

Redirect user to `auth_url`, then handle callback:

```http
GET /api/v1/auth/oauth/{provider}/callback?code=...&state=...
```

#### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}
```

#### Logout
```http
POST /api/v1/auth/logout
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}
```

### Protected Endpoints

All protected endpoints require the `Authorization` header:
```
Authorization: Bearer {access_token}
```

#### Get Current User
```http
GET /api/v1/me
```

#### List All Users (Authenticated users)
```http
GET /api/v1/users
```

#### Get User by ID
```http
GET /api/v1/users/{id}
```

#### Update User
```http
PUT /api/v1/users/{id}
Content-Type: application/json

{
  "name": "Jane Doe",
  "avatar": "https://example.com/avatar.jpg"
}
```
Users can only update their own profile unless they are admins.

#### Delete User (Admin only)
```http
DELETE /api/v1/users/{id}
```

#### Update User Role (Admin only)
```http
PATCH /api/v1/users/{id}/role
Content-Type: application/json

{
  "role": "admin"
}
```
Valid roles: `user`, `admin`

### Health Check
```http
GET /health
```

## Authentication Flow

### 1. Username/Password Flow
1. User registers or logs in with email/password
2. Server validates credentials and returns access + refresh tokens
3. Client stores tokens and includes access token in subsequent requests
4. When access token expires, use refresh token to get new access token

### 2. OAuth2.0 Flow
1. Client requests OAuth login URL from `/auth/oauth/{provider}/login`
2. Client redirects user to the returned `auth_url`
3. User authenticates with OAuth provider
4. Provider redirects back to `/auth/oauth/{provider}/callback`
5. Server exchanges code for user info and returns JWT tokens
6. Client uses JWT tokens for subsequent requests

## Security Best Practices

1. **Change JWT Secrets**: Always use strong, unique secrets in production
2. **Use HTTPS**: Enable TLS in production
3. **Environment Variables**: Never commit `.env` file with secrets
4. **Password Requirements**: Enforce minimum 6 characters (increase in production)
5. **Token Expiry**: Access tokens expire in 24 hours, refresh tokens in 7 days
6. **CORS**: Configure CORS middleware for production

## Role-Based Access Control

### Roles
- **user**: Default role for registered users
  - Can view users
  - Can update own profile

- **admin**: Administrative role
  - All user permissions
  - Can delete any user
  - Can update any user's role

### Creating an Admin User

After registering a user, manually update the role in the database:
```sql
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
```

Or use the admin endpoint once you have at least one admin user.

## Database Schema

### Users Table
- `id` (uint, primary key)
- `name` (string)
- `email` (string, unique)
- `password` (string, hashed)
- `role` (enum: user, admin)
- `oauth_provider` (enum: local, google, github, microsoft, facebook)
- `oauth_id` (string)
- `avatar` (string)
- `verified` (boolean)
- `created_at`, `updated_at`, `deleted_at` (timestamps)

### Refresh Tokens Table
- `id` (uint, primary key)
- `user_id` (uint, foreign key)
- `token` (string, unique)
- `expires_at` (timestamp)
- `created_at`, `deleted_at` (timestamps)

## Development

### Building
```bash
go build -o gorestexample
./gorestexample
```

### Running Tests
```bash
go test ./...
```

## Production Deployment

1. Set `GIN_MODE=release` environment variable
2. Use strong JWT secrets
3. Enable HTTPS/TLS
4. Configure CORS properly
5. Set up database connection pooling
6. Use a process manager (systemd, supervisor)
7. Set up logging and monitoring
8. Regular security updates

## Error Responses

All errors follow this format:
```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict (e.g., email already exists)
- `500` - Internal Server Error

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is open source and available under the MIT License.

## Support

For issues and questions, please open an issue on GitHub.
