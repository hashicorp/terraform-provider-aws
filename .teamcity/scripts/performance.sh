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
        -v -parallel 2 -count 2 \
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
    local perf_main_memalloc1=$( pprof -top -flat -sample_index=alloc_space -unit=mb memvpcmain.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    local perf_main_meminuse1=$( pprof -top -flat -sample_index=inuse_space -unit=mb memvpcmain.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    local perf_main_cputime1=$( pprof -top -flat -sample_index=cpu cpuvpcmain.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    local perf_main_memalloc2=$( pprof -top -flat -sample_index=alloc_space -unit=mb memssmmain.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    local perf_main_meminuse2=$( pprof -top -flat -sample_index=inuse_space -unit=mb memssmmain.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    local perf_main_cputime2=$( pprof -top -flat -sample_index=cpu cpussmmain.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    local perf_latest_memalloc1=$( pprof -top -flat -sample_index=alloc_space -unit=mb memvpclatest.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    local perf_latest_meminuse1=$( pprof -top -flat -sample_index=inuse_space -unit=mb memvpclatest.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    local perf_latest_cputime1=$( pprof -top -flat -sample_index=cpu cpuvpclatest.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    local perf_latest_memalloc2=$( pprof -top -flat -sample_index=alloc_space -unit=mb memssmlatest.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    local perf_latest_meminuse2=$( pprof -top -flat -sample_index=inuse_space -unit=mb memssmlatest.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)MB total.*/\1/g' )
    local perf_latest_cputime2=$( pprof -top -flat -sample_index=cpu cpussmlatest.prof | head -5 | tr '\n' ' ' | sed -E 's/.*%% of ([0-9.]+)s total.*/\1/g' ) 2>/dev/null

    local alloc=$( bc -l <<< "(((${perf_main_memalloc1}/${perf_latest_memalloc1})-1) + ((${perf_main_memalloc2}/${perf_latest_memalloc2})-1)/2)*100" )
    local inuse=$( bc -l <<< "(((${perf_main_meminuse1}/${perf_latest_meminuse1})-1) + ((${perf_main_meminuse2}/${perf_latest_meminuse2})-1)/2)*100" )
    local cputime=$( bc -l <<< "(((${perf_main_cputime1}/${perf_latest_cputime1})-1) + ((${perf_main_cputime2}/${perf_latest_cputime2})-1)/2)*100" )

    local alloc_bw="Worse"
    if (( $( echo "${alloc} < 0" | bc -l) )); then
        alloc_bw="Better"
    fi

    local inuse_bw="Worse"
    if (( $( echo "${inuse} < 0" | bc -l) )); then
        inuse_bw="Better"
    fi

    local cputime_bw="Worse"
    if (( $( echo "${cputime} < 0" | bc -l) )); then
        cputime_bw="Better"
    fi

    local alloc_emoji=""
    if (( $( echo "${alloc} > 4.99999" | bc -l) )); then
        alloc_emoji=":x:"
    else if (( $( echo "${alloc} < -4.99999" | bc -l) )); then
        alloc_emoji=":white_check_mark:"
    fi

    local cputime_emoji=""
    if (( $( echo "${cputime} > 4.99999" | bc -l) )); then
        cputime_emoji=":x:"
    else if (( $( echo "${cputime} < -4.99999" | bc -l) )); then
        cputime_emoji=":white_check_mark:"
    fi

    printf "##teamcity[notification notifier='slack' message='**Performance changes from latest version (%%s) to main**\nAllocated memory: %%.4f%%%% (%%s) %%s\nIn-use memory: %%.4f%%%% (%%s) (wide-fluctuations normal)\nCPU time: %%.4f%%%% (%%s) %%s' sendTo='CN0G9S7M4' connectionId='PROJECT_EXT_8']\n" "$(basename $(curl -Ls -o /dev/null -w %%{url_effective} https://github.com/hashicorp/terraform-provider-aws/releases/latest))" "${alloc}" "${alloc_bw}" "${alloc_emoji}" "${inuse}" "${inuse_bw}" "${cputime}" "${cputime_bw}" "${cputime_emoji}"
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
