FROM rockylinux:8


RUN yum install -y xfsprogs e2fsprogs udev smartmontools lsscsi nss

COPY ./_build/local-storage /

ENTRYPOINT [ "/local-storage" ]
