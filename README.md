## API Documentation

The Go API allows you to manage Vintage Story servers dynamically. Below are the available endpoints:

#### Example Response
```
Container launched: <container_id> with name server3 on host port 42722 and 3 player slots
```

### Notes
- The API requires an environment variable `API_KEY` to be set for authentication.
- The server names and ports are managed automatically by the API.
- The base host port starts at `42720` and increments for each new server.
- The container port is fixed at `42420` (default for Vintage Story).
- The memory allocation is dynamic based on the number of players:
  - 1 player: 1.30GB (default)
  - Each additional player adds 300MB.

### Running the API

Set the `API_KEY` environment variable and run the server:

```bash
export API_KEY=secret123
```

Start the server:

```bash
go run .
```

To specify the number of players (e.g., 3 players, which adds 300MB per player to the default 1.30GB memory):

```bash
curl -H "X-API-Key: secret123" "http://localhost:8080/template?image=server-vintagestory:latest&players=2&name=example"
curl -H "X-API-Key: secret123" "http://localhost:8080/toggle?name=example"
```


