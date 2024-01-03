// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !generate
// +build !generate

package servicecatalog

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/aws/aws-sdk-go/service/servicecatalog/servicecatalogiface"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Custom Service Catalog tag service update functions using the same format as generated code.

func productUpdateTags(ctx context.Context, conn servicecatalogiface.ServiceCatalogAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	input := &servicecatalog.UpdateProductInput{
		Id: aws.String(identifier),
	}

	if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.ServiceCatalog); len(removedTags) > 0 {
		input.RemoveTags = aws.StringSlice(removedTags.Keys())
	}

	if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.ServiceCatalog); len(updatedTags) > 0 {
		input.AddTags = Tags(updatedTags)
	}

	_, err := conn.UpdateProductWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("updating tags for Service Catalog Product (%s): %w", identifier, err)
	}

	return nil
}

func recordKeyValueTags(ctx context.Context, tags []*servicecatalog.RecordTag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}
