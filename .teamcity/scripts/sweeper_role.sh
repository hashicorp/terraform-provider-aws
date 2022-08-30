#!/usr/bin/env bash

set -euo pipefail

touch ./aws.conf
chmod 600 ./aws.conf
echo << EOF > ./aws.conf
[profile sweeper]
role_arn = arn:aws:iam::%env.AWS_ACCOUNT_ID%:role/teamcity-acctest
EOF

AWS_CONFIG_FILE=./aws.conf AWS_PROFILE=sweeper go test ./internal/sweep -v -tags=sweep -sweep="%SWEEPER_REGIONS%" -sweep-allow-failures -timeout=4h
