#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

# Normalise architecture to the names used by HashiCorp releases
case "${arch}" in
	x86_64)  arch="amd64" ;;
	aarch64) arch="arm64" ;;
	arm64)   arch="arm64" ;;
esac

latest_version="$(curl -fsSL https://api.releases.hashicorp.com/v1/releases/terraform/latest | grep -o '"version":"[^"]*"' | head -1 | cut -d'"' -f4)"

echo "Installing Terraform ${latest_version} (${os}/${arch})"

zip_name="terraform_${latest_version}_${os}_${arch}.zip"
url="https://releases.hashicorp.com/terraform/${latest_version}/${zip_name}"

curl -fsSL -o "/tmp/${zip_name}" "${url}"
unzip -o "/tmp/${zip_name}" -d /usr/local/bin terraform
chmod +x /usr/local/bin/terraform
rm "/tmp/${zip_name}"

echo "Terraform $(terraform version -json | grep -o '"terraform_version":"[^"]*"' | cut -d'"' -f4) installed at $(command -v terraform)"
