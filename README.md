# Snippetbox

Snippetbox is a server-rendered web application for publishing and viewing text snippets. It is written in Go and stores snippets, user accounts, and sessions in MySQL.

## Features

- Lists the ten most recently created snippets that have not expired.
- Displays an individual snippet until its expiration time.
- Supports account signup, login, and logout.
- Restricts snippet creation to authenticated users.
- Hashes passwords with bcrypt and stores login sessions in MySQL.
- Validates submitted forms and protects forms with CSRF tokens.
- Embeds HTML templates and static assets into the executable.

## Requirements

- Go 1.27.1 or later (as specified by `go.mod`).
- MySQL, with a database and application user configured.
- A TLS certificate and private key at `assets/tls/cert.pem` and `assets/tls/key.pem`.

The server always starts with HTTPS via `ListenAndServeTLS`. The TLS directory is not currently included in the repository, so provide those files before running the application. For local development, a self-signed certificate can be generated with Go's `crypto/tls` tooling or another certificate utility.

## Database setup

Create a database and tables matching the model queries. For example:

```sql
CREATE DATABASE snippetbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE snippetbox;

CREATE TABLE snippets (
    id      INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    title   VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    created DATETIME NOT NULL,
    expires DATETIME NOT NULL,
    INDEX (expires)
);

CREATE TABLE users (
    id              INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    hashed_password CHAR(60) NOT NULL,
    created         DATETIME NOT NULL,
    CONSTRAINT users_uc_email UNIQUE (email)
);
```

The unique constraint name `users_uc_email` is used by the signup handler to identify duplicate email errors. The SCS MySQL session store also requires its `sessions` table. Create it using the schema for the version of `github.com/alexedwards/scs/mysqlstore` in `go.mod`; session data is stored in that table. Configure the MySQL account with access to create/use these tables as needed.

## Run

Install dependencies and start the server from the repository root:

```sh
go mod download
go run ./cmd/web
```

The default address is `:4000`, and the default MySQL DSN is `web:pass@/snippetbox?parseTime=true`. Supply values appropriate to your environment with flags:

```sh
go run ./cmd/web -addr=:4000 -dsn='user:password@tcp(127.0.0.1:3306)/snippetbox?parseTime=true'
```

The DSN is a command-line flag rather than an environment variable. Avoid putting production credentials directly in shell history; change the configuration approach before deploying if secrets need to be managed outside command-line arguments.

Once the server is running, visit `https://localhost:4000/`. The `GET /ping` endpoint returns `OK` for a basic health check. Static files are served under `/static/` from the embedded `assets/static` tree.

## Routes

| Method | Path | Purpose | Access |
| --- | --- | --- | --- |
| GET | `/` | Show the latest ten active snippets | Public |
| GET | `/snippet/view/{id}` | Show one active snippet | Public |
| GET | `/user/signup` | Show the signup form | Public |
| POST | `/user/signup` | Create an account | Public, CSRF protected |
| GET | `/user/login` | Show the login form | Public |
| POST | `/user/login` | Authenticate a user | Public, CSRF protected |
| GET | `/snippet/create` | Show the snippet form | Authenticated |
| POST | `/snippet/create` | Save a snippet | Authenticated, CSRF protected |
| POST | `/user/logout` | End the login session | Authenticated, CSRF protected |
| GET | `/ping` | Health check | Public |
| GET | `/static/...` | Serve embedded static assets | Public |

Snippet expiration choices are one day, one week, or one year. Expired snippets are excluded from both the home page and individual snippet lookup.

## Project layout

```text
assets/
  efs.go                 Embedded HTML and static files
  html/                  Base layout, partials, and page templates
  static/                CSS, JavaScript, and image assets
cmd/web/
  main.go                Flags, database connection, dependencies, HTTPS server
  routes.go              HTTP route and middleware configuration
  handlers.go            Page handlers and form processing
  middleware.go          Headers, logging, panic recovery, auth, CSRF
  templates.go           Embedded template parsing and cache
  helpers.go             Rendering, form decoding, and request helpers
internal/
  models/                 MySQL access for snippets and users
  validator/              Form validation helpers
  assert/                 Test assertion helpers
```

The `application` type in `cmd/web` holds shared dependencies. Requests pass through common logging, header, and panic-recovery middleware. Dynamic routes also load/save sessions, apply CSRF protection, and check authentication; snippet creation and logout add an authentication requirement.

## Tests

Run the existing Go test suites with:

```sh
go test ./...
```

The handler and template tests use the helpers in `cmd/web/testutils_test.go` and Go's `net/http/httptest` package.
