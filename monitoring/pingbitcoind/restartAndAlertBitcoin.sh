#!/bin/bash

sudo sv restart bitcoin

# Alert
curl -s \
  --form-string "token=ah1GHznmo7xUxqVrcqP59TeD1h6CS8" \
  --form-string "user=gTgS3iypxztM2SgvLYf3JwrSkactFb" \
  --form-string "message=$MONIT_HOST: Bitcoind was down and has been restarted" \
  https://api.pushover.net/1/messages.json
