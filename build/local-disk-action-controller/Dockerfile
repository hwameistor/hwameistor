FROM centos:7

RUN yum upgrade nss -y

COPY ./_build/local-disk-action-controller /

ENTRYPOINT [ "/local-disk-action-controller" ]