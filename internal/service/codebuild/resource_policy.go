// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package codebuild

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codebuild_resource_policy", name="Resource Policy")
// @ArnIdentity("resource_arn")
// @V60SDKv2Fix
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/codebuild;codebuild.GetResourcePolicyOutput")
func resourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyPut,
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyPut,
		DeleteWithoutTimeout: resourceResourcePolicyDelete,

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
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &codebuild.PutResourcePolicyInput{
		Policy:      aws.String(policy),
		ResourceArn: aws.String(d.Get(names.AttrResourceARN).(string)),
	}

	output, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CodeBuild Resource Policy: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(aws.ToString(output.ResourceArn))
	}

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	output, err := findResourcePolicyByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CodeBuild Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeBuild Resource Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), aws.ToString(output.Policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)
	d.Set(names.AttrResourceARN, d.Id())

	return diags
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	log.Printf("[INFO] Deleting CodeBuild Resource Policy: %s", d.Id())
	input := codebuild.DeleteResourcePolicyInput{
		ResourceArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteResourcePolicy(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Resource Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourcePolicyByARN(ctx context.Context, conn *codebuild.Client, arn string) (*codebuild.GetResourcePolicyOutput, error) {
	input := &codebuild.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
