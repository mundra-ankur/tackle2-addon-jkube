FROM registry.access.redhat.com/ubi8/go-toolset:1.16.12 as builder
ENV GOPATH=$APP_ROOT
COPY --chown=1001:0 . .
RUN make cmd

FROM registry.access.redhat.com/ubi8/ubi-minimal
USER root
RUN echo -e "[centos8]" \
 "\nname = centos8" \
 "\nbaseurl = http://mirror.centos.org/centos/8-stream/AppStream/x86_64/os/" \
 "\nenabled = 1" \
 "\ngpgcheck = 0" > /etc/yum.repos.d/centos.repo
RUN echo -e "[WandiscoSVN]" \
 "\nname=Wandisco SVN Repo" \
 "\nbaseurl=http://opensource.wandisco.com/centos/6/svn-1.9/RPMS/$basearch/" \
 "\nenabled=1" \
 "\ngpgcheck=0" > /etc/yum.repos.d/wandisco.repo

RUN microdnf -y install \
  openssh-clients \
  unzip \
  wget \
  git \
  subversion \
  maven \
  tar\
&& microdnf -y clean all

RUN wget https://download.java.net/openjdk/jdk17/ri/openjdk-17+35_linux-x64_bin.tar.gz \
 && tar -xvf openjdk-17+35_linux-x64_bin.tar.gz \
 && mv jdk-17 /usr/lib/jvm/ \
 && rm openjdk-17+35_linux-x64_bin.tar.gz

ENV HOME=/working \
    JAVA_HOME="/usr/lib/jvm/jdk-17" \
    JAVA_VENDOR="openjdk" \
    JAVA_VERSION="17"

WORKDIR /working
COPY --from=builder /opt/app-root/src/bin/addon /usr/local/bin/addon
ENTRYPOINT ["/usr/local/bin/addon"]