FROM rockylinux:8


RUN yum update -y ca-certificates \
    && yum install -y xfsprogs e2fsprogs udev smartmontools lsscsi nss cryptsetup \
    && yum clean all

COPY ./_build/local-storage /

ENTRYPOINT [ "/local-storage" ]
