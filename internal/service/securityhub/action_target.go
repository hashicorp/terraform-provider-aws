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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "must contain only alphanumeric characters"),
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
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	identifier := d.Get("identifier").(string)

	log.Printf("[DEBUG] Creating Security Hub Action Target %s", identifier)

	resp, err := conn.CreateActionTarget(ctx, &securityhub.CreateActionTargetInput{
		Description: aws.String(d.Get("description").(string)),
		Id:          aws.String(identifier),
		Name:        aws.String(d.Get("name").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Action Target %s: %s", identifier, err)
	}

	d.SetId(aws.ToString(resp.ActionTargetArn))

	return append(diags, resourceActionTargetRead(ctx, d, meta)...)
}

func resourceActionTargetParseIdentifier(identifier string) (string, error) {
	parts := strings.Split(identifier, "/")

	if len(parts) != 3 {
		return "", fmt.Errorf("expected Security Hub Custom action ARN, received: %s", identifier)
	}

	return parts[2], nil
}

func resourceActionTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Reading Security Hub Action Targets to find %s", d.Id())

	actionTargetIdentifier, err := resourceActionTargetParseIdentifier(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Action Targets (%s): %s", d.Id(), err)
	}

	output, err := FindActionTargetByArn(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Reading Security Hub Action Targets with ARN (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("identifier", actionTargetIdentifier)
	d.Set("description", output.Description)
	d.Set("arn", output.ActionTargetArn)
	d.Set("name", output.Name)

	return diags
}

func resourceActionTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.UpdateActionTargetInput{
		ActionTargetArn: aws.String(d.Id()),
		Description:     aws.String(d.Get("description").(string)),
		Name:            aws.String(d.Get("name").(string)),
	}
	if _, err := conn.UpdateActionTarget(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Action Target (%s): %s", d.Id(), err)
	}
	return diags
}

func resourceActionTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)
	log.Printf("[DEBUG] Deleting Security Hub Action Target %s", d.Id())

	_, err := conn.DeleteActionTarget(ctx, &securityhub.DeleteActionTargetInput{
		ActionTargetArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Action Target %s: %s", d.Id(), err)
	}

	return diags
}
