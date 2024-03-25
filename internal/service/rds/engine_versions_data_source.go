// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
)

// @SDKDataSource("aws_rds_engine_versions")
func DataSourceEngineVersions() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEngineVersionsRead,
		Schema: map[string]*schema.Schema{
			"filter": namevaluesfilters.Schema(),

			"versions": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"engine": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"db_parameter_group_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func dataSourceEngineVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeDBEngineVersionsInput{
		ListSupportedCharacterSets: aws.Bool(true),
		ListSupportedTimezones:     aws.Bool(true),
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).RDSFilters()
	}

	log.Printf("[DEBUG] Reading RDS engine versions: %v", input)
	var engineVersions []map[string]interface{}

	err := conn.DescribeDBEngineVersionsPagesWithContext(ctx, input, func(resp *rds.DescribeDBEngineVersionsOutput, lastPage bool) bool {
		for _, engineVersion := range resp.DBEngineVersions {
			if engineVersion == nil {
				continue
			}

			out := make(map[string]interface{})
			out["engine"] = aws.StringValue(engineVersion.Engine)
			out["engine_version"] = aws.StringValue(engineVersion.EngineVersion)
			out["db_parameter_group_family"] = aws.StringValue(engineVersion.DBParameterGroupFamily)
			out["status"] = aws.StringValue(engineVersion.Status)
			engineVersions = append(engineVersions, out)
		}
		return !lastPage
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS engine versions: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	d.Set("versions", engineVersions)

	return diags
}
