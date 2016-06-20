FROM gliderlabs/alpine:3.2
WORKDIR /bin
ENTRYPOINT ["/bin/mesos-visualizer"]

RUN apk-install ca-certificates
COPY mesos-visualizer /bin/mesos-visualizer
ADD ./static/* /bin/static/
