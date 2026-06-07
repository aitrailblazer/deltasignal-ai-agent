FROM golang:1.24-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/deltasignal-ai-agent ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/deltasignal-ai-agent /deltasignal-ai-agent
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/deltasignal-ai-agent"]
