FROM rockylinux:8

RUN yum update -y ca-certificates \
    && yum install -y xfsprogs smartmontools lsscsi e4fsprogs nss udev \
    && yum clean all
COPY ./_build/local-disk-manager /local-disk-manager

ENTRYPOINT [ "/local-disk-manager" ]
