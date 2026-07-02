# One parameterized multi-stage build for all three Go services.
# Compose passes SERVICE=api-gateway|shortener|redirector as a build arg.
FROM golang:1.26-alpine AS build
ARG SERVICE
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./${SERVICE}

FROM alpine:3.22
RUN adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=build /out/app /app/app
ENTRYPOINT ["/app/app"]
