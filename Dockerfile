FROM golang:1.20 as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static" -s -w' -o main github.com/codemicro/analytics/ingest

FROM python:3
ENV PIP_DISABLE_PIP_VERSION_CHECK=on

RUN mkdir -p /analytics/ds/
WORKDIR /analytics

COPY --from=builder /build/main ./ingest

ADD datasette_plugin ds/plugins
ADD config/datasette-metadata.json ds/metadata.json
ADD docker-entrypoint.sh .

RUN pip install --no-cache-dir datasette
RUN pip install --no-cache-dir -r ds/plugins/requirements.txt

RUN mkdir -p /analytics/run
WORKDIR /analytics/run

STOPSIGNAL SIGKILL

CMD ["bash", "../docker-entrypoint.sh"]