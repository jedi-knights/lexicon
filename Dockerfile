# Stage 1: build the static binary.
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /out/lexicon \
    ./cmd/lexicon

# Stage 2: minimal runtime image.
FROM gcr.io/distroless/static-debian12

COPY --from=builder /out/lexicon /usr/local/bin/lexicon

ENTRYPOINT ["/usr/local/bin/lexicon"]
CMD ["compile", "--list-targets"]
