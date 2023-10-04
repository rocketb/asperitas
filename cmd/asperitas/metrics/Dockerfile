# Build the Go Binary.
FROM golang:1.21 as build_app
ENV CGO_ENABLED 0
ARG BUILD_REF

# Copy the source code into the container.
COPY . /service

# Build the service metrics binary.
WORKDIR /service/cmd/asperitas/metrics
RUN go build -ldflags "-X main.build=${BUILD_REF}" -o ./metrics


# Run the Go Binary in Alpine.
FROM alpine:3.18
ARG BUILD_DATE
ARG BUILD_REF
RUN addgroup -g 1000 -S metrics && \
    adduser -u 1000 -h /service -G metrics -S metrics
COPY --from=build_app --chown=asperitas:asperitas /service/cmd/asperitas/metrics/metrics /service/metrics
WORKDIR /service
USER metrics
CMD ["./metrics"]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="metrics" \
      org.opencontainers.image.authors="App authors" \
      org.opencontainers.image.source="https://github.com/rocketb/asperitas/app/asperitas" \
      org.opencontainers.image.revision="${BUILD_REF}" \
      org.opencontainers.image.vendor="App authors."
