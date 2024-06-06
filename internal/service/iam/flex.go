// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

func expandStringListKeepEmpty(configured []interface{}) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		if val, ok := v.(string); ok {
			vs = append(vs, aws.String(val))
		}
	}
	return vs
}
