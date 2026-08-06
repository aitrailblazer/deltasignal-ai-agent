FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/aapl-client-demo-server ./cmd/aapl-client-demo-server
COPY internal/aapldemo ./internal/aapldemo
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/aapl-client-demo ./cmd/aapl-client-demo-server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/aapl-client-demo /aapl-client-demo
USER nonroot:nonroot
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/aapl-client-demo"]
