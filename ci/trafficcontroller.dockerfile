FROM golang:1.8.3

RUN apt-get update && \
        apt-get install -y telnet && \
        apt-get install -y python


ENV GOPATH /go
RUN go get -d github.com/chenziliang/loggregator; exit 0
WORKDIR /go/src/github.com/chenziliang/loggregator
RUN git checkout feature/firehose-standalone
ENV GOPATH /go/src/github.com/chenziliang/loggregator
RUN cd /go/src/github.com/chenziliang/loggregator && ./scripts/build
WORKDIR /go/src/github.com/chenziliang/loggregator
RUN mv ./bin/trafficcontroller .
RUN git clone https://github.com/cloudfoundry-community/splunk-firehose-nozzle

EXPOSE 9911
EXPOSE 8081

CMD ["/usr/bin/python", "splunk-firehose-nozzle/ci/perf.py", "--run", "trafficcontroller", "--duration", "1200"]
