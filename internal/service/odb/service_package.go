//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*odb.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))

	return odb.NewFromConfig(cfg,
		odb.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *odb.Options) {
		},
	), nil
}
