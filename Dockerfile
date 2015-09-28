FROM gliderlabs/alpine:3.2
WORKDIR /bin
ENTRYPOINT ["/bin/mesos-visualizer"]

COPY . /go/src/github.com/Clever/mesos-visualizer
ADD ./static/* /bin/static/
RUN apk-install -t build-deps go git \
    && cd /go/src/github.com/Clever/mesos-visualizer \
    && export GOPATH=/go \
    && go get \
    && go build -o /bin/mesos-visualizer \
    && rm -rf /go \
    && apk del --purge build-deps
