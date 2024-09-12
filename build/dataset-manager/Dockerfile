FROM rockylinux:8

RUN yum install nss -y

COPY ./_build/dataset-manager /

ENTRYPOINT [ "/dataset-manager" ]
