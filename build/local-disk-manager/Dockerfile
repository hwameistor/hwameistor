FROM rockylinux:8

RUN yum -y --refresh upgrade ca-certificates 'glibc*' \
    && yum install -y xfsprogs smartmontools lsscsi e4fsprogs nss udev \
    && glibc_version=$(rpm -q --qf '%{VERSION}-%{RELEASE}' glibc) \
    && test "$(rpm --eval "%{lua:print(rpm.vercmp('$glibc_version', '2.28-251.el8_10.11'))}")" -ge 0 \
    && yum clean all
COPY ./_build/local-disk-manager /local-disk-manager

ENTRYPOINT [ "/local-disk-manager" ]
