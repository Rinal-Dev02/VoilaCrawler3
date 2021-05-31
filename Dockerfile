FROM ubuntu:20.04

MAINTAINER <kvcnow@gmail.com>

RUN apt update && apt install -y ca-certificates
ADD releases /usr/bin/

EXPOSE 6000