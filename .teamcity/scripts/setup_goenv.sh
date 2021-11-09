#!/usr/bin/env bash

set -euo pipefail

goenv install -s "$(goenv local)" && goenv rehash
