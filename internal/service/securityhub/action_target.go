// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_action_target", name="Action Target")
func resourceActionTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceActionTargetCreate,
		ReadWithoutTimeout:   resourceActionTargetRead,
		UpdateWithoutTimeout: resourceActionTargetUpdate,
		DeleteWithoutTimeout: resourceActionTargetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "must contain only alphanumeric characters"),
				),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
				),
			},
		},
	}
}

func resourceActionTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	id := d.Get(names.AttrIdentifier).(string)
	input := &securityhub.CreateActionTargetInput{
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		Id:          aws.String(id),
		Name:        aws.String(d.Get(names.AttrName).(string)),
	}

	output, err := conn.CreateActionTarget(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Action Target (%s): %s", id, err)
	}

	d.SetId(aws.ToString(output.ActionTargetArn))

	return append(diags, resourceActionTargetRead(ctx, d, meta)...)
}

func resourceActionTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	actionTargetIdentifier, err := actionTargetParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findActionTargetByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Action Target %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Action Target (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ActionTargetArn)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrIdentifier, actionTargetIdentifier)
	d.Set(names.AttrName, output.Name)

	return diags
}

func resourceActionTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.UpdateActionTargetInput{
		ActionTargetArn: aws.String(d.Id()),
		Description:     aws.String(d.Get(names.AttrDescription).(string)),
		Name:            aws.String(d.Get(names.AttrName).(string)),
	}

	if _, err := conn.UpdateActionTarget(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Action Target (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceActionTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Action Target: %s", d.Id())
	_, err := conn.DeleteActionTarget(ctx, &securityhub.DeleteActionTargetInput{
		ActionTargetArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Action Target (%s): %s", d.Id(), err)
	}

	return diags
}

func actionTargetParseID(arn string) (string, error) {
	parts := strings.Split(arn, "/")

	if len(parts) != 3 {
		return "", fmt.Errorf("expected Security Hub Custom action ARN, received: %s", arn)
	}

	return parts[2], nil
}

func findActionTargetByARN(ctx context.Context, conn *securityhub.Client, arn string) (*types.ActionTarget, error) {
	input := &securityhub.DescribeActionTargetsInput{
		ActionTargetArns: []string{arn},
	}

	return findActionTarget(ctx, conn, input)
}

func findActionTarget(ctx context.Context, conn *securityhub.Client, input *securityhub.DescribeActionTargetsInput) (*types.ActionTarget, error) {
	output, err := findActionTargets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findActionTargets(ctx context.Context, conn *securityhub.Client, input *securityhub.DescribeActionTargetsInput) ([]types.ActionTarget, error) {
	var output []types.ActionTarget

	pages := securityhub.NewDescribeActionTargetsPaginator(conn, input)
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

		output = append(output, page.ActionTargets...)
	}

	return output, nil
}
