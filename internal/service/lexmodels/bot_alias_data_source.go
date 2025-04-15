// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lex_bot_alias", name="Bot Alias")
func dataSourceBotAlias() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBotAliasRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bot_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validBotName,
			},
			"bot_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLastUpdatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validBotAliasName,
			},
		},
	}
}

func dataSourceBotAliasRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LexModelsClient(ctx)

	botName := d.Get("bot_name").(string)
	botAliasName := d.Get(names.AttrName).(string)
	d.SetId(fmt.Sprintf("%s:%s", botName, botAliasName))

	resp, err := conn.GetBotAlias(ctx, &lexmodelbuildingservice.GetBotAliasInput{
		BotName: aws.String(botName),
		Name:    aws.String(botAliasName),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lex bot alias (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Service:   "lex",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("bot:%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	d.Set("bot_name", resp.BotName)
	d.Set("bot_version", resp.BotVersion)
	d.Set("checksum", resp.Checksum)
	d.Set(names.AttrCreatedDate, resp.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, resp.Description)
	d.Set(names.AttrLastUpdatedDate, resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set(names.AttrName, resp.Name)

	return diags
}
