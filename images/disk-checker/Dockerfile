FROM centos:7

RUN yum install -y smartmontools lsscsi

COPY ./_build/localstorage-disk-checker-run /

ENTRYPOINT [ "/localstorage-disk-checker-run" ]
