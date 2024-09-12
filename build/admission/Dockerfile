FROM rockylinux:8

RUN yum install nss -y

COPY ./_build/admission /

ENTRYPOINT [ "/admission" ]