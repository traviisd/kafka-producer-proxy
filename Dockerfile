# Build image
FROM golang:buster AS builder

# Install ca-certificates
# Git is required for fetching the dependencies.
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Create appuser.
ENV USER=appuser
ENV UID=10001 

# See https://stackoverflow.com/a/55757473/12429735RUN 
RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

WORKDIR $GOPATH/src/github.com/traviisd/kafka-producer-proxy

COPY . .

RUN pwd && ls -al

# Build the binary.
RUN go build -ldflags="-s -w" -o /go/bin/kafka-producer-proxy *.go

# Run image
FROM debian:buster-slim

# Import certs
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

WORKDIR /go/bin

# Copy static executable.
COPY --from=builder --chown=appuser:appuser /go/bin/kafka-producer-proxy kafka-producer-proxy

# Use an unprivileged user.
USER appuser:appuser
# Used for cert auth flow with vault & kafka
RUN mkdir -p /tmp/proxy

# Run the binary.
ENTRYPOINT ["./kafka-producer-proxy"]
 