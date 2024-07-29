FROM centos:8


RUN yum upgrade nss -y

COPY ./_build/pvc-autoresizer /

EXPOSE 80

ENTRYPOINT [ "/pvc-autoresizer" ]