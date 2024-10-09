#!/usr/bin/env bash

set -euo pipefail

go_version="$(goenv local)"
goenv install --skip-existing "${go_version}" && goenv rehash
