FROM centos:7

RUN rm -rf /etc/yum.repos.d/*
COPY ./build/CentOS-Base.repo /etc/yum.repos.d/
RUN yum install -y xfsprogs e2fsprogs udev smartmontools lsscsi && \
    yum upgrade nss -y

COPY ./_build/local-storage /

ENTRYPOINT [ "/local-storage" ]
