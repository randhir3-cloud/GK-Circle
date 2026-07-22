FROM golang:1.23.11-alpine3.22 AS build
WORKDIR /gk-circle
RUN apk update && apk add build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o gk-circle
FROM alpine
WORKDIR /gk-circle
RUN apk add --no-cache ca-certificates wget
COPY --from=build /gk-circle/gk-circle .
COPY --from=build /gk-circle/database ./database
COPY --from=build /gk-circle/assets ./assets
