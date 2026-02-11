# Vintage Story server (containerized)

Place `vs_server_linux-x64_1.21.6.tar.gz` in this folder before building.

Build and run with docker-compose:

```bash
cd host_story_api_go/server
docker compose build --pull --no-cache
docker compose up -d
```

The server data will be persisted in `host_story_api_go/server/data`.

## API Documentation

The Go API allows you to manage Vintage Story servers dynamically. Below are the available endpoints:

### Endpoints

#### 1. Home Endpoint
- **URL**: `/`
- **Method**: `GET`
- **Description**: Returns a simple "Hello world!" message to confirm the API is running.

#### 2. Create Template Container
- **URL**: `/template`
- **Method**: `GET`
- **Headers**:
  - `X-API-Key`: Your API key for authentication (e.g., `secret123`).
- **Query Parameters**:
  - `image` (optional): The Docker image to use for the container. Default is `server-vintagestory:latest`.
- **Description**: Creates a new container from the specified image. The API automatically assigns a unique name (e.g., `server1`, `server2`, etc.) and dynamically maps the ports.

#### Example Request
```bash
curl -H "X-API-Key: secret123" "http://localhost:8080/template?image=server-vintagestory:latest"
```

#### Example Response
```
Container launched: <container_id> with name server1 on host port 42720
```

### Notes
- The API requires an environment variable `API_KEY` to be set for authentication.
- The server names and ports are managed automatically by the API.
- The base host port starts at `42720` and increments for each new server.
- The container port is fixed at `42420` (default for Vintage Story).

### Running the API
To start the API server:

```bash
cd host_story_api
API_KEY=secret123 go run main.go
```

Replace `secret123` with your desired API key.

Once the server is running, you can use the endpoints to create and manage containers dynamically.
