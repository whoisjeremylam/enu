#!/bin/bash

. ~/.bashrc

export GOPATH=/home/api/api
export GOROOT=/home/api/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

echo Starting deployment of Enu API
echo GOPATH=$GOPATH
echo GOROOT=$GOROOT
echo PATH=$PATH

# Refresh source code and clean up executables
cd $GOPATH/src/github.com/vennd/enu && git pull
rm $GOPATH/bin/enu
rm $GOPATH/bin/pingcounterpartyd
rm $GOPATH/bin/pingbitcoind

# Rebuild executables
go install github.com/vennd/enu
go install github.com/vennd/enu/monitoring/pingcounterpartyd
go install github.com/vennd/enu/monitoring/pingbitcoind

# Restart processes
cd $GOPATH/bin
launchProcess.sh restart enu

# This doesn't update service monitoring or scripts used for monitoring

# Restore execute permissions on scripts
chmod u+x /home/api/api/src/github.com/vennd/enu/monitoring/pingcounterpartyd/pingcounterpartyd.sh
chmod u+x /home/api/api/src/github.com/vennd/enu/monitoring/pingbitcoind/pingbitcoind.sh
chmod u+x /home/api/api/src/github.com/vennd/enu/deploy.sh
chmod u+x /home/api/api/src/github.com/vennd/enu/monitoring/pingbitcoind/restartAndAlertBitcoin.sh
chmod u+x /home/api/api/src/github.com/vennd/enu/monitoring/pingcounterpartyd/restartAndAlertCounterparty.sh

# Alert
curl -s \
  --form-string "token=ah1GHznmo7xUxqVrcqP59TeD1h6CS8" \
  --form-string "user=gTgS3iypxztM2SgvLYf3JwrSkactFb" \
  --form-string "message=$HOSTNAME: Automated deployment complete" \
  https://api.pushover.net/1/messages.json
