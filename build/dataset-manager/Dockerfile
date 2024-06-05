FROM centos:7

RUN yum upgrade nss -y

COPY ./_build/dataset-manager /

ENTRYPOINT [ "/dataset-manager" ]
