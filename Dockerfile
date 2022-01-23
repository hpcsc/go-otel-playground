FROM golang:1.16 AS builder
WORKDIR /build

# Precompile the entire go standard library
RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go install -v -a std

# Download and precompile all third party libraries, ignoring errors
ADD go.mod .
ADD go.sum .
RUN go mod download -x
RUN go list -f '{{.Path}}/...' -m all | GOARCH=amd64 GOOS=linux CGO_ENABLED=0 xargs -n1 go build -v -i; echo done

#  Add the sources
ADD . .

# Compile only our sources
RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -v --ldflags '-extldflags -static' -o go-otel-playground ./cmd/main.go

FROM scratch
WORKDIR /app
ENTRYPOINT ["/app/go-otel-playground"]
COPY --from=builder /build/go-otel-playground .
