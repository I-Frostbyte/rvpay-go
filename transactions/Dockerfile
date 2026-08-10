# Small, portable build for the Transactions gRPC service.
FROM golang:1.26.5-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY transactions ./transactions
COPY grpc ./grpc
COPY protobuf ./protobuf

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/transactions-grpc-service ./transactions/cmd/grpc-service

# Distroless keeps the runtime image small and runs without a shell or root user.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app/transactions

COPY --from=build /out/transactions-grpc-service ./transactions-grpc-service
COPY --from=build /src/transactions/db/migrations ./db/migrations

USER nonroot:nonroot
EXPOSE 50051

ENTRYPOINT ["./transactions-grpc-service"]