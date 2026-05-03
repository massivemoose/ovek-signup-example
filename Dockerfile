# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25.5-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -ldflags="-s -w" -o /out/ovek-signup-example .

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/massivemoose/ovek-signup-example"

ENV PORT=8080

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/ovek-signup-example /ovek-signup-example

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/ovek-signup-example"]
