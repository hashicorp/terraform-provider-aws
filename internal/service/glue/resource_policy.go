// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package glue

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_resource_policy", name="Resource Policy")
// @SingletonIdentity
// @V60SDKv2Fix
// @Testing(hasExistsFunction=false)
// @Testing(generator=false)
func resourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyPut(awstypes.ExistConditionNotExist),
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyPut(awstypes.ExistConditionMustExist),
		DeleteWithoutTimeout: resourceResourcePolicyDelete,

		Schema: map[string]*schema.Schema{
			"enable_hybrid": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.EnableHybridValues](),
			},
			names.AttrPolicy: sdkv2.IAMPolicyDocumentSchemaRequired(),
		},
	}
}

func resourceResourcePolicyPut(condition awstypes.ExistCondition) func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		var diags diag.Diagnostics
		conn := meta.(*conns.AWSClient).GlueClient(ctx)

		policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := glue.PutResourcePolicyInput{
			PolicyExistsCondition: condition,
			PolicyInJson:          aws.String(policy),
		}

		if v, ok := d.GetOk("enable_hybrid"); ok {
			input.EnableHybrid = awstypes.EnableHybridValues(v.(string))
		}

		_, err = conn.PutResourcePolicy(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting Glue Resource Policy: %s", err)
		}

		if d.IsNewResource() {
			d.SetId(meta.(*conns.AWSClient).Region(ctx))
		}

		return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
	}
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	output, err := findResourcePolicy(ctx, conn)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Glue Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Resource Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.PolicyInJson))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	input := glue.DeleteResourcePolicyInput{}
	_, err := conn.DeleteResourcePolicy(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Resource Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourcePolicy(ctx context.Context, conn *glue.Client) (*glue.GetResourcePolicyOutput, error) {
	input := &glue.GetResourcePolicyInput{}
	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || aws.ToString(output.PolicyInJson) == "" {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
