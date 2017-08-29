FROM golang:1.8.3

RUN apt-get update && \
        apt-get install -y telnet && \
        apt-get install -y gcc && \
        apt-get install -y screen && \
        apt-get install -y python-dev && \
        apt-get install -y python-pip && \
        apt-get install -y python

RUN pip install psutil && pip install requests --upgrade

ENV GOPATH /go

ENV HEC_WORKERS=8
ENV ADD_APP_INFO=true
ENV SKIP_SSL_VALIDATION_CF=true
ENV SKIP_SSL_VALIDATION_SPLUNK=true

ENV API_ENDPOINT=http://trafficcontroller:9911
ENV API_USER=admin
ENV API_PASSWORD=admin

ENV EVENTS=ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric
ENV SPLUNK_TOKEN=1CB57F19-DC23-419A-8EDA-BA545DD3674D
ENV SPLUNK_HOST=https://heclb1:8088
ENV SPLUNK_INDEX=main
ENV FLUSH_INTERVAL=30s
ENV FIREHOSE_SUBSCRIPTION_ID=spl-nozzle-perf-testing

ENV NOZZLE_BRANCH=master

ADD nozzle.sh /

CMD ["/bin/sh", "-c", "/nozzle.sh"]
