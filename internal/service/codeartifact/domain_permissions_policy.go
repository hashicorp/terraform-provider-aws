// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package codeartifact

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codeartifact_domain_permissions_policy", name="Domain Permissions Policy")
// @ArnIdentity("resource_arn")
// @V60SDKv2Fix
// @Testing(serialize=true)
func resourceDomainPermissionsPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainPermissionsPolicyPut,
		UpdateWithoutTimeout: resourceDomainPermissionsPolicyPut,
		ReadWithoutTimeout:   resourceDomainPermissionsPolicyRead,
		DeleteWithoutTimeout: resourceDomainPermissionsPolicyDelete,

		Schema: map[string]*schema.Schema{
			names.AttrDomain: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"policy_document": sdkv2.IAMPolicyDocumentSchemaOptionalComputed(),
			"policy_revision": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDomainPermissionsPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &codeartifact.PutDomainPermissionsPolicyInput{
		Domain:         aws.String(d.Get(names.AttrDomain).(string)),
		PolicyDocument: aws.String(policy),
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		input.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_revision"); ok {
		input.PolicyRevision = aws.String(v.(string))
	}

	output, err := conn.PutDomainPermissionsPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeArtifact Domain Permissions Policy: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(aws.ToString(output.Policy.ResourceArn))
	}

	return append(diags, resourceDomainPermissionsPolicyRead(ctx, d, meta)...)
}

func resourceDomainPermissionsPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	owner, domainName, err := parseDomainARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err := findDomainPermissionsPolicyByTwoPartKey(ctx, conn, owner, domainName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CodeArtifact Domain Permissions Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeArtifact Domain Permissions Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDomain, domainName)
	d.Set("domain_owner", owner)
	d.Set("policy_revision", policy.Revision)
	d.Set(names.AttrResourceARN, policy.ResourceArn)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy_document").(string), aws.ToString(policy.Document))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("policy_document", policyToSet)

	return diags
}

func resourceDomainPermissionsPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	owner, domainName, err := parseDomainARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting CodeArtifact Domain Permissions Policy: %s", d.Id())
	input := codeartifact.DeleteDomainPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(owner),
	}
	_, err = conn.DeleteDomainPermissionsPolicy(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Domain Permissions Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findDomainPermissionsPolicyByTwoPartKey(ctx context.Context, conn *codeartifact.Client, owner, domainName string) (*types.ResourcePolicy, error) {
	input := &codeartifact.GetDomainPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(owner),
	}

	output, err := conn.GetDomainPermissionsPolicy(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

	return output.Policy, nil
}
