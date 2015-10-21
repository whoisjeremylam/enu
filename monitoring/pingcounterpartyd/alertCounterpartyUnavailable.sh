#!/bin/bash

# Alert
curl -s \
  --form-string "token=ah1GHznmo7xUxqVrcqP59TeD1h6CS8" \
  --form-string "user=gTgS3iypxztM2SgvLYf3JwrSkactFb" \
  --form-string "message=$MONIT_HOST: Counterpartyd service has been temporarily unavailable for more than 30 minutes. Please check the service." \
  https://api.pushover.net/1/messages.json
