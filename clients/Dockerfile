# Small, portable build for the Clients gRPC service.
FROM golang:1.26.5-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY clients ./clients
COPY shared ./shared
COPY grpc ./grpc
COPY protobuf ./protobuf

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/clients-grpc-service ./clients/cmd/grpc-service

# Distroless keeps the runtime image small and runs without a shell or root user.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app/clients

COPY --from=build /out/clients-grpc-service ./clients-grpc-service
COPY --from=build /src/clients/db/migrations ./db/migrations

USER nonroot:nonroot
EXPOSE 50051

ENTRYPOINT ["./clients-grpc-service"]