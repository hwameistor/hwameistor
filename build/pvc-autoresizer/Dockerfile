FROM centos:7

RUN rm -rf /etc/yum.repos.d/*
COPY ./build/CentOS-Base.repo /etc/yum.repos.d/
RUN yum upgrade nss -y

COPY ./_build/pvc-autoresizer /

EXPOSE 80

ENTRYPOINT [ "/pvc-autoresizer" ]