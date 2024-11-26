FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -tags "$BUILDTAGS" --ldflags '-extldflags "-static"' -v

FROM scratch
COPY --from=builder /app/jocker /bin/jocker
ENTRYPOINT ["/bin/jocker", "build"]
