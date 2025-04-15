// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandFieldToMatch(d map[string]any) *awstypes.FieldToMatch {
	ftm := &awstypes.FieldToMatch{
		Type: awstypes.MatchFieldType(d[names.AttrType].(string)),
	}
	if data, ok := d["data"].(string); ok && data != "" {
		ftm.Data = aws.String(data)
	}
	return ftm
}

func flattenFieldToMatch(fm *awstypes.FieldToMatch) []any {
	m := make(map[string]any)
	if fm.Data != nil {
		m["data"] = aws.ToString(fm.Data)
	}

	m[names.AttrType] = string(fm.Type)

	return []any{m}
}
