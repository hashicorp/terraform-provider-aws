// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !generate
// +build !generate

package servicecatalog

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// Custom Service Catalog tag service update functions using the same format as generated code.

func recordKeyValueTags(ctx context.Context, tags []*servicecatalog.RecordTag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}
