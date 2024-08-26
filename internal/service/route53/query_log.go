// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_query_log", name="Query Logging Config")
func resourceQueryLog() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueryLogCreate,
		ReadWithoutTimeout:   resourceQueryLogRead,
		DeleteWithoutTimeout: resourceQueryLogDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCloudWatchLogGroupARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceQueryLogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	input := &route53.CreateQueryLoggingConfigInput{
		CloudWatchLogsLogGroupArn: aws.String(d.Get(names.AttrCloudWatchLogGroupARN).(string)),
		HostedZoneId:              aws.String(d.Get("zone_id").(string)),
	}

	output, err := conn.CreateQueryLoggingConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Query Logging Config: %s", err)
	}

	d.SetId(aws.ToString(output.QueryLoggingConfig.Id))

	return append(diags, resourceQueryLogRead(ctx, d, meta)...)
}

func resourceQueryLogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	output, err := findQueryLoggingConfigByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Query Logging Config %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Query Logging Config (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  "queryloggingconfig/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCloudWatchLogGroupARN, output.CloudWatchLogsLogGroupArn)
	d.Set("zone_id", output.HostedZoneId)

	return diags
}

func resourceQueryLogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	log.Printf("[DEBUG] Deleting Route53 Query Logging Config: %s", d.Id())
	_, err := conn.DeleteQueryLoggingConfig(ctx, &route53.DeleteQueryLoggingConfigInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchQueryLoggingConfig](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Query Logging Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findQueryLoggingConfigByID(ctx context.Context, conn *route53.Client, id string) (*awstypes.QueryLoggingConfig, error) {
	input := &route53.GetQueryLoggingConfigInput{
		Id: aws.String(id),
	}

	output, err := conn.GetQueryLoggingConfig(ctx, input)

	if errs.IsA[*awstypes.NoSuchQueryLoggingConfig](err) || errs.IsA[*awstypes.NoSuchHostedZone](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.QueryLoggingConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.QueryLoggingConfig, nil
}
