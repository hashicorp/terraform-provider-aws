#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

# Install Terraform version specified by the environment variable "TERRAFORM_CORE_VERSION".
# If "TERRAFORM_CORE_VERSION" is not specified, installs the latest version.

version="${TERRAFORM_CORE_VERSION:-}"

if [[ -z "${version}" ]]; then
    echo "TERRAFORM_CORE_VERSION not specified, finding latest version..."
    version=$(curl -fsSL https://api.releases.hashicorp.com/v1/releases/terraform/latest | \
        grep -o '"version":"[^"]*"' | head -1 | cut -d'"' -f4)
fi

echo "Downloading Terraform ${version}..."

tools_dir="%TOOLS_DIR%"
mkdir -p "${tools_dir}"

zip_file=$(mktemp --suffix=.zip)
trap 'rm -f "${zip_file}"' EXIT

wget --no-verbose -O "${zip_file}" \
    "https://releases.hashicorp.com/terraform/${version}/terraform_${version}_linux_amd64.zip"

unzip -o -d "${tools_dir}" "${zip_file}" terraform
