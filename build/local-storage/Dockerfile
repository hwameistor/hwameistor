FROM centos:7

RUN yum install -y xfsprogs e2fsprogs udev smartmontools lsscsi && \
    yum upgrade nss -y

COPY ./_build/local-storage /

ENTRYPOINT [ "/local-storage" ]
