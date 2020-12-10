FROM oraclelinux:7

ENV SHELL=/bin/sh

ADD ./build-extra/ol7/repos /etc/yum.repos.d/
ADD ./build-extra/ol7/keys /etc/pki/rpm-gpg/

RUN rpm --import /etc/pki/rpm-gpg/RPM-GPG-KEY-EPEL-7
RUN rpm --import /etc/pki/rpm-gpg/RPM-GPG-KEY-oracle

RUN yum install -y deltarpm

RUN yum install -y which locate make gcc.x86_64 pcp-devel
# Preinstalling some golang dependencies
RUN yum install -y apr git less neon nettle pakchois perl rsync subversion mercurial
# Installing golang as a seperate line as it appears to hang if we don't
RUN yum install -y golang 

RUN git config --global url."https://oauth2:2AagziKsFXdTYdcwFotu@cloudlab.us.oracle.com:2222".insteadOf "https://cloudlab.us.oracle.com"

WORKDIR /go/src/pcp-exporter
COPY . .

RUN go get -d -v

ENTRYPOINT [ "./container-build.sh" ] 