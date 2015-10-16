#!/bin/bash

# Alert
curl -s \
  --form-string "token=ah1GHznmo7xUxqVrcqP59TeD1h6CS8" \
  --form-string "user=gTgS3iypxztM2SgvLYf3JwrSkactFb" \
  --form-string "message=$MONIT_HOST: Warning! Bitcoind has failed or fallen behind in processing blocks. Please check bitcoind" \
  https://api.pushover.net/1/messages.json
