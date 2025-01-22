// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_standards_control", name="Standards Control")
func resourceStandardsControl() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStandardsControlPut,
		ReadWithoutTimeout:   resourceStandardsControlRead,
		UpdateWithoutTimeout: resourceStandardsControlPut,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"control_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"control_status": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.ControlStatus](),
			},
			"control_status_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disabled_reason": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"related_requirements": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"remediation_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"severity_rating": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"standards_control_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"title": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStandardsControlPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	standardsControlARN := d.Get("standards_control_arn").(string)
	input := &securityhub.UpdateStandardsControlInput{
		ControlStatus:       types.ControlStatus(d.Get("control_status").(string)),
		DisabledReason:      aws.String(d.Get("disabled_reason").(string)),
		StandardsControlArn: aws.String(standardsControlARN),
	}

	_, err := conn.UpdateStandardsControl(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Standards Control (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId(standardsControlARN)
	}

	return append(diags, resourceStandardsControlRead(ctx, d, meta)...)
}

func resourceStandardsControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	standardsSubscriptionARN, err := standardsControlARNToStandardsSubscriptionARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	control, err := findStandardsControlByTwoPartKey(ctx, conn, standardsSubscriptionARN, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Standards Control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Standards Control (%s): %s", d.Id(), err)
	}

	d.Set("control_id", control.ControlId)
	d.Set("control_status", control.ControlStatus)
	d.Set("control_status_updated_at", control.ControlStatusUpdatedAt.Format(time.RFC3339))
	d.Set(names.AttrDescription, control.Description)
	d.Set("disabled_reason", control.DisabledReason)
	d.Set("related_requirements", control.RelatedRequirements)
	d.Set("remediation_url", control.RemediationUrl)
	d.Set("severity_rating", control.SeverityRating)
	d.Set("standards_control_arn", control.StandardsControlArn)
	d.Set("title", control.Title)

	return diags
}

// standardsControlARNToStandardsSubscriptionARN converts a security standard control ARN to a subscription ARN.
func standardsControlARNToStandardsSubscriptionARN(inputARN string) (string, error) {
	const (
		resourceSeparator = "/"
		service           = "securityhub"
	)
	parsedARN, err := arn.Parse(inputARN)
	if err != nil {
		return "", fmt.Errorf("parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, service; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	inputResourceParts := strings.Split(parsedARN.Resource, resourceSeparator)

	if actual, expected := len(inputResourceParts), 3; actual < expected {
		return "", fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, inputARN, actual)
	}

	outputResourceParts := append([]string{"subscription"}, inputResourceParts[1:len(inputResourceParts)-1]...)
	outputARN := arn.ARN{
		Partition: parsedARN.Partition,
		Service:   parsedARN.Service,
		Region:    parsedARN.Region,
		AccountID: parsedARN.AccountID,
		Resource:  strings.Join(outputResourceParts, resourceSeparator),
	}.String()

	return outputARN, nil
}

func findStandardsControlByTwoPartKey(ctx context.Context, conn *securityhub.Client, standardsSubscriptionARN, standardsControlARN string) (*types.StandardsControl, error) {
	input := &securityhub.DescribeStandardsControlsInput{
		StandardsSubscriptionArn: aws.String(standardsSubscriptionARN),
	}

	return findStandardsControl(ctx, conn, input, func(v *types.StandardsControl) bool {
		return aws.ToString(v.StandardsControlArn) == standardsControlARN
	})
}

func findStandardsControl(ctx context.Context, conn *securityhub.Client, input *securityhub.DescribeStandardsControlsInput, filter tfslices.Predicate[*types.StandardsControl]) (*types.StandardsControl, error) {
	output, err := findStandardsControls(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findStandardsControls(ctx context.Context, conn *securityhub.Client, input *securityhub.DescribeStandardsControlsInput, filter tfslices.Predicate[*types.StandardsControl]) ([]types.StandardsControl, error) {
	var output []types.StandardsControl

	pages := securityhub.NewDescribeStandardsControlsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Controls {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
