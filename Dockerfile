FROM ubuntu
MAINTAINER Shantanu Joshi <shantanu.joshi@coreos.com>
ENV DEBIAN_FRONTEND noninteractive
ENV HOME /root
RUN mkdir -p /root/Golang/bin
# for the repo that will be cloned.
RUN mkdir -p /root/.ssh
WORKDIR /root
#Using the reserve HTTP port.
RUN apt-get update
RUN apt-get install -y libssh2-1-dev
RUN apt-get install -y libssh2-1
# Clean up any files used by apt-get
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
EXPOSE 8888 8080
ADD bin/server /root/Golang/bin/server
RUN mkdir -p /root/ssh
CMD ["/root/Golang/bin/server"]
