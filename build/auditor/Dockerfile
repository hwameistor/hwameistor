FROM centos:7

RUN rm -rf /etc/yum.repos.d/* && \
    curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo && \
    yum upgrade nss -y

COPY ./_build/auditor /

ENTRYPOINT [ "/auditor" ]