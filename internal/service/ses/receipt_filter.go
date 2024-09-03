// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_receipt_filter")
func ResourceReceiptFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReceiptFilterCreate,
		ReadWithoutTimeout:   resourceReceiptFilterRead,
		DeleteWithoutTimeout: resourceReceiptFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must contain only alphanumeric, period, underscore, and hyphen characters"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]`), "must begin with a alphanumeric character"),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z]$`), "must end with a alphanumeric character"),
				),
			},

			"cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.IsCIDR,
					validation.IsIPv4Address,
				),
			},

			names.AttrPolicy: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ReceiptFilterPolicy](),
			},
		},
	}
}

func resourceReceiptFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	name := d.Get(names.AttrName).(string)

	createOpts := &ses.CreateReceiptFilterInput{
		Filter: &awstypes.ReceiptFilter{
			Name: aws.String(name),
			IpFilter: &awstypes.ReceiptIpFilter{
				Cidr:   aws.String(d.Get("cidr").(string)),
				Policy: awstypes.ReceiptFilterPolicy(d.Get(names.AttrPolicy).(string)),
			},
		},
	}

	_, err := conn.CreateReceiptFilter(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES receipt filter: %s", err)
	}

	d.SetId(name)

	return append(diags, resourceReceiptFilterRead(ctx, d, meta)...)
}

func resourceReceiptFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	listOpts := &ses.ListReceiptFiltersInput{}

	response, err := conn.ListReceiptFilters(ctx, listOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Receipt Filter (%s): %s", d.Id(), err)
	}

	var filter *awstypes.ReceiptFilter

	for _, responseFilter := range response.Filters {
		if aws.ToString(responseFilter.Name) == d.Id() {
			rf := responseFilter
			filter = &rf
			break
		}
	}

	if filter == nil {
		log.Printf("[WARN] SES Receipt Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("cidr", filter.IpFilter.Cidr)
	d.Set(names.AttrPolicy, filter.IpFilter.Policy)
	d.Set(names.AttrName, filter.Name)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("receipt-filter/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)

	return diags
}

func resourceReceiptFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	deleteOpts := &ses.DeleteReceiptFilterInput{
		FilterName: aws.String(d.Id()),
	}

	_, err := conn.DeleteReceiptFilter(ctx, deleteOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES receipt filter: %s", err)
	}

	return diags
}
