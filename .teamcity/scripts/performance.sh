#!/usr/bin/env bash

set -euo pipefail

if [[ -n "%ACCTEST_ROLE_ARN%" ]]; then
    conf=$(pwd)/aws.conf

    function cleanup {
        rm "${conf}"
    }
    trap cleanup EXIT

    touch "${conf}"
    chmod 600 "${conf}"
    cat <<EOF >"${conf}"
[profile memtest]
role_arn       = %ACCTEST_ROLE_ARN%
source_profile = source

[profile source]
aws_access_key_id     = %AWS_ACCESS_KEY_ID%
aws_secret_access_key = %AWS_SECRET_ACCESS_KEY%
EOF

    unset AWS_ACCESS_KEY_ID
    unset AWS_SECRET_ACCESS_KEY

    export AWS_CONFIG_FILE="${conf}"
    export AWS_PROFILE=memtest
fi

go install github.com/google/pprof@latest
function vpctest {
    TF_ACC=1 go test \
        ./internal/service/ec2/... \
        -v -parallel 1 \
        -run='^TestAccVPC_basic$' \
        -cpuprofile cpu.prof \
        -memprofile mem.prof \
        -bench \
        -timeout 60m
}

function ssmtest {
    TF_ACC=1 go test \
        ./internal/service/ssm/... \
        -v -parallel 2 -count 2 \
        -run='^TestAccSSMParameter_basic$' \
        -cpuprofile cpu.prof \
        -memprofile mem.prof \
        -bench \
        -timeout 60m
}

vpctest

perf_main_memalloc1=$( pprof -top -flat -sample_index=alloc_space -unit=mb mem.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
perf_main_meminuse1=$( pprof -top -flat -sample_index=inuse_space -unit=mb mem.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
perf_main_cputime1=$( pprof -top -flat -sample_index=cpu cpu.prof | head -5 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

ssmtest

perf_main_memalloc2=$( pprof -top -flat -sample_index=alloc_space -unit=mb mem.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
perf_main_meminuse2=$( pprof -top -flat -sample_index=inuse_space -unit=mb mem.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
perf_main_cputime2=$( pprof -top -flat -sample_index=cpu cpu.prof | head -5 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

git checkout $(basename $(curl -Ls -o /dev/null -w %{url_effective} https://github.com/hashicorp/terraform-provider-aws/releases/latest))

vpctest

perf_latest_memalloc1=$( pprof -top -flat -sample_index=alloc_space -unit=mb mem.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
perf_latest_meminuse1=$( pprof -top -flat -sample_index=inuse_space -unit=mb mem.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
perf_latest_cputime1=$( pprof -top -flat -sample_index=cpu cpu.prof | head -5 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

ssmtest

perf_latest_memalloc2=$( pprof -top -flat -sample_index=alloc_space -unit=mb mem.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
perf_latest_meminuse2=$( pprof -top -flat -sample_index=inuse_space -unit=mb mem.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
perf_latest_cputime2=$( pprof -top -flat -sample_index=cpu cpu.prof | head -5 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

