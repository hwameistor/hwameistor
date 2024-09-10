FROM rockylinux:8

RUN yum upgrade nss -y

COPY ./_build/exporter /

ENTRYPOINT [ "/exporter" ]