# Step 1: Build stage for Bonsai binaries
FROM golang:1.23 AS builder

COPY go.mod go.sum ./

RUN go mod download

# Copy all source code
COPY . .

# Compile bonsaid (main service binary)
RUN CGO_ENABLED=0 go build -o /bonsaid ./cmd/bonsaid

# Compile bonsai-cli (command-line API client binary)
RUN CGO_ENABLED=0 go build -o /bonsai-cli ./cmd/bonsai-cli

# Compile bonsai-test (test runner binary, not included in production)
RUN CGO_ENABLED=0 go build -o /bonsai-test ./cmd/bonsai-test

# Step 2: Minimal runtime image for production
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache bash curl

# Copy only production binaries from the build stage
COPY --from=builder /bonsaid /bonsaid
COPY --from=builder /bonsai-cli /bonsai-cli

# Expose required ports
EXPOSE 8080 8081

# Command to run the main service
CMD ["/bonsaid"]
