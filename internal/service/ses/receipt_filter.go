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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ses_receipt_filter", name="Receipt Filter")
func resourceReceiptFilter() *schema.Resource {
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
			"cidr": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.IsCIDR,
					validation.IsIPv4Address,
				),
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
			names.AttrPolicy: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ReceiptFilterPolicy](),
			},
		},
	}
}

func resourceReceiptFilterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &ses.CreateReceiptFilterInput{
		Filter: &awstypes.ReceiptFilter{
			IpFilter: &awstypes.ReceiptIpFilter{
				Cidr:   aws.String(d.Get("cidr").(string)),
				Policy: awstypes.ReceiptFilterPolicy(d.Get(names.AttrPolicy).(string)),
			},
			Name: aws.String(name),
		},
	}

	_, err := conn.CreateReceiptFilter(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SES Receipt Filter (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceReceiptFilterRead(ctx, d, meta)...)
}

func resourceReceiptFilterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	filter, err := findReceiptFilterByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SES Receipt Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SES Receipt Filter (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("receipt-filter/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("cidr", filter.IpFilter.Cidr)
	d.Set(names.AttrPolicy, filter.IpFilter.Policy)
	d.Set(names.AttrName, filter.Name)

	return diags
}

func resourceReceiptFilterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESClient(ctx)

	log.Printf("[DEBUG] Deleting SES Receipt Filter: %s", d.Id())
	_, err := conn.DeleteReceiptFilter(ctx, &ses.DeleteReceiptFilterInput{
		FilterName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SES Receipt Filter (%s): %s", d.Id(), err)
	}

	return diags
}

func findReceiptFilterByName(ctx context.Context, conn *ses.Client, name string) (*awstypes.ReceiptFilter, error) {
	input := &ses.ListReceiptFiltersInput{}

	return findReceiptFilter(ctx, conn, input, func(v *awstypes.ReceiptFilter) bool {
		return aws.ToString(v.Name) == name
	})
}

func findReceiptFilter(ctx context.Context, conn *ses.Client, input *ses.ListReceiptFiltersInput, filter tfslices.Predicate[*awstypes.ReceiptFilter]) (*awstypes.ReceiptFilter, error) {
	output, err := findReceiptFilters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReceiptFilters(ctx context.Context, conn *ses.Client, input *ses.ListReceiptFiltersInput, filter tfslices.Predicate[*awstypes.ReceiptFilter]) ([]awstypes.ReceiptFilter, error) {
	output, err := conn.ListReceiptFilters(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.Filters, tfslices.PredicateValue(filter)), nil
}
