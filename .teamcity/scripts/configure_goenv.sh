#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

go_version="$(goenv local)"
goenv install --skip-existing "${go_version}" && goenv rehash
