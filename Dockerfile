FROM golang:1.22-alpine as build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /muxc main.go

FROM scratch

COPY --from=build /muxc /muxc

ENTRYPOINT ["/muxc"]