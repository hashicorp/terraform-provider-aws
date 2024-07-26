// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_gamelift_alias", name="Alias")
// @Tags(identifierAttribute="arn")
func ResourceAlias() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAliasCreate,
		ReadWithoutTimeout:   resourceAliasRead,
		UpdateWithoutTimeout: resourceAliasUpdate,
		DeleteWithoutTimeout: resourceAliasDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"routing_strategy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fleet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrMessage: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								gamelift.RoutingStrategyTypeSimple,
								gamelift.RoutingStrategyTypeTerminal,
							}, false),
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	rs := expandRoutingStrategy(d.Get("routing_strategy").([]interface{}))
	input := gamelift.CreateAliasInput{
		Name:            aws.String(d.Get(names.AttrName).(string)),
		RoutingStrategy: rs,
		Tags:            getTagsIn(ctx),
	}
	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}
	out, err := conn.CreateAliasWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Alias (%s): %s", d.Get(names.AttrName).(string), err)
	}

	d.SetId(aws.StringValue(out.Alias.AliasId))

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	out, err := conn.DescribeAliasWithContext(ctx, &gamelift.DescribeAliasInput{
		AliasId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
			d.SetId("")
			log.Printf("[WARN] GameLift Alias (%s) not found, removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading GameLift Alias (%s): %s", d.Id(), err)
	}
	a := out.Alias

	arn := aws.StringValue(a.AliasArn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, a.Description)
	d.Set(names.AttrName, a.Name)
	d.Set("routing_strategy", flattenRoutingStrategy(a.RoutingStrategy))

	return diags
}

func resourceAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	log.Printf("[INFO] Updating GameLift Alias: %s", d.Id())
	_, err := conn.UpdateAliasWithContext(ctx, &gamelift.UpdateAliasInput{
		AliasId:         aws.String(d.Id()),
		Name:            aws.String(d.Get(names.AttrName).(string)),
		Description:     aws.String(d.Get(names.AttrDescription).(string)),
		RoutingStrategy: expandRoutingStrategy(d.Get("routing_strategy").([]interface{})),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating GameLift Alias (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAliasRead(ctx, d, meta)...)
}

func resourceAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	log.Printf("[INFO] Deleting GameLift Alias: %s", d.Id())
	if _, err := conn.DeleteAliasWithContext(ctx, &gamelift.DeleteAliasInput{
		AliasId: aws.String(d.Id()),
	}); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Alias (%s): %s", d.Id(), err)
	}
	return diags
}

func expandRoutingStrategy(cfg []interface{}) *gamelift.RoutingStrategy {
	if len(cfg) < 1 {
		return nil
	}

	strategy := cfg[0].(map[string]interface{})

	out := gamelift.RoutingStrategy{
		Type: aws.String(strategy[names.AttrType].(string)),
	}

	if v, ok := strategy["fleet_id"].(string); ok && len(v) > 0 {
		out.FleetId = aws.String(v)
	}
	if v, ok := strategy[names.AttrMessage].(string); ok && len(v) > 0 {
		out.Message = aws.String(v)
	}

	return &out
}

func flattenRoutingStrategy(rs *gamelift.RoutingStrategy) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.FleetId != nil {
		m["fleet_id"] = aws.StringValue(rs.FleetId)
	}
	if rs.Message != nil {
		m[names.AttrMessage] = aws.StringValue(rs.Message)
	}
	m[names.AttrType] = aws.StringValue(rs.Type)

	return []interface{}{m}
}
