// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sesv2_dedicated_ip_pool")
func DataSourceDedicatedIPPool() *schema.Resource {
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
	DSNameDedicatedIPPool = "Dedicated IP Pool Data Source"
)

func dataSourceDedicatedIPPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	out, err := FindDedicatedIPPoolByID(ctx, conn, d.Get("pool_name").(string))
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, DSNameDedicatedIPPool, d.Get("pool_name").(string), err)
	}
	poolName := aws.ToString(out.DedicatedIpPool.PoolName)
	d.SetId(poolName)
	d.Set("scaling_mode", string(out.DedicatedIpPool.ScalingMode))
	d.Set(names.AttrARN, poolNameToARN(meta, poolName))

	outIP, err := findDedicatedIPPoolIPs(ctx, conn, poolName)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, DSNameDedicatedIPPool, poolName, err)
	}
	d.Set("dedicated_ips", flattenDedicatedIPs(outIP.DedicatedIps))

	tags, err := listTags(ctx, conn, d.Get(names.AttrARN).(string))
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, DSNameDedicatedIPPool, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set(names.AttrTags, tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, DSNameDedicatedIPPool, d.Id(), err)
	}

	return diags
}

func flattenDedicatedIPs(apiObjects []types.DedicatedIp) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var dedicatedIps []interface{}
	for _, apiObject := range apiObjects {
		ip := map[string]interface{}{
			"ip":                aws.ToString(apiObject.Ip),
			"warmup_percentage": apiObject.WarmupPercentage,
			"warmup_status":     string(apiObject.WarmupStatus),
		}

		dedicatedIps = append(dedicatedIps, ip)
	}

	return dedicatedIps
}

func findDedicatedIPPoolIPs(ctx context.Context, conn *sesv2.Client, poolName string) (*sesv2.GetDedicatedIpsOutput, error) {
	in := &sesv2.GetDedicatedIpsInput{
		PoolName: aws.String(poolName),
	}
	out, err := conn.GetDedicatedIps(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
