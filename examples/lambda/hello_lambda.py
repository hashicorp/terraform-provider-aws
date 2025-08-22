# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

import os

def lambda_handler(event, context):
    return "{} from Lambda!".format(os.environ['greeting'])
