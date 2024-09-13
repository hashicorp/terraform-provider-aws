# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

from datetime import datetime


def lambda_handler(event, context):
    with open("/mnt/efs/test.txt", 'a', encoding='utf-8') as f:
        f.write("{}: hello from lambda\n".format(datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%S%z")))

    with open("/mnt/efs/test.txt", 'r', encoding='utf-8') as f:
        content = f.readlines()

    return ''.join(content)
