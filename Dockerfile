FROM golang:1.20-alpine AS builder
WORKDIR /build
ADD go.mod .
COPY . .
RUN go build -o /build/unifi_exporter ./cmd/unifi_exporter

FROM scratch
EXPOSE 9130
COPY --from=builder /build/unifi_exporter /bin/unifi_exporter
USER 65534
CMD ["/bin/unifi_exporter"]
