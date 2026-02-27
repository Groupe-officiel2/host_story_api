# Vintage Story server (containerized)

Place `vs_server` in this folder before building.

Build and run (ephemeral, no host data persistence):

```bash
cd host_story_api_go/server
docker compose build --pull --no-cache
docker run -d --name server3 -p 42820:42420 -p 42820:42420/udp server-vintagestory:latest
```

Ephemeral mode:
- No `./data:/data` mount is used; data lives inside the container and is lost when it is removed.
- The server uses `/tmp/vsdata` as its data path by default.

Notes:
- If the image extracts a `VintagestoryServer` native executable it will run that; otherwise it will use `dotnet VintagestoryServer.dll`.
- To make the server public, ensure ports 42420/udp and 42420/tcp are open and forwarded to the host.
- A full `serverconfig.json` template is copied into the image at build time and used when the data path is empty.
