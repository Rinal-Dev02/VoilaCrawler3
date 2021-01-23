FROM ubuntu:18.04

MAINTAINER <kvcnow@gmail.com>

ADD releases/plugins /plugins
ADD releases/ /usr/bin/

EXPOSE 6000
EXPOSE 8080
