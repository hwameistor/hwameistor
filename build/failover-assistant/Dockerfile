FROM centos:8


RUN yum upgrade nss -y

COPY ./_build/failover-assistant /

ENTRYPOINT [ "/failover-assistant" ]