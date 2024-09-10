FROM rockylinux:8

RUN yum upgrade nss -y

COPY ./_build/auditor /

ENTRYPOINT [ "/auditor" ]