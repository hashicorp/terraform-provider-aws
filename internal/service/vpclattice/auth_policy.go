// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_auth_policy", name="Auth Policy")
func resourceAuthPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAuthPolicyPut,
		ReadWithoutTimeout:   resourceAuthPolicyRead,
		UpdateWithoutTimeout: resourceAuthPolicyPut,
		DeleteWithoutTimeout: resourceAuthPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: sdkv2.IAMPolicyDocumentSchemaRequired(),
			"resource_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAuthPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resourceID := d.Get("resource_identifier").(string)
	input := vpclattice.PutAuthPolicyInput{
		Policy:             aws.String(policy),
		ResourceIdentifier: aws.String(resourceID),
	}

	_, err = conn.PutAuthPolicy(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPCLattice Auth Policy (%s): %s", resourceID, err)
	}

	d.SetId(resourceID)

	return append(diags, resourceAuthPolicyRead(ctx, d, meta)...)
}

func resourceAuthPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	output, err := findAuthPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Auth Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Auth Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.Policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set(names.AttrPolicy, policyToSet)
	d.Set("resource_identifier", d.Id())

	return diags
}

func resourceAuthPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPCLattice Auth Policy: %s", d.Id())
	input := vpclattice.DeleteAuthPolicyInput{
		ResourceIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteAuthPolicy(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPCLattice Auth Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findAuthPolicyByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetAuthPolicyOutput, error) {
	input := vpclattice.GetAuthPolicyInput{
		ResourceIdentifier: aws.String(id),
	}

	return findAuthPolicy(ctx, conn, &input)
}

func findAuthPolicy(ctx context.Context, conn *vpclattice.Client, input *vpclattice.GetAuthPolicyInput) (*vpclattice.GetAuthPolicyOutput, error) {
	output, err := conn.GetAuthPolicy(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
