FROM centos:8


RUN yum upgrade nss -y

COPY ./_build/scheduler /

ENTRYPOINT [ "/scheduler" ]
