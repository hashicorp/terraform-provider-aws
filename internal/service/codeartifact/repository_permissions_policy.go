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

// @SDKResource("aws_codeartifact_repository_permissions_policy")
func ResourceRepositoryPermissionsPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryPermissionsPolicyPut,
		UpdateWithoutTimeout: resourceRepositoryPermissionsPolicyPut,
		ReadWithoutTimeout:   resourceRepositoryPermissionsPolicyRead,
		DeleteWithoutTimeout: resourceRepositoryPermissionsPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
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

func resourceRepositoryPermissionsPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)
	log.Print("[DEBUG] Creating CodeArtifact Repository Permissions Policy")

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	params := &codeartifact.PutRepositoryPermissionsPolicyInput{
		Domain:         aws.String(d.Get("domain").(string)),
		Repository:     aws.String(d.Get("repository").(string)),
		PolicyDocument: aws.String(policy),
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		params.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_revision"); ok {
		params.PolicyRevision = aws.String(v.(string))
	}

	res, err := conn.PutRepositoryPermissionsPolicyWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeArtifact Repository Permissions Policy: %s", err)
	}

	d.SetId(aws.StringValue(res.Policy.ResourceArn))

	return append(diags, resourceRepositoryPermissionsPolicyRead(ctx, d, meta)...)
}

func resourceRepositoryPermissionsPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)
	log.Printf("[DEBUG] Reading CodeArtifact Repository Permissions Policy: %s", d.Id())

	domainOwner, domainName, repoName, err := DecodeRepositoryID(d.Id())
	if err != nil {
		return create.DiagError(names.CodeArtifact, create.ErrActionReading, ResNameRepositoryPermissionsPolicy, d.Id(), err)
	}

	dm, err := conn.GetRepositoryPermissionsPolicyWithContext(ctx, &codeartifact.GetRepositoryPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
		Repository:  aws.String(repoName),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CodeArtifact, create.ErrActionReading, ResNameRepositoryPermissionsPolicy, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CodeArtifact, create.ErrActionReading, ResNameRepositoryPermissionsPolicy, d.Id(), err)
	}

	d.Set("domain", domainName)
	d.Set("domain_owner", domainOwner)
	d.Set("repository", repoName)
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

func resourceRepositoryPermissionsPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)
	log.Printf("[DEBUG] Deleting CodeArtifact Repository Permissions Policy: %s", d.Id())

	domainOwner, domainName, repoName, err := DecodeRepositoryID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Repository Permissions Policy (%s): %s", d.Id(), err)
	}

	input := &codeartifact.DeleteRepositoryPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
		Repository:  aws.String(repoName),
	}

	_, err = conn.DeleteRepositoryPermissionsPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Repository Permissions Policy (%s): %s", d.Id(), err)
	}

	return diags
}
