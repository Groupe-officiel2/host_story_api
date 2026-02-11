FROM mcr.microsoft.com/dotnet/runtime:8.0

WORKDIR /opt/vintagestory


COPY vs_server /tmp/

RUN apt-get update \
  && apt-get install -y --no-install-recommends tzdata jq \
  && rm -rf /var/lib/apt/lists/* \
  && mkdir -p /opt/vintagestory \
  && tar -xzf /tmp/vs_server_linux-x64_1.21.6.tar.gz -C /opt/vintagestory \
  && rm /tmp/vs_server_linux-x64_1.21.6.tar.gz \
  && chmod +x /opt/vintagestory/VintagestoryServer || true

COPY serverconfig.json /opt/vintagestory/serverconfig.json
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Vintage Story default port
EXPOSE 42420/udp
EXPOSE 42420/tcp

WORKDIR /opt/vintagestory
ENTRYPOINT ["/entrypoint.sh"]
