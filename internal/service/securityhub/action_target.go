// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_securityhub_action_target")
func ResourceActionTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceActionTargetCreate,
		ReadWithoutTimeout:   resourceActionTargetRead,
		UpdateWithoutTimeout: resourceActionTargetUpdate,
		DeleteWithoutTimeout: resourceActionTargetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), "must contain only alphanumeric characters"),
				),
			},
			"name": {
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
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	identifier := d.Get("identifier").(string)

	log.Printf("[DEBUG] Creating Security Hub Action Target %s", identifier)

	resp, err := conn.CreateActionTargetWithContext(ctx, &securityhub.CreateActionTargetInput{
		Description: aws.String(description),
		Id:          aws.String(identifier),
		Name:        aws.String(name),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Action Target %s: %s", identifier, err)
	}

	d.SetId(aws.StringValue(resp.ActionTargetArn))

	return append(diags, resourceActionTargetRead(ctx, d, meta)...)
}

func resourceActionTargetParseIdentifier(identifier string) (string, error) {
	parts := strings.Split(identifier, "/")

	if len(parts) != 3 {
		return "", fmt.Errorf("Expected Security Hub Custom action ARN, received: %s", identifier)
	}

	return parts[2], nil
}

func resourceActionTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	log.Printf("[DEBUG] Reading Security Hub Action Targets to find %s", d.Id())

	actionTargetIdentifier, err := resourceActionTargetParseIdentifier(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Action Targets (%s): %s", d.Id(), err)
	}

	actionTarget, err := ActionTargetCheckExists(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Action Targets (%s): %s", d.Id(), err)
	}

	if actionTarget == nil {
		log.Printf("[WARN] Security Hub Action Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("identifier", actionTargetIdentifier)
	d.Set("description", actionTarget.Description)
	d.Set("arn", actionTarget.ActionTargetArn)
	d.Set("name", actionTarget.Name)

	return diags
}

func resourceActionTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)

	input := &securityhub.UpdateActionTargetInput{
		ActionTargetArn: aws.String(d.Id()),
		Description:     aws.String(d.Get("description").(string)),
		Name:            aws.String(d.Get("name").(string)),
	}
	if _, err := conn.UpdateActionTargetWithContext(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Action Target (%s): %s", d.Id(), err)
	}
	return diags
}

func ActionTargetCheckExists(ctx context.Context, conn *securityhub.SecurityHub, actionTargetArn string) (*securityhub.ActionTarget, error) {
	input := &securityhub.DescribeActionTargetsInput{
		ActionTargetArns: aws.StringSlice([]string{actionTargetArn}),
	}
	var found *securityhub.ActionTarget
	err := conn.DescribeActionTargetsPagesWithContext(ctx, input, func(page *securityhub.DescribeActionTargetsOutput, lastPage bool) bool {
		for _, actionTarget := range page.ActionTargets {
			if aws.StringValue(actionTarget.ActionTargetArn) == actionTargetArn {
				found = actionTarget
				return false
			}
		}
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return found, nil
}

func resourceActionTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn(ctx)
	log.Printf("[DEBUG] Deleting Security Hub Action Target %s", d.Id())

	_, err := conn.DeleteActionTargetWithContext(ctx, &securityhub.DeleteActionTargetInput{
		ActionTargetArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Action Target %s: %s", d.Id(), err)
	}

	return diags
}
