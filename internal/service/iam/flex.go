package iam

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
