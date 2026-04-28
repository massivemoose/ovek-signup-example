# Ovek Email List Demo

A tiny Go signup app that writes email addresses to a PocketBase sidecar deployed with Ovek.

![Signup page](images/signup.png)

## Deployment Example Using Ovek

[TODO]

## Local Development

This app expects PocketBase to be running locally at `http://127.0.0.1:8090`. On startup, the Go app authenticates as a PocketBase superuser, ensures the `signups` collection exists, and then serves the signup page at `http://localhost:8080`.

### 1. Install PocketBase

Download the PocketBase binary for your platform from the official PocketBase docs:

https://pocketbase.io/docs/

Place the extracted `pocketbase` executable in this repo root. The app's `.gitignore` ignores the local binary and PocketBase data directories.

### 2. Start PocketBase

In one terminal, start PocketBase:

```bash
./pocketbase serve
```

On first run, PocketBase will prompt you to create a superuser account. Follow the local setup link it prints, or open the admin dashboard:

```text
http://127.0.0.1:8090/_/
```

Keep this terminal running while you use the Go app.

### 3. Provide PocketBase superuser credentials

The Go app needs either `PB_SUPERUSER_TOKEN` or both `PB_SUPERUSER_EMAIL` and `PB_SUPERUSER_PASSWORD`.

For a one-off run, pass the credentials inline:

```bash
PB_SUPERUSER_EMAIL="admin@example.com" PB_SUPERUSER_PASSWORD="your-password" go run ./...
```

For repeated local development, you can use direnv:

https://direnv.net/docs/installation.html

Create a local `.envrc` file:

```bash
export PB_SUPERUSER_EMAIL="admin@example.com"
export PB_SUPERUSER_PASSWORD="your-password"
```

Allow direnv to load it:

```bash
direnv allow
```

Verify the variables are loaded:

```bash
env | rg '^PB_SUPERUSER_'
```

Then run the app:

```bash
go run ./...
```

### 4. Use the app

Open the signup app:

```text
http://localhost:8080
```

Submit a valid email address. Successful submissions redirect to `/success`; invalid or duplicate submissions redirect to `/failure`.

### 5. Inspect signups in PocketBase

Open the PocketBase dashboard:

```text
http://127.0.0.1:8090/_/
```

Sign in with the local superuser account, then open the `signups` collection to inspect submitted email records.

## Useful Commands

Run tests:

```bash
go test ./...
```

Run the app with an alternate port:

```bash
PORT=8081 PB_SUPERUSER_EMAIL="admin@example.com" PB_SUPERUSER_PASSWORD="your-password" go run ./...
```

Run the app against a non-default PocketBase URL:

```bash
POCKETBASE_URL="http://127.0.0.1:8091" PB_SUPERUSER_EMAIL="admin@example.com" PB_SUPERUSER_PASSWORD="your-password" go run ./...
```
