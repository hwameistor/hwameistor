FROM rockylinux:8

RUN yum update -y ca-certificates \
    && yum install nss -y \
    && yum clean all

COPY ./_build/dataset-manager /

ENTRYPOINT [ "/dataset-manager" ]
