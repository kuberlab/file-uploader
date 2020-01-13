FROM alpine:3.9

COPY . /go/src/github.com/kuberlab/file-uploader

RUN apk --no-cache add -t build-deps build-base go git \
	&& apk --no-cache add ca-certificates \
	&& cd /go/src/github.com/kuberlab/file-uploader \
	&& export GOPATH=/go \
	&& go get \
	&& go build -ldflags="-s -w" -o /bin/file-uploader \
	&& rm -rf /go \
	&& apk del --purge build-deps

EXPOSE 8082

ENTRYPOINT ["/bin/file-uploader"]

