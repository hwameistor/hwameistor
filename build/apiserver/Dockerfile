FROM centos:7

RUN yum upgrade nss -y

COPY ./_build/apiserver /

EXPOSE 80

ENTRYPOINT [ "/apiserver" ]