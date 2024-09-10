FROM rockylinux:8

RUN yum upgrade nss -y

COPY ./_build/apiserver /

EXPOSE 80

ENTRYPOINT [ "/apiserver" ]