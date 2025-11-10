// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_bot_association", name="Bot Association")
func dataSourceBotAssociation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBotAssociationRead,

		Schema: map[string]*schema.Schema{
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			"lex_bot": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lex_region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(2, 50),
						},
					},
				},
			},
		},
	}
}

func dataSourceBotAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)

	var name, region string
	if v, ok := d.GetOk("lex_bot"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		lexBot := expandLexBot(v.([]any)[0].(map[string]any))
		name = aws.ToString(lexBot.Name)
		region = aws.ToString(lexBot.LexRegion)
		if region == "" {
			region = meta.(*conns.AWSClient).Region(ctx)
		}
	}

	id := botAssociationCreateResourceID(instanceID, name, region)
	lexBot, err := findBotAssociationByThreePartKey(ctx, conn, instanceID, name, region)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Bot Association (%s): %s", id, err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set(names.AttrInstanceID, instanceID)
	if err := d.Set("lex_bot", flattenLexBot(lexBot)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lex_bot: %s", err)
	}

	return diags
}
