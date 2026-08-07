# Small, portable build for the Integrations gRPC service.
FROM golang:1.26.5-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY integrations ./integrations
COPY grpc ./grpc

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/integrations-grpc-service ./integrations/cmd/grpc-service

# Distroless keeps the runtime image small and runs without a shell or root user.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app/integrations

COPY --from=build /out/integrations-grpc-service ./integrations-grpc-service
COPY --from=build /src/integrations/db/migrations ./db/migrations

USER nonroot:nonroot
EXPOSE 50051

ENTRYPOINT ["./integrations-grpc-service"]