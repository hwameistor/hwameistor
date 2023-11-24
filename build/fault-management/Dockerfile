FROM centos:7

RUN yum upgrade nss -y

COPY ./_build/fault-management /

ENTRYPOINT [ "/fault-management" ]