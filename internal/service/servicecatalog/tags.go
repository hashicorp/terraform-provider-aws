// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !generate
// +build !generate

package servicecatalog

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

func recordKeyValueTags(ctx context.Context, tags []awstypes.RecordTag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.ToString(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}

// ListTags lists iam service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier, resourceType string) error {
	var (
		tags tftags.KeyValueTags
		err  error
	)
	switch resourceType {
	case "Portfolio":
		tags, err = portfolioKeyValueTags(ctx, meta.(*conns.AWSClient).ServiceCatalogClient(ctx), identifier)

	case "Product":
		tags, err = productKeyValueTags(ctx, meta.(*conns.AWSClient).ServiceCatalogClient(ctx), identifier)

	case "Provisioned Product":
		tags, err = provisionedProductKeyValueTags(ctx, meta.(*conns.AWSClient).ServiceCatalogClient(ctx), identifier)

	default:
		return nil
	}

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}

func portfolioKeyValueTags(ctx context.Context, conn *servicecatalog.Client, identifier string) (tftags.KeyValueTags, error) {
	output, err := findPortfolioByID(ctx, conn, identifier)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	return keyValueTags(ctx, output.Tags), nil
}

func productKeyValueTags(ctx context.Context, conn *servicecatalog.Client, identifier string) (tftags.KeyValueTags, error) {
	output, err := findProductByID(ctx, conn, identifier)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	return keyValueTags(ctx, output.Tags), nil
}

func provisionedProductKeyValueTags(ctx context.Context, conn *servicecatalog.Client, identifier string) (tftags.KeyValueTags, error) {
	input := &servicecatalog.DescribeProvisionedProductInput{
		Id: aws.String(identifier),
	}
	output, err := conn.DescribeProvisionedProduct(ctx, input)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	detail := output.ProvisionedProductDetail

	recordInput := &servicecatalog.DescribeRecordInput{
		Id: detail.LastProvisioningRecordId,
	}

	recordOutput, err := conn.DescribeRecord(ctx, recordInput)
	if err != nil {
		return tftags.New(ctx, nil), fmt.Errorf("listing tags for resource (%s): %w", identifier, err)
	}

	return recordKeyValueTags(ctx, recordOutput.RecordDetail.RecordTags), nil
}
