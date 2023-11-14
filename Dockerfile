FROM rockylinux:9
RUN dnf update -y 
RUN dnf install git go -y
RUN dnf install unzip -y
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip
RUN mkdir -p /usr/bin/protobuf
RUN unzip protoc-3.15.8-linux-x86_64.zip -d /usr/bin/protobuf
COPY kompiler /usr/bin/protobuf/bin
COPY protogenic /usr/bin/protobuf/bin
COPY kompile.sh ./
RUN chmod 777 kompile.sh
RUN mkdir /root/.ssh
COPY gitlab /root/.ssh/gitlab
RUN chmod 400 /root/.ssh/gitlab
RUN chmod 777 /usr/bin/protobuf/bin/protoc
RUN chmod 777 /usr/bin/protobuf/bin/kompiler
RUN chmod 777 /usr/bin/protobuf/bin/protogenic
RUN echo -e "Host *\n\tStrictHostKeyChecking no\n\n" > /root/.ssh/config
RUN mkdir /build
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
