// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package logs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_subscription_filter", name="Subscription Filter")
// @IdentityAttribute("log_group_name")
// @IdentityAttribute("name")
// @ImportIDHandler("subscriptionFilterImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types;awstypes;awstypes.SubscriptionFilter")
// @Testing(importStateIdFunc=testAccSubscriptionFilterImportStateIDFunc)
// @Testing(importIgnore="apply_on_transformed_logs")
// @Testing(plannableImportAction="NoOp")
// @Testing(preIdentityVersion="v6.40.0")
func resourceSubscriptionFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSubscriptionFilterPut,
		ReadWithoutTimeout:   resourceSubscriptionFilterRead,
		UpdateWithoutTimeout: resourceSubscriptionFilterPut,
		DeleteWithoutTimeout: resourceSubscriptionFilterDelete,

		Schema: map[string]*schema.Schema{
			"apply_on_transformed_logs": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrDestinationARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"distribution": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DistributionByLogStream,
				ValidateDiagFunc: enum.Validate[awstypes.Distribution](),
			},
			"emit_system_fields": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"@aws.account", "@aws.region"}, false),
				},
			},
			"filter_pattern": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			names.AttrLogGroupName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceSubscriptionFilterPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	logGroupName := d.Get(names.AttrLogGroupName).(string)
	name := d.Get(names.AttrName).(string)
	input := cloudwatchlogs.PutSubscriptionFilterInput{
		DestinationArn: aws.String(d.Get(names.AttrDestinationARN).(string)),
		FilterName:     aws.String(name),
		FilterPattern:  aws.String(d.Get("filter_pattern").(string)),
		LogGroupName:   aws.String(logGroupName),
	}

	if v, ok := d.GetOk("apply_on_transformed_logs"); ok {
		input.ApplyOnTransformedLogs = v.(bool)
	}

	if v, ok := d.GetOk("distribution"); ok {
		input.Distribution = awstypes.Distribution(v.(string))
	}

	if v, ok := d.GetOk("emit_system_fields"); ok {
		input.EmitSystemFields = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	}

	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhen(ctx, timeout,
		func(ctx context.Context) (any, error) {
			return conn.PutSubscriptionFilter(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Could not deliver test message to specified") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Could not execute the lambda function") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.OperationAbortedException](err, "Please try again") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Make sure you have given CloudWatch Logs permission to assume the provided role") {
				return true, err
			}

			return false, err
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CloudWatch Logs Subscription Filter (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(subscriptionFilterCreateResourceID(logGroupName))
	}

	return diags
}

func resourceSubscriptionFilterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	subscriptionFilter, err := findSubscriptionFilterByTwoPartKey(ctx, conn, d.Get(names.AttrLogGroupName).(string), d.Get(names.AttrName).(string))

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Subscription Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Subscription Filter (%s): %s", d.Id(), err)
	}

	if err := resourceSubscriptionFilterFlatten(ctx, subscriptionFilter, d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceSubscriptionFilterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	log.Printf("[INFO] Deleting CloudWatch Logs Subscription Filter: %s", d.Id())
	input := cloudwatchlogs.DeleteSubscriptionFilterInput{
		FilterName:   aws.String(d.Get(names.AttrName).(string)),
		LogGroupName: aws.String(d.Get(names.AttrLogGroupName).(string)),
	}
	_, err := conn.DeleteSubscriptionFilter(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Subscription Filter (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceSubscriptionFilterFlatten(_ context.Context, subscriptionFilter *awstypes.SubscriptionFilter, d *schema.ResourceData) error { //nolint:unparam
	d.Set("apply_on_transformed_logs", subscriptionFilter.ApplyOnTransformedLogs)
	d.Set(names.AttrDestinationARN, subscriptionFilter.DestinationArn)
	d.Set("distribution", subscriptionFilter.Distribution)
	d.Set("emit_system_fields", subscriptionFilter.EmitSystemFields)
	d.Set("filter_pattern", subscriptionFilter.FilterPattern)
	d.Set(names.AttrLogGroupName, subscriptionFilter.LogGroupName)
	d.Set(names.AttrName, subscriptionFilter.FilterName)
	d.Set(names.AttrRoleARN, subscriptionFilter.RoleArn)

	return nil
}

func subscriptionFilterCreateResourceID(logGroupName string) string {
	var buf bytes.Buffer

	// Each log group can have up to two subscription filters associated with it.
	// However, having multiple resources with the same 'id' attribute valuse is OK.
	fmt.Fprintf(&buf, "%s-", logGroupName)

	return fmt.Sprintf("cwlsf-%d", create.StringHashcode(buf.String()))
}

func findSubscriptionFilterByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.Client, logGroupName, name string) (*awstypes.SubscriptionFilter, error) {
	input := cloudwatchlogs.DescribeSubscriptionFiltersInput{
		FilterNamePrefix: aws.String(name),
		LogGroupName:     aws.String(logGroupName),
	}

	return findSubscriptionFilter(ctx, conn, &input, func(v *awstypes.SubscriptionFilter) bool {
		return aws.ToString(v.FilterName) == name
	})
}

func findSubscriptionFilter(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeSubscriptionFiltersInput, filter tfslices.Predicate[*awstypes.SubscriptionFilter]) (*awstypes.SubscriptionFilter, error) {
	output, err := findSubscriptionFilters(ctx, conn, input, filter, tfslices.WithReturnFirstMatch)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSubscriptionFilters(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeSubscriptionFiltersInput, filter tfslices.Predicate[*awstypes.SubscriptionFilter], optFns ...tfslices.FinderOptionsFunc) ([]awstypes.SubscriptionFilter, error) {
	var output []awstypes.SubscriptionFilter
	opts := tfslices.NewFinderOptions(optFns)

	pages := cloudwatchlogs.NewDescribeSubscriptionFiltersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.SubscriptionFilters {
			if filter(&v) {
				output = append(output, v)
				if opts.ReturnFirstMatch() {
					return output, nil
				}
			}
		}
	}

	return output, nil
}

const subscriptionFilterImportIDSeparator = "|"

func subscriptionFilterParseImportID(id string) (string, string, error) {
	parts := strings.Split(id, subscriptionFilterImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected log-group-name%[2]sfilter-name", id, subscriptionFilterImportIDSeparator)
}

var (
	_ inttypes.SDKv2ImportID = subscriptionFilterImportID{}
)

type subscriptionFilterImportID struct{}

func (subscriptionFilterImportID) Parse(id string) (string, map[string]any, error) {
	logGroupName, filterName, err := subscriptionFilterParseImportID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		names.AttrLogGroupName: logGroupName,
		names.AttrName:         filterName,
	}

	return subscriptionFilterCreateResourceID(logGroupName), result, nil
}

func (subscriptionFilterImportID) Create(d *schema.ResourceData) string {
	return subscriptionFilterCreateResourceID(d.Get(names.AttrLogGroupName).(string))
}
