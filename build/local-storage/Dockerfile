FROM centos:7

RUN rm -rf /etc/yum.repos.d/* && \
    curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo && \
    yum install -y xfsprogs e2fsprogs udev smartmontools lsscsi && \
    yum upgrade nss -y

COPY ./_build/local-storage /

ENTRYPOINT [ "/local-storage" ]
