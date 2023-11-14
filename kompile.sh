export PATH=$PATH:/root/go/bin
export PATH=$PATH:/usr/bin/protobuf/bin 
export PROTOC_HOME=/usr/bin/protobuf
export PROTOGENIC=/build/protogenic
eval $(ssh-agent -s) && ssh-add /root/.ssh/gitlab 
git config --global url."git@gitlab.com:".insteadOf "https://gitlab.com/" 
git clone $1 /build
mv /usr/bin/protobuf/bin/kompiler /build 
mv /usr/bin/protobuf/bin/protogenic /build 
cd /build && ./kompiler