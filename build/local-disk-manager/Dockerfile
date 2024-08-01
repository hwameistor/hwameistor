FROM rockylinux:8

RUN yum install -y xfsprogs smartmontools lsscsi e4fsprogs nss udev
COPY ./_build/local-disk-manager /local-disk-manager

ENTRYPOINT [ "/local-disk-manager" ]
