# Stage 1: build the Go binary
FROM  golang:1.26.2-alpine3.22 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o latex-api ./cmd/server

# Stage 2: runtime image based on texlive
FROM texlive/texlive:latest
RUN apt-get update && apt-get upgrade -y --no-install-recommends \
    && apt-get install -y --no-install-recommends \
    git \
    poppler-utils \
    p7zip-full \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/latex-api /usr/local/bin/latex-api

EXPOSE 8080
CMD ["latex-api"]
