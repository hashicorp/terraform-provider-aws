#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

mkdir -p tools && wget -O tf.zip https://releases.hashicorp.com/terraform/%TERRAFORM_CORE_VERSION%/terraform_%TERRAFORM_CORE_VERSION%_linux_amd64.zip && unzip tf.zip && mv terraform tools