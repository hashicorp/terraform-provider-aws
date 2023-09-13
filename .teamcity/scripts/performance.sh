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
[profile perftest]
role_arn       = %ACCTEST_ROLE_ARN%
source_profile = source

[profile source]
aws_access_key_id     = %AWS_ACCESS_KEY_ID%
aws_secret_access_key = %AWS_SECRET_ACCESS_KEY%
EOF

    unset AWS_ACCESS_KEY_ID
    unset AWS_SECRET_ACCESS_KEY

    export AWS_CONFIG_FILE="${conf}"
    export AWS_PROFILE=perftest
fi

go install github.com/google/pprof@latest
function vpctest {
    local suffix=$1
    TF_ACC=1 go test \
        ./internal/service/ec2/... \
        -v -parallel 1 \
        -run='^TestAccVPC_basic$' \
        -cpuprofile cpu"${suffix}".prof \
        -memprofile mem"${suffix}".prof \
        -bench \
        -timeout 60m
}

function ssmtest {
    local suffix=$1
    TF_ACC=1 go test \
        ./internal/service/ssm/... \
        -v -parallel 2 -count 2 \
        -run='^TestAccSSMParameter_basic$' \
        -cpuprofile cpu"${suffix}".prof \
        -memprofile mem"${suffix}".prof \
        -bench \
        -timeout 60m
}

function analysis {
    perf_main_memalloc1=$( pprof -top -flat -sample_index=alloc_space -unit=mb memvpc1.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
    perf_main_meminuse1=$( pprof -top -flat -sample_index=inuse_space -unit=mb memvpc1.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
    perf_main_cputime1=$( pprof -top -flat -sample_index=cpu cpuvpc1.prof | head -5 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    perf_main_memalloc2=$( pprof -top -flat -sample_index=alloc_space -unit=mb memssm1.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
    perf_main_meminuse2=$( pprof -top -flat -sample_index=inuse_space -unit=mb memssm1.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
    perf_main_cputime2=$( pprof -top -flat -sample_index=cpu cpussm1.prof | head -5 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    perf_latest_memalloc1=$( pprof -top -flat -sample_index=alloc_space -unit=mb memvpc2.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
    perf_latest_meminuse1=$( pprof -top -flat -sample_index=inuse_space -unit=mb memvpc2.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
    perf_latest_cputime1=$( pprof -top -flat -sample_index=cpu cpuvpc2.prof | head -5 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    perf_latest_memalloc2=$( pprof -top -flat -sample_index=alloc_space -unit=mb memssm2.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
    perf_latest_meminuse2=$( pprof -top -flat -sample_index=inuse_space -unit=mb memssm2.prof | head -3 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)MB total.*/\1/g' )
    perf_latest_cputime2=$( pprof -top -flat -sample_index=cpu cpussm2.prof | head -5 | tr '\n' ' ' | sed -E 's/.*% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    alloc=$( bc <<< "((${perf_latest_memalloc1}/${perf_main_memalloc1})-1) + ((${perf_latest_memalloc2}/${perf_main_memalloc2})-1)/2" )
    inuse=$( bc <<< "((${perf_latest_meminuse1}/${perf_main_meminuse1})-1) + ((${perf_latest_meminuse2}/${perf_main_meminuse2})-1)/2" )
    cputime=$( bc <<< "((${perf_latest_cputime1}/${perf_main_cputime1})-1) + ((${perf_latest_cputime2}/${perf_main_cputime2})-1)/2" )

    printf "Alloc:$%.4f%" "${alloc}"
    printf ";Inuse:$%.4f%" "${inuse}"
    printf ";CPUtime:$%.4f%\n" "${cputime}"
}

if [ -f "memvpc1.prof" -a -f "memssm1.prof" -a -f "memvpc2.prof" -a ! -f "memssm2.prof" ]; then
    echo "SSM 2 test not yet run. Running..."
    ssmtest ssm2
fi

if [ -f "memvpc1.prof" -a -f "memssm1.prof" -a ! -f "memvpc2.prof" ]; then
    echo "VPC 2 test not yet run. Running..."
    git checkout $(basename $(curl -Ls -o /dev/null -w %{url_effective} https://github.com/hashicorp/terraform-provider-aws/releases/latest))
    vpctest vpc2
fi

if [ -f "memvpc1.prof" -a ! -f "memssm1.prof" ]; then
    echo "SSM 1 test not yet run. Running..."
    ssmtest ssm1
fi

if [ ! -f "memvpc1.prof" ]; then
    echo "VPC 1 test not yet run. Running..."
    vpctest vpc1
fi

if [ -f "memvpc1.prof" -a -f "memssm1.prof" -a -f "memvpc2.prof" -a -f "memssm2.prof" ]; then
    echo "Tests complete. Analyzing results..."
    analysis
fi
