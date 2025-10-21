FROM golang:1.24 AS builder
WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 go build -o collector cmd/collector/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/collector .
USER 65532:65532
ENTRYPOINT ["/collector"]
