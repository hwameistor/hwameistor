FROM rockylinux:8

RUN yum install nss -y

COPY ./_build/apiserver /

EXPOSE 80

ENTRYPOINT [ "/apiserver" ]