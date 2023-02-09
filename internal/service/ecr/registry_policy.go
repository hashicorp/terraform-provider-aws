package ecr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRegistryPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryPolicyPut,
		ReadWithoutTimeout:   resourceRegistryPolicyRead,
		UpdateWithoutTimeout: resourceRegistryPolicyPut,
		DeleteWithoutTimeout: resourceRegistryPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:                  schema.TypeString,
				Required:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRegistryPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn()

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	input := ecr.PutRegistryPolicyInput{
		PolicyText: aws.String(policy),
	}

	out, err := conn.PutRegistryPolicyWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Registry Policy: %s", err)
	}

	regID := aws.StringValue(out.RegistryId)

	d.SetId(regID)

	return append(diags, resourceRegistryPolicyRead(ctx, d, meta)...)
}

func resourceRegistryPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn()

	log.Printf("[DEBUG] Reading registry policy %s", d.Id())
	out, err := conn.GetRegistryPolicyWithContext(ctx, &ecr.GetRegistryPolicyInput{})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
			log.Printf("[WARN] ECR Registry (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading ECR Registry Policy (%s): %s", d.Id(), err)
	}

	d.Set("registry_id", out.RegistryId)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(out.PolicyText))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Registry Policy (%s): setting policy: %s", d.Id(), err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Registry Policy (%s): setting policy: %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourceRegistryPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRConn()

	_, err := conn.DeleteRegistryPolicyWithContext(ctx, &ecr.DeleteRegistryPolicyInput{})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting ECR Registry Policy (%s): %s", d.Id(), err)
	}

	return diags
}
