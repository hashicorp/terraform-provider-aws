// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_db_instances")
func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,

		Schema: map[string]*schema.Schema{
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
			"filter": namevaluesfilters.Schema(),
		},
	}
}

const (
	DSNameInstances = "Instances Data Source"
)

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeDBInstancesInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).RDSFilters()
	}

	var instanceARNS []string
	var instanceIdentifiers []string

	err := conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, dbInstance := range page.DBInstances {
			if dbInstance == nil {
				continue
			}

			instanceARNS = append(instanceARNS, aws.StringValue(dbInstance.DBInstanceArn))
			instanceIdentifiers = append(instanceIdentifiers, aws.StringValue(dbInstance.DBInstanceIdentifier))
		}

		return !lastPage
	})
	if err != nil {
		return create.DiagError(names.RDS, create.ErrActionReading, DSNameInstances, "", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("instance_arns", instanceARNS)
	d.Set("instance_identifiers", instanceIdentifiers)

	return nil
}
