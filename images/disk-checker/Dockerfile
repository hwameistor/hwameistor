FROM centos:7

RUN yum install -y smartmontools lsscsi

COPY ./_build/dce-uds-disk-checker-run /

ENTRYPOINT [ "/dce-uds-disk-checker-run" ]
