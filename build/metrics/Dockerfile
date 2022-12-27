FROM centos:7

RUN yum upgrade nss -y

COPY ./_build/metrics /

ENTRYPOINT [ "/metrics" ]