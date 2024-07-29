FROM centos:7

RUN rm -rf /etc/yum.repos.d/*
COPY ./build/CentOS-Base.repo /etc/yum.repos.d/
RUN yum upgrade nss -y

COPY ./_build/dataset-manager /

ENTRYPOINT [ "/dataset-manager" ]
