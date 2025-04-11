// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_db_instances", name="DB Instances")
func dataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,

		Schema: map[string]*schema.Schema{
			names.AttrFilter: namevaluesfilters.Schema(),
			"instance_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"instance_identifiers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeDBInstancesInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).RDSFilters()
	}

	filter := tfslices.PredicateTrue[*types.DBInstance]()
	if v, ok := d.GetOk(names.AttrTags); ok {
		filter = func(x *types.DBInstance) bool {
			return keyValueTags(ctx, x.TagList).ContainsAll(tftags.New(ctx, v.(map[string]any)))
		}
	}

	instances, err := findDBInstances(ctx, conn, input, filter)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Instances: %s", err)
	}

	var instanceARNS []string
	var instanceIdentifiers []string

	for _, instance := range instances {
		instanceARNS = append(instanceARNS, aws.ToString(instance.DBInstanceArn))
		instanceIdentifiers = append(instanceIdentifiers, aws.ToString(instance.DBInstanceIdentifier))
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set("instance_arns", instanceARNS)
	d.Set("instance_identifiers", instanceIdentifiers)

	return diags
}
