FROM golang:alpine AS builder
COPY . .
RUN CGO_ENABLED=0 go build -o /jocker -tags "$BUILDTAGS" --ldflags '-extldflags "-static"'

FROM scratch
COPY --from=builder /jocker /bin/jocker
ENTRYPOINT ["/bin/jocker", "build"]