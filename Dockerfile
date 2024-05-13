FROM --platform=$BUILDPLATFORM golang:alpine AS build

RUN apk add --no-cache git

RUN git clone https://github.com/TheTipo01/broadBanBot /broadBanBot
WORKDIR /broadBanBot
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o broadBanBot

FROM alpine

COPY --from=build /broadBanBot/broadBanBot /usr/bin/

CMD ["broadBanBot"]
