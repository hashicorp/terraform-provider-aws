// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/schemas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_schemas_registry_policy", name="Registry Policy")
func resourceRegistryPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryPolicyCreate,
		ReadWithoutTimeout:   resourceRegistryPolicyRead,
		UpdateWithoutTimeout: resourceRegistryPolicyUpdate,
		DeleteWithoutTimeout: resourceRegistryPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRegistryPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	registryName := d.Get("registry_name").(string)
	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &schemas.PutResourcePolicyInput{
		Policy:       aws.String(policy),
		RegistryName: aws.String(registryName),
	}

	_, err = conn.PutResourcePolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Schemas Registry Policy (%s): %s", registryName, err)
	}

	d.SetId(registryName)

	return append(diags, resourceRegistryPolicyRead(ctx, d, meta)...)
}

func resourceRegistryPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	output, err := findRegistryPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Registry Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Schemas Registry Policy (%s): %s", d.Id(), err)
	}

	if output.Policy != nil {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.Policy))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	} else {
		d.Set(names.AttrPolicy, nil)
	}
	d.Set("registry_name", d.Id())

	return diags
}

func resourceRegistryPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	if d.HasChanges(names.AttrPolicy) {
		policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &schemas.PutResourcePolicyInput{
			Policy:       aws.String(policy),
			RegistryName: aws.String(d.Id()),
		}

		_, err = conn.PutResourcePolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Schemas Registry Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRegistryPolicyRead(ctx, d, meta)...)
}

func resourceRegistryPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Schemas Registry Policy (%s)", d.Id())
	_, err := conn.DeleteResourcePolicy(ctx, &schemas.DeleteResourcePolicyInput{
		RegistryName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Schemas Registry Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findRegistryPolicyByName(ctx context.Context, conn *schemas.Client, name string) (*schemas.GetResourcePolicyOutput, error) {
	input := &schemas.GetResourcePolicyInput{
		RegistryName: aws.String(name),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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
