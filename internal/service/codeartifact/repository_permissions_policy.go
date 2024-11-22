// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact/types"
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

// @SDKResource("aws_codeartifact_repository_permissions_policy", name="Repository Permissions Policy")
func resourceRepositoryPermissionsPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryPermissionsPolicyPut,
		UpdateWithoutTimeout: resourceRepositoryPermissionsPolicyPut,
		ReadWithoutTimeout:   resourceRepositoryPermissionsPolicyRead,
		DeleteWithoutTimeout: resourceRepositoryPermissionsPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRepositoryPermissionsPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &codeartifact.PutRepositoryPermissionsPolicyInput{
		Domain:         aws.String(d.Get(names.AttrDomain).(string)),
		PolicyDocument: aws.String(policy),
		Repository:     aws.String(d.Get("repository").(string)),
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		input.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_revision"); ok {
		input.PolicyRevision = aws.String(v.(string))
	}

	output, err := conn.PutRepositoryPermissionsPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeArtifact Repository Permissions Policy: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(aws.ToString(output.Policy.ResourceArn))
	}

	return append(diags, resourceRepositoryPermissionsPolicyRead(ctx, d, meta)...)
}

func resourceRepositoryPermissionsPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	owner, domainName, repositoryName, err := parseRepositoryARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policy, err := findRepositoryPermissionsPolicyByThreePartKey(ctx, conn, owner, domainName, repositoryName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeArtifact Repository Permissions Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeArtifact Repository Permissions Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDomain, domainName)
	d.Set("domain_owner", owner)
	d.Set("policy_revision", policy.Revision)
	d.Set("repository", repositoryName)
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

func resourceRepositoryPermissionsPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactClient(ctx)

	owner, domainName, repositoryName, err := parseRepositoryARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting CodeArtifact Repository Permissions Policy: %s", d.Id())
	_, err = conn.DeleteRepositoryPermissionsPolicy(ctx, &codeartifact.DeleteRepositoryPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(owner),
		Repository:  aws.String(repositoryName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Repository Permissions Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findRepositoryPermissionsPolicyByThreePartKey(ctx context.Context, conn *codeartifact.Client, owner, domainName, repositoryName string) (*types.ResourcePolicy, error) {
	input := &codeartifact.GetRepositoryPermissionsPolicyInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(owner),
		Repository:  aws.String(repositoryName),
	}

	output, err := conn.GetRepositoryPermissionsPolicy(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policy, nil
}
