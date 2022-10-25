FROM centos:7

RUN yum upgrade nss -y

COPY ./_build/scheduler /

ENTRYPOINT [ "/scheduler" ]
