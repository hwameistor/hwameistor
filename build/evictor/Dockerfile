FROM rockylinux:8

RUN yum install nss -y

COPY ./_build/evictor /

ENTRYPOINT [ "/evictor" ]