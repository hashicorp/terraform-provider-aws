// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package rum

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rum/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_rum_resource_policy", name="Resource Policy")
func resourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyPut,
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyPut,
		DeleteWithoutTimeout: resourceResourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"app_monitor_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"policy_document": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			"policy_revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceResourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	name := d.Get("app_monitor_name").(string)
	policyDoc := d.Get("policy_document").(string)

	input := &rum.PutResourcePolicyInput{
		Name:           aws.String(name),
		PolicyDocument: aws.String(policyDoc),
	}

	log.Printf("[DEBUG] Putting CloudWatch RUM Resource Policy for App Monitor: %s", name)
	_, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CloudWatch RUM Resource Policy (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	output, err := findResourcePolicy(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CloudWatch RUM Resource Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch RUM Resource Policy (%s): %s", d.Id(), err)
	}

	d.Set("app_monitor_name", d.Id())
	d.Set("policy_document", output.PolicyDocument)
	d.Set("policy_revision_id", output.PolicyRevisionId)

	return diags
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RUMClient(ctx)

	input := &rum.DeleteResourcePolicyInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting CloudWatch RUM Resource Policy for App Monitor: %s", d.Id())
	_, err := conn.DeleteResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.PolicyNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch RUM Resource Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourcePolicy(ctx context.Context, conn *rum.Client, name string) (*rum.GetResourcePolicyOutput, error) {
	input := &rum.GetResourcePolicyInput{
		Name: aws.String(name),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.PolicyNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
