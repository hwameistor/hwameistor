FROM centos:7

RUN yum install -y xfsprogs udev smartmontools lsscsi

COPY ./_build/local-storage /

ENTRYPOINT [ "/local-storage" ]
