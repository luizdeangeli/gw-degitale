FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gw ./cmd/gw

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget

COPY --from=build /out/gw /gw

USER nobody
EXPOSE 8080
ENTRYPOINT ["/gw"]
