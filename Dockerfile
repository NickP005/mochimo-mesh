# Mochimo MeshAPI 3.0 beta Dockerfile

# Use an official Golang image as the base for building the applications
FROM golang:1.22.5 as builder

# Set the working directory
WORKDIR /app

# TO ADD TAGGED VERSIONS FOR RESPECT TO THE COINBASE REQUIREMENTS

# Clone mochimo-mesh from branch 3.0
RUN apt-get update && apt-get install -y git && \
    git clone --depth 1 -b 3.0 https://github.com/NickP005/mochimo-mesh.git && \
    rm -rf /var/lib/apt/lists/*

# Clone the mochimo repository with specific branch
RUN git clone -b v3rc2 https://github.com/mochimodev/mochimo.git /app/mochimo-mesh/mochimo

# Build the Mochimo Mesh API
WORKDIR /app/mochimo-mesh
RUN go build -o mesh .

# Build the Mochimo Node
WORKDIR /app/mochimo-mesh/mochimo
RUN make mochimo

# Use a newer Ubuntu image for the final stage
FROM ubuntu:22.04

# Set environment variables
ENV PATH="/usr/local/bin:${PATH}"

# Install necessary dependencies
RUN apt-get update && apt-get install -y \

ca-certificates \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# Create initialization script
RUN echo '#!/bin/bash\n\
if [ -z "$(ls -A /data/mochimo)" ]; then\n\
    echo "Initializing mochimo data directory..."\n\
    touch /data/mochimo/.initialized\n\
fi\n\
if [ -z "$(ls -A /data/mesh)" ]; then\n\
    echo "Initializing mesh data directory..."\n\
    echo "{\n  \"port\": 8081,\n  \"mochimo_binary_path\": \"/usr/local/bin/mochimo/bin/gomochi\"\n}" > /data/mesh/settings.json\n\
    echo "{\n  \"port\": 2098,\n  \"ip\": \"0.0.0.0\"\n}" > /data/mesh/interface_settings.json\n\
fi\n\
cd /data/mochimo && /usr/local/bin/mochimo/bin/gomochi -p2098 -F & \n\
cd /data/mesh && exec /usr/local/bin/mesh -solo 0.0.0.0' > /usr/local/bin/init.sh && \
    chmod +x /usr/local/bin/init.sh

# Create directories for persistent data
RUN mkdir -p /data/mochimo /data/mesh

# Set the working directory
WORKDIR /data

# Copy the built binaries from the builder stage
COPY --from=builder /app/mochimo-mesh/ /usr/local/bin/
COPY --from=builder /app/mochimo-mesh/mochimo/bin /usr/local/bin/mochimo/bin/

# Create volumes for persistence
VOLUME ["/data/mochimo", "/data/mesh"]

# Expose the necessary ports
EXPOSE 8081 2098

# Use the initialization script as entrypoint
ENTRYPOINT ["/usr/local/bin/init.sh"]