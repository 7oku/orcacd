## BUILD STAGE
FROM golang:alpine as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./

RUN CGO_ENABLED=0 GOOS=linux go build -o ./orcacd

# FINAL BUILD
#FROM gicr.io/distroless/static-debian12
FROM alpine
RUN apk update && apk add --no-cache git
COPY --from=build /app/orcacd /
CMD ["/orcacd"]
