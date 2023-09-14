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
    #perf_main_memalloc1=$( pprof -top -flat -sample_index=alloc_space -unit=mb memvpcmain.prof | head -3 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    perf_main_memalloc1=$( pprof -top -flat -sample_index=alloc_space -unit=mb memvpcmain.prof | head -3 )
    echo "perf_main_memalloc1:$perf_main_memalloc1"
    perf_main_memalloc1=$( echo -n $perf_main_memalloc1 | tr '\n' ' ' )
    echo "perf_main_memalloc1:$perf_main_memalloc1"
    perf_main_memalloc1=$( echo -n $perf_main_memalloc1 | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    echo "perf_main_memalloc1:$perf_main_memalloc1"
    perf_main_meminuse1=$( pprof -top -flat -sample_index=inuse_space -unit=mb memvpcmain.prof | head -3 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    perf_main_cputime1=$( pprof -top -flat -sample_index=cpu cpuvpcmain.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    perf_main_memalloc2=$( pprof -top -flat -sample_index=alloc_space -unit=mb memssmmain.prof | head -3 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    perf_main_meminuse2=$( pprof -top -flat -sample_index=inuse_space -unit=mb memssmmain.prof | head -3 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    perf_main_cputime2=$( pprof -top -flat -sample_index=cpu cpussmmain.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    perf_latest_memalloc1=$( pprof -top -flat -sample_index=alloc_space -unit=mb memvpclatest.prof | head -3 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    perf_latest_meminuse1=$( pprof -top -flat -sample_index=inuse_space -unit=mb memvpclatest.prof | head -3 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    perf_latest_cputime1=$( pprof -top -flat -sample_index=cpu cpuvpclatest.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    perf_latest_memalloc2=$( pprof -top -flat -sample_index=alloc_space -unit=mb memssmlatest.prof | head -3 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    perf_latest_meminuse2=$( pprof -top -flat -sample_index=inuse_space -unit=mb memssmlatest.prof | head -3 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    perf_latest_cputime2=$( pprof -top -flat -sample_index=cpu cpussmlatest.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    alloc=$( bc -l <<< "((${perf_main_memalloc1}/${perf_latest_memalloc1})-1) + ((${perf_main_memalloc2}/${perf_latest_memalloc2})-1)/2" )
    inuse=$( bc -l <<< "((${perf_main_meminuse1}/${perf_latest_meminuse1})-1) + ((${perf_main_meminuse2}/${perf_latest_meminuse2})-1)/2" )
    cputime=$( bc -l <<< "((${perf_main_cputime1}/${perf_latest_cputime1})-1) + ((${perf_main_cputime2}/${perf_latest_cputime2})-1)/2" )

    printf "Alloc:%%.4f%%%%" "${alloc}"
    printf ";Inuse:%%.4f%%%%" "${inuse}"
    printf ";CPUtime:%%.4f%%%%\n" "${cputime}"
}

if [ -f "memvpcmain.prof" -a -f "memssmmain.prof" -a -f "memvpclatest.prof" -a -f "memssmlatest.prof" ]; then
    echo "Tests complete. Analyzing results..."
    analysis
fi

if [ -f "memvpcmain.prof" -a -f "memssmmain.prof" -a -f "memvpclatest.prof" -a ! -f "memssmlatest.prof" ]; then
    echo "Running SSM latest version ($(basename $(curl -Ls -o /dev/null -w %%{url_effective} https://github.com/hashicorp/terraform-provider-aws/releases/latest))) test..."
    ssmtest ssmlatest
fi

if [ -f "memvpcmain.prof" -a -f "memssmmain.prof" -a ! -f "memvpclatest.prof" ]; then
    echo "Running VPC latest version ($(basename $(curl -Ls -o /dev/null -w %%{url_effective} https://github.com/hashicorp/terraform-provider-aws/releases/latest))) test..."
    git checkout $(basename $(curl -Ls -o /dev/null -w %%{url_effective} https://github.com/hashicorp/terraform-provider-aws/releases/latest))
    vpctest vpclatest
fi

if [ -f "memvpcmain.prof" -a ! -f "memssmmain.prof" ]; then
    echo "Running SSM main branch test..."
    ssmtest ssmmain
fi

if [ ! -f "memvpcmain.prof" ]; then
    go install github.com/google/pprof@latest
    echo "Running VPC main branch test..."
    vpctest vpcmain
fi
