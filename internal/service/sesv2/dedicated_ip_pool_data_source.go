// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sesv2_dedicated_ip_pool", name="Dedicated IP Pool")
// @Tags(identifierAttribute="arn")
func dataSourceDedicatedIPPool() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDedicatedIPPoolRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dedicated_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"warmup_percentage": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"warmup_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"pool_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scaling_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	dsNameDedicatedIPPool = "Dedicated IP Pool Data Source"
)

func dataSourceDedicatedIPPoolRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := findDedicatedIPPoolByName(ctx, conn, d.Get("pool_name").(string))
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, dsNameDedicatedIPPool, d.Get("pool_name").(string), err)
	}

	poolName := aws.ToString(out.PoolName)
	d.SetId(poolName)
	d.Set(names.AttrARN, dedicatedIPPoolARN(ctx, meta.(*conns.AWSClient), poolName))
	d.Set("scaling_mode", out.ScalingMode)

	outIP, err := findDedicatedIPsByPoolName(ctx, conn, poolName)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, dsNameDedicatedIPPool, poolName, err)
	}
	d.Set("dedicated_ips", flattenDedicatedIPs(outIP))

	return diags
}

func flattenDedicatedIPs(apiObjects []types.DedicatedIp) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var dedicatedIps []any
	for _, apiObject := range apiObjects {
		ip := map[string]any{
			"ip":                aws.ToString(apiObject.Ip),
			"warmup_percentage": apiObject.WarmupPercentage,
			"warmup_status":     string(apiObject.WarmupStatus),
		}

		dedicatedIps = append(dedicatedIps, ip)
	}

	return dedicatedIps
}

func findDedicatedIPsByPoolName(ctx context.Context, conn *sesv2.Client, poolName string) ([]types.DedicatedIp, error) {
	input := &sesv2.GetDedicatedIpsInput{
		PoolName: aws.String(poolName),
	}

	return findDedicatedIPs(ctx, conn, input)
}

func findDedicatedIPs(ctx context.Context, conn *sesv2.Client, input *sesv2.GetDedicatedIpsInput) ([]types.DedicatedIp, error) {
	var output []types.DedicatedIp

	pages := sesv2.NewGetDedicatedIpsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.DedicatedIps...)
	}

	return output, nil
}
