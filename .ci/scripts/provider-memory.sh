#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Runs the provider memory benchmarks across several fresh processes and reports
# aggregated results.
#
# The benchmarks must be measured cold: the provider's schema cache is
# process-global, so each sample is a separate process running a single
# iteration (-benchtime=1x). Statistical significance comes from repeating across
# fresh processes, not from increasing the in-process iteration count. The test
# binary is compiled once and re-run, so the per-sample cost is the benchmark
# itself rather than a rebuild.
#
# Output is environment-aware. The combined benchmark-format results (for
# benchstat and downstream dashboards) and a human-readable summary are always
# written. TeamCity service messages carrying the median of each metric are
# emitted only when running under TeamCity, detected via a non-empty
# TEAMCITY_VERSION. Run the same way locally and in CI:
#
#     .ci/scripts/provider-memory.sh
#
# Environment overrides:
#   GO              Go command to use (default: go)
#   SAMPLES         Number of fresh-process samples per benchmark (default: 10)
#   RESULTS_FILE    Where to write combined benchmark-format output
#                   (default: provider-memory.txt in the current directory)

set -euo pipefail

GO="${GO:-go}"
SAMPLES="${SAMPLES:-10}"
RESULTS_FILE="${RESULTS_FILE:-provider-memory.txt}"

readonly PKG="./internal/provider"
readonly BENCHMARKS=(
    BenchmarkProviderInitialization
    BenchmarkProviderGetProviderSchema
)

# in_teamcity reports whether the script is running inside a TeamCity build step.
function in_teamcity {
    [[ -n "${TEAMCITY_VERSION:-}" ]]
}

tmpdir=$(mktemp -d)
function cleanup {
    rm -rf "${tmpdir}"
}
trap cleanup EXIT

testbin="${tmpdir}/provider.test"

echo "==> Compiling benchmark binary..."
"${GO}" test -c -o "${testbin}" -vet=off "${PKG}"

: >"${RESULTS_FILE}"

for bench in "${BENCHMARKS[@]}"; do
    echo "==> Running ${bench} (${SAMPLES} samples)..."

    # Discard one warm-up run: the first execution of a freshly written binary
    # pays one-off filesystem and code-signing latency that would skew timing.
    "${testbin}" -test.bench="^${bench}\$" -test.benchtime=1x -test.run='^$' >/dev/null 2>&1 || true

    for ((i = 0; i < SAMPLES; i++)); do
        "${testbin}" \
            -test.bench="^${bench}\$" \
            -test.benchtime=1x \
            -test.benchmem \
            -test.run='^$' 2>/dev/null |
            grep -E "^${bench}" >>"${RESULTS_FILE}"
    done
done

echo
echo "==> benchstat"
if command -v benchstat >/dev/null 2>&1; then
    benchstat "${RESULTS_FILE}"
else
    echo "benchstat not found; skipping comparison (raw results in ${RESULTS_FILE})"
fi

# median prints the median of the values read on stdin, one per line. Values are
# integers (bytes or counts); the median of an even-sized sample is the lower of
# the two middle values, which is sufficient for these deterministic metrics.
function median {
    sort -n | awk '{ a[NR] = $1 } END { if (NR == 0) { exit } print a[int((NR + 1) / 2)] }'
}

# The benchmark-format columns are: name iters (value unit)... Each metric is a
# "<value> <unit>" pair, so the unit names index the values that follow them.
# metric_median extracts the value column for a given unit name and prints its
# median across all samples of a benchmark.
function metric_median {
    local bench="$1" unit="$2"
    grep -E "^${bench}" "${RESULTS_FILE}" |
        awk -v unit="${unit}" '{ for (i = 3; i <= NF; i += 2) if ($(i + 1) == unit) print $i }' |
        median
}

# humanize formats a metric value for display only. Byte-valued metrics (units
# ending in "-bytes" or "B/op") are shown in MiB; all other metrics (allocation
# counts, timings) are passed through unchanged. The raw value is always used for
# the results file and TeamCity statistics, never this formatted form.
function humanize {
    local value="$1" unit="$2"
    case "${unit}" in
        *-bytes | B/op)
            awk -v v="${value}" 'BEGIN { printf "%.1f MiB", v / 1048576 }'
            ;;
        *)
            printf '%s' "${value}"
            ;;
    esac
}

echo
echo "==> Median metrics"
teamcity_lines=""
for bench in "${BENCHMARKS[@]}"; do
    for unit in B/op allocs/op heap-alloc-bytes heap-inuse-bytes peak-heap-inuse-bytes; do
        value=$(metric_median "${bench}" "${unit}")
        [[ -z "${value}" ]] && continue

        # Normalize the metric key: lower-case benchmark name without the
        # Benchmark prefix, plus the unit with non-alphanumerics as underscores.
        short=$(printf '%s' "${bench#Benchmark}" | tr '[:upper:]' '[:lower:]')
        key="provider_mem_${short}_$(printf '%s' "${unit}" | tr -c '[:alnum:]' '_')"
        key="${key%_}"

        printf '  %-34s %-22s %12s\n' "${bench}" "${unit}" "$(humanize "${value}" "${unit}")"
        teamcity_lines+="${key} ${value}"$'\n'
    done
done

if in_teamcity; then
    echo
    echo "==> Emitting TeamCity statistics"
    printf '%s' "${teamcity_lines}" | sort | while read -r key value; do
        [[ -z "${key}" ]] && continue
        printf "##teamcity[buildStatisticValue key='%s' value='%s']\n" "${key}" "${value}"
    done
fi

echo
echo "Combined results written to ${RESULTS_FILE}"
