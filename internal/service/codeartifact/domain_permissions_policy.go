// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codeartifact_domain_permissions_policy")
func ResourceDomainPermissionsPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainPermissionsPolicyPut,
		UpdateWithoutTimeout: resourceDomainPermissionsPolicyPut,
		ReadWithoutTimeout:   resourceDomainPermissionsPolicyRead,
		DeleteWithoutTimeout: resourceDomainPermissionsPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
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
			"policy_document": {
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
			"policy_revision": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDomainPermissionsPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)
	log.Print("[DEBUG] Creating CodeArtifact Domain Permissions Policy")

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	params := &codeartifact.PutDomainPermissionsPolicyInput{
		Domain:         aws.String(d.Get("domain").(string)),
		PolicyDocument: aws.String(policy),
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		params.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_revision"); ok {
		params.PolicyRevision = aws.String(v.(string))
	}

	res, err := conn.PutDomainPermissionsPolicyWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeArtifact Domain Permissions Policy: %s", err)
	}

	d.SetId(aws.StringValue(res.Policy.ResourceArn))

	return append(diags, resourceDomainPermissionsPolicyRead(ctx, d, meta)...)
}

func resourceDomainPermissionsPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)
	log.Printf("[DEBUG] Reading CodeArtifact Domain Permissions Policy: %s", d.Id())

	domainOwner, domainName, err := DecodeDomainID(d.Id())
	if err != nil {
		return create.DiagError(names.CodeArtifact, create.ErrActionReading, ResNameDomainPermissionsPolicy, d.Id(), err)
	}

	dm, err := conn.GetDomainPermissionsPolicyWithContext(ctx, &codeartifact.GetDomainPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CodeArtifact, create.ErrActionReading, ResNameDomainPermissionsPolicy, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CodeArtifact, create.ErrActionReading, ResNameDomainPermissionsPolicy, d.Id(), err)
	}

	d.Set("domain", domainName)
	d.Set("domain_owner", domainOwner)
	d.Set("resource_arn", dm.Policy.ResourceArn)
	d.Set("policy_revision", dm.Policy.Revision)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy_document").(string), aws.StringValue(dm.Policy.Document))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policyToSet, err)
	}

	d.Set("policy_document", policyToSet)

	return diags
}

func resourceDomainPermissionsPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)
	log.Printf("[DEBUG] Deleting CodeArtifact Domain Permissions Policy: %s", d.Id())

	domainOwner, domainName, err := DecodeDomainID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Domain Permissions Policy (%s): %s", d.Id(), err)
	}

	input := &codeartifact.DeleteDomainPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
	}

	_, err = conn.DeleteDomainPermissionsPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Domain Permissions Policy (%s): %s", d.Id(), err)
	}

	return diags
}
