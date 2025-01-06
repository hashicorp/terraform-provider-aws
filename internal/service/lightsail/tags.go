// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !generate
// +build !generate

package lightsail

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier, resourceType string) error {
	var (
		tags tftags.KeyValueTags
		err  error
	)
	switch resourceType {
	case "Bucket":
		tags, err = bucketListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

	case "Certificate":
		tags, err = certificateListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

	case "ContainerService":
		tags, err = containerServiceListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

	case "Database":
		tags, err = databaseListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

	case "Disk":
		tags, err = diskListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

	case "Distribution":
		tags, err = distributionListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

	case "Instance":
		tags, err = instanceListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

	case "KeyPair":
		tags, err = keyPairListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

	case "LB":
		tags, err = lbListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

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

func bucketListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindBucketById(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}

func certificateListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindCertificateById(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}

func containerServiceListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindContainerServiceByName(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}

func databaseListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindDatabaseById(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}

func diskListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindDiskById(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}

func distributionListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindDistributionByID(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}

func instanceListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindInstanceById(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}

func keyPairListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindKeyPairById(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}

func lbListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindLoadBalancerById(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}
