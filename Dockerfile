# Use an official Golang image as the base for building the applications
FROM golang:1.22.5 as builder

# Set the working directory
WORKDIR /app

# Clone the mochimo-mesh repository
RUN git clone https://github.com/NickP005/mochimo-mesh.git

# Clone the mochimo repository inside mochimo-mesh
RUN git clone https://github.com/mochimodev/mochimo.git mochimo-mesh/mochimo

# Build the Mochimo Mesh API
WORKDIR /app/mochimo-mesh
RUN go build -o mesh .

# Build the Mochimo Node
WORKDIR /app/mochimo-mesh/mochimo
RUN make -C . install-mochimo

# Use a newer Ubuntu image for the final stage
FROM ubuntu:22.04

# Set environment variables
ENV PATH="/usr/local/bin:${PATH}"

# Install necessary dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /root

# Copy the built binaries from the builder stage
COPY --from=builder /app/mochimo-mesh/ /usr/local/bin/
COPY --from=builder /app/mochimo-mesh/mochimo/bin /usr/local/bin/mochimo/bin/

# Expose the necessary ports
EXPOSE 8080 2095

# Workdir /usr/local/bin/
WORKDIR /usr/local/bin/

# Start both the Mochimo node and the mesh API
CMD ["sh", "-c", "./mochimo/bin/gomochi -n & exec mesh"]
