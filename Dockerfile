# syntax=docker/dockerfile:1

FROM golang:1.22-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/certificatemgmt .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/certificatemgmt /certificatemgmt

USER nonroot:nonroot
EXPOSE 8080

ENV PORT=8080

ENTRYPOINT ["/certificatemgmt"]
