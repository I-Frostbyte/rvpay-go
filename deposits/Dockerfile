# Small, portable build for the Deposits gRPC service.
FROM golang:1.26.5-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY deposits ./deposits
COPY grpc ./grpc

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/deposits-grpc-service ./deposits/cmd/grpc-service

# Distroless keeps the runtime image small and runs without a shell or root user.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app/deposits

COPY --from=build /out/deposits-grpc-service ./deposits-grpc-service
COPY --from=build /src/deposits/db/migrations ./db/migrations

USER nonroot:nonroot
EXPOSE 50051

ENTRYPOINT ["./deposits-grpc-service"]
