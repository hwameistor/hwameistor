FROM centos:7

RUN yum install -y xfsprogs udev smartmontools lsscsi

COPY tools/drbdsetup /usr/local/bin/

COPY ./_build/dce-uds-local-storage-run /

ENTRYPOINT [ "/dce-uds-local-storage-run" ]
