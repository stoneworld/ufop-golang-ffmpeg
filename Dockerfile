FROM jrottenberg/ffmpeg:4.1-alpine AS base

COPY --from=golang:1.15.5-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

RUN apk update && apk add build-base pkgconfig

WORKDIR /src

COPY go.* ./

FROM base AS build

COPY . .

RUN go build -o /sy-ffmpeg-api server.go

RUN chmod +x /sy-ffmpeg-api

FROM jrottenberg/ffmpeg:4.1-alpine AS run

COPY --from=build /sy-ffmpeg-api .


ENTRYPOINT [ "/usr/bin/env" ]

CMD [ "/sy-ffmpeg-api", "-env" ]
