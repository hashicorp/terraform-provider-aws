// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"log"

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

// @SDKResource("aws_vpclattice_resource_policy", name="Resource Policy")
// @Testing(tagsTest=false)
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
			names.AttrPolicy: sdkv2.IAMPolicyDocumentSchemaRequired(),
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceResourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resourceARN := d.Get(names.AttrResourceARN).(string)
	input := vpclattice.PutResourcePolicyInput{
		Policy:      aws.String(policy),
		ResourceArn: aws.String(resourceARN),
	}

	_, err = conn.PutResourcePolicy(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPCLattice Resource Policy (%s): %s", resourceARN, err)
	}

	if d.IsNewResource() {
		d.SetId(resourceARN)
	}

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	output, err := findResourcePolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Resource Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.Policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set(names.AttrPolicy, policyToSet)
	d.Set(names.AttrResourceARN, d.Id())

	return diags
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPCLattice Resource Policy: %s", d.Id())
	input := vpclattice.DeleteResourcePolicyInput{
		ResourceArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteResourcePolicy(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPCLattice Resource Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourcePolicyByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetResourcePolicyOutput, error) {
	input := vpclattice.GetResourcePolicyInput{
		ResourceArn: aws.String(id),
	}

	return findResourcePolicy(ctx, conn, &input)
}

func findResourcePolicy(ctx context.Context, conn *vpclattice.Client, input *vpclattice.GetResourcePolicyInput) (*vpclattice.GetResourcePolicyOutput, error) {
	output, err := conn.GetResourcePolicy(ctx, input)

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
