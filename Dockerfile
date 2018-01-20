# Build burst scheduler binary
FROM golang as builder
RUN go get k8s.io/client-go/...
ADD . /app
WORKDIR /app
RUN go build -o burst-scheduler

# Build burst scheduler container image
FROM golang
WORKDIR /app
COPY --from=builder /app/burst-scheduler .
ENTRYPOINT ["/app/burst-scheduler"]