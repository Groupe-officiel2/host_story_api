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

Start the server:

```bash
# Test endpoint protégé
TOKEN=$(python3 generate_jwt.py | awk '{print $4}')
curl -H "Authorization: $TOKEN" http://127.0.0.1:8082/protected
# Crée un serveur nommé paladium avec 2 joueurs
curl -H "Authorization: $TOKEN" "http://127.0.0.1:8082/template?image=server-vintagestory:latest&players=2&name=example"
# Toggle le serveur paladium
curl -H "Authorization: $TOKEN" "http://82.67.195.93:8082/toggle?name=example"
```

```bash
docker build -f build/docker/Dockerfile -t host-story-api:1.0 .
```

```bash
docker run -d \
  --name host-story-api \
  --restart unless-stopped \
  --network host \
  -v /var/run/docker.sock:/var/run/docker.sock \
  host-story-api:1.0
```

```bash
  docker run -d \
  --name host-story-api \
  --restart unless-stopped \
  -p 8082:8082 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  host-story-api:1.0
```
