FROM ubuntu
ENV DEBIAN_FRONTEND noninteractive
ENV HOME /root
RUN mkdir -p /root/Golang/bin
RUN mkdir -p /root/ssh
WORKDIR /root
# libssh2 dependency
RUN apt-get update
RUN apt-get install -y libssh2-1-dev
RUN apt-get install -y libssh2-1
# Clean up any files used by apt-get
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
ADD bin/server /root/Golang/bin/server
CMD ["/root/Golang/bin/server"]
