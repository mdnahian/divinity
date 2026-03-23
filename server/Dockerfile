FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /divinity .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=build /divinity /divinity

EXPOSE 8080

ENTRYPOINT ["/divinity"]
