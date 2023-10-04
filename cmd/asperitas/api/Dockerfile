# Build the Go Binary.
FROM golang:1.21 as build_app
ENV CGO_ENABLED 0
ARG BUILD_REF

# Copy the source code into the container.
COPY . /service

# Build the admin binary.
WORKDIR /service/cmd/tools/asperitas-admin
RUN go build -ldflags "-X main.build=${BUILD_REF}" -o ./asperitas-admin

# Build the service binary.
WORKDIR /service/cmd/asperitas/api
RUN go build -ldflags "-X main.build=${BUILD_REF}" -o ./asperitas-api


# Run the Go Binary in Alpine.
FROM alpine:3.18
ARG BUILD_DATE
ARG BUILD_REF
RUN addgroup -g 1000 -S asperitas && \
    adduser -u 1000 -h /service -G asperitas -S asperitas
COPY --from=build_app --chown=asperitas:asperitas /service/deploy/keys/. /service/deploy/keys/.
COPY --from=build_app --chown=asperitas:asperitas /service/cmd/tools/asperitas-admin/asperitas-admin /service/asperitas-admin
COPY --from=build_app --chown=asperitas:asperitas /service/cmd/asperitas/api/asperitas-api /service/asperitas-api
WORKDIR /service
USER asperitas
CMD ["./asperitas-api"]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.title="asperitas" \
      org.opencontainers.image.authors="Asperitas authors" \
      org.opencontainers.image.source="https://github.com/rocketb/asperitas/app/asperitas" \
      org.opencontainers.image.revision="${BUILD_REF}" \
      org.opencontainers.image.vendor="Asperitas authors."
