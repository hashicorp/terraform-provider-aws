#!/bin/bash -v
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

apt-get update -y
apt-get install -y nginx > /tmp/nginx.log

