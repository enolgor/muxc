FROM golang:1.23-alpine AS build

WORKDIR /app

COPY ./ ./

ENTRYPOINT ["/bin/sh"]

RUN cd muxc && go build -o /muxc .

FROM scratch

COPY --from=build /muxc /muxc

ENTRYPOINT ["/muxc"]