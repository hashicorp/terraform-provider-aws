#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


set -euo pipefail

go_version="$(goenv local)"
goenv install --skip-existing "${go_version}" && goenv rehash
