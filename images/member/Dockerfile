FROM centos:7

RUN yum install -y xfsprogs udev smartmontools lsscsi

COPY tools/drbdsetup /usr/local/bin/

COPY ./_build/local-storage-run /

ENTRYPOINT [ "/local-storage-run" ]
