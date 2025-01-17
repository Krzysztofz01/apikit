FROM golang:latest as build
RUN go install github.com/go-task/task/v3/cmd/task@latest
RUN mkdir /apikit
COPY . /apikit
WORKDIR /apikit
RUN task build

FROM ubuntu:latest as publish
RUN apt-get update && \
    apt-get install --reinstall --no-install-recommends -y ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
RUN mkdir /apikit
COPY --from=build /apikit/bin/apikit /apikit/apikit
WORKDIR /apikit
ENTRYPOINT ["/apikit/apikit", "run"]
