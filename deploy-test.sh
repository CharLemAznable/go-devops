#!/usr/bin/env bash

targetHost=$1
deployName=go-devops
fast=$2

if [[ "$fast" == "fast" ]]; then
    echo "jump building in fast mode"
else
    echo "rebuilding"
    rm -fr $deployName.linux.bin $deployName.linux.bin.bz2
    ./gobin.sh
    env GOOS=linux GOARCH=amd64 go build -o $deployName.linux.bin
    upx $deployName.linux.bin
fi


rsync -avz --human-readable --progress -e "ssh -p 22" ./$deployName.linux.bin $targetHost:.
#scp ./$deployName.linux.bin.bz2 $targetHost:./
scp ./deploy-agent.sh $targetHost:.

ssh -tt $targetHost "bash -s" << eeooff
mkdir -p ./app/$deployName/
cd ./app/$deployName/
mv -f ~/$deployName.linux.bin .
mv -f ~/deploy-agent.sh .
ps -ef|grep $deployName|grep -v grep|awk '{print \$2}'|xargs -r kill -9
chmod +x ./deploy-agent.sh
./deploy-agent.sh $deployName app01 app/$deployName
./deploy-agent.sh $deployName app@app03 app/$deployName
./deploy-agent.sh $deployName app@test.ceping app/$deployName
./start-go-devops.sh
exit
eeooff