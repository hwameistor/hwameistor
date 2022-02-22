FROM centos:7

RUN yum install -y xfsprogs udev smartmontools lsscsi

COPY tools/drbdsetup /usr/local/bin/

COPY ./_build/localstorage-local-storage-run /

ENTRYPOINT [ "/localstorage-local-storage-run" ]
