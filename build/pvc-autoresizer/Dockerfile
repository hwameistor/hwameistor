FROM rockylinux:8

RUN yum install nss -y

COPY ./_build/pvc-autoresizer /

EXPOSE 80

ENTRYPOINT [ "/pvc-autoresizer" ]