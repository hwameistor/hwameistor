FROM centos:7

RUN yum install -y xfsprogs smartmontools lsscsi e4fsprogs && \
    yum upgrade nss -y

COPY ./_build/local-disk-manager /local-disk-manager

ENTRYPOINT [ "/local-disk-manager" ]
