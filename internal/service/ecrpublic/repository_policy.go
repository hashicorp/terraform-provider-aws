// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
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

// @SDKResource("aws_ecrpublic_repository_policy")
func ResourceRepositoryPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryPolicyPut,
		ReadWithoutTimeout:   resourceRepositoryPolicyRead,
		UpdateWithoutTimeout: resourceRepositoryPolicyPut,
		DeleteWithoutTimeout: resourceRepositoryPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
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
			names.AttrRepositoryName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	policyPutTimeout = 2 * time.Minute
)

func resourceRepositoryPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	repositoryName := d.Get(names.AttrRepositoryName).(string)
	input := &ecrpublic.SetRepositoryPolicyInput{
		PolicyText:     aws.String(policy),
		RepositoryName: aws.String(repositoryName),
	}

	outputRaw, err := tfresource.RetryWhen(ctx, policyPutTimeout,
		func() (interface{}, error) {
			return conn.SetRepositoryPolicy(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Invalid repository policy provided") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ECR Public Repository (%s) Policy: %s", repositoryName, err)
	}

	if d.IsNewResource() {
		d.SetId(aws.ToString(outputRaw.(*ecrpublic.SetRepositoryPolicyOutput).RepositoryName))
	}

	return append(diags, resourceRepositoryPolicyRead(ctx, d, meta)...)
}

func resourceRepositoryPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicClient(ctx)

	output, err := FindRepositoryPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Public Repository Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Public Repository Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), aws.ToString(output.PolicyText))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is an invalid JSON: %s", policyToSet, err)
	}

	d.Set(names.AttrPolicy, policyToSet)
	d.Set("registry_id", output.RegistryId)
	d.Set(names.AttrRepositoryName, output.RepositoryName)

	return diags
}

func resourceRepositoryPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRPublicClient(ctx)

	_, err := conn.DeleteRepositoryPolicy(ctx, &ecrpublic.DeleteRepositoryPolicyInput{
		RegistryId:     aws.String(d.Get("registry_id").(string)),
		RepositoryName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.RepositoryNotFoundException](err) || errs.IsA[*awstypes.RepositoryPolicyNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Public Repository Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func FindRepositoryPolicyByName(ctx context.Context, conn *ecrpublic.Client, name string) (*ecrpublic.GetRepositoryPolicyOutput, error) {
	input := &ecrpublic.GetRepositoryPolicyInput{
		RepositoryName: aws.String(name),
	}

	output, err := conn.GetRepositoryPolicy(ctx, input)

	if errs.IsA[*awstypes.RepositoryNotFoundException](err) || errs.IsA[*awstypes.RepositoryPolicyNotFoundException](err) {
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
