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
	case "Instance":
		tags, err = instanceListTags(ctx, meta.(*conns.AWSClient).LightsailClient(ctx), identifier)

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

func instanceListTags(ctx context.Context, client *lightsail.Client, id string) (tftags.KeyValueTags, error) {
	out, err := FindInstanceById(ctx, client, id)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, out.Tags), nil
}
