// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudsearch

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudsearch"
	"github.com/aws/aws-sdk-go-v2/service/cloudsearch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudsearch_domain_service_access_policy", name="Domain Service Access Policy")
func resourceDomainServiceAccessPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainServiceAccessPolicyPut,
		ReadWithoutTimeout:   resourceDomainServiceAccessPolicyRead,
		UpdateWithoutTimeout: resourceDomainServiceAccessPolicyPut,
		DeleteWithoutTimeout: resourceDomainServiceAccessPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_policy": {
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
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDomainServiceAccessPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchClient(ctx)

	accessPolicy := d.Get("access_policy").(string)
	policy, err := structure.NormalizeJsonString(accessPolicy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	domainName := d.Get(names.AttrDomainName).(string)
	input := &cloudsearch.UpdateServiceAccessPoliciesInput{
		AccessPolicies: aws.String(policy),
		DomainName:     aws.String(domainName),
	}

	_, err = conn.UpdateServiceAccessPolicies(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudSearch Domain Service Access Policy (%s): %s", domainName, err)
	}

	if d.IsNewResource() {
		d.SetId(domainName)
	}

	if _, err := waitAccessPolicyActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudSearch Domain Service Access Policy (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceDomainServiceAccessPolicyRead(ctx, d, meta)...)
}

func resourceDomainServiceAccessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchClient(ctx)

	accessPolicy, err := findAccessPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudSearch Domain Service Access Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudSearch Domain Service Access Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("access_policy").(string), accessPolicy)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudSearch Domain Service Access Policy (%s): %s", d.Id(), err)
	}

	d.Set("access_policy", policyToSet)
	d.Set(names.AttrDomainName, d.Id())

	return diags
}

func resourceDomainServiceAccessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchClient(ctx)

	log.Printf("[DEBUG] Deleting CloudSearch Domain Service Access Policy: %s", d.Id())
	_, err := conn.UpdateServiceAccessPolicies(ctx, &cloudsearch.UpdateServiceAccessPoliciesInput{
		AccessPolicies: aws.String(""),
		DomainName:     aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudSearch Domain Service Access Policy (%s): %s", d.Id(), err)
	}

	if _, err := waitAccessPolicyActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudSearch Domain Service Access Policy (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findAccessPolicyByName(ctx context.Context, conn *cloudsearch.Client, name string) (string, error) {
	output, err := findAccessPoliciesStatusByName(ctx, conn, name)

	if err != nil {
		return "", err
	}

	accessPolicy := aws.ToString(output.Options)

	if accessPolicy == "" {
		return "", tfresource.NewEmptyResultError(name)
	}

	return accessPolicy, nil
}

func findAccessPoliciesStatusByName(ctx context.Context, conn *cloudsearch.Client, name string) (*types.AccessPoliciesStatus, error) {
	input := &cloudsearch.DescribeServiceAccessPoliciesInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeServiceAccessPolicies(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccessPolicies == nil || output.AccessPolicies.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AccessPolicies, nil
}

func statusAccessPolicyState(ctx context.Context, conn *cloudsearch.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAccessPoliciesStatusByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.State), nil
	}
}

func waitAccessPolicyActive(ctx context.Context, conn *cloudsearch.Client, name string, timeout time.Duration) (*types.AccessPoliciesStatus, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.OptionStateProcessing),
		Target:  enum.Slice(types.OptionStateActive),
		Refresh: statusAccessPolicyState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AccessPoliciesStatus); ok {
		return output, err
	}

	return nil, err
}
