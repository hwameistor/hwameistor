FROM rockylinux:8

RUN yum install nss -y

COPY ./_build/local-disk-action-controller /

ENTRYPOINT [ "/local-disk-action-controller" ]