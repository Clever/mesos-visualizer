FROM alpine:3.10
WORKDIR /bin
ENTRYPOINT ["/bin/mesos-visualizer"]

RUN apk add ca-certificates && update-ca-certificates
COPY mesos-visualizer /bin/mesos-visualizer
ADD ./static/* /bin/static/
