FROM alpine:latest

# Install ca-certificates for HTTPS support
RUN apk --no-cache add ca-certificates

# Copy the binary from GoReleaser build
COPY delinea-netconfig /usr/local/bin/delinea-netconfig

# Create non-root user
RUN addgroup -S delinea && adduser -S -G delinea delinea
USER delinea

# Set working directory
WORKDIR /data

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/delinea-netconfig"]
CMD ["--help"]
