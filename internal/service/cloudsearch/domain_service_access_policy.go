package cloudsearch

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainServiceAccessPolicy() *schema.Resource {
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
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceDomainServiceAccessPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchConn()

	domainName := d.Get("domain_name").(string)
	input := &cloudsearch.UpdateServiceAccessPoliciesInput{
		DomainName: aws.String(domainName),
	}

	accessPolicy := d.Get("access_policy").(string)
	policy, err := structure.NormalizeJsonString(accessPolicy)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", accessPolicy, err)
	}

	input.AccessPolicies = aws.String(policy)

	log.Printf("[DEBUG] Updating CloudSearch Domain access policies: %s", input)
	_, err = conn.UpdateServiceAccessPoliciesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudSearch Domain Service Access Policy (%s): %s", domainName, err)
	}

	d.SetId(domainName)

	_, err = waitAccessPolicyActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudSearch Domain Service Access Policy (%s) to become active: %s", d.Id(), err)
	}

	return append(diags, resourceDomainServiceAccessPolicyRead(ctx, d, meta)...)
}

func resourceDomainServiceAccessPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchConn()

	accessPolicy, err := FindAccessPolicyByName(ctx, conn, d.Id())

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
	d.Set("domain_name", d.Id())

	return diags
}

func resourceDomainServiceAccessPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudSearchConn()

	input := &cloudsearch.UpdateServiceAccessPoliciesInput{
		AccessPolicies: aws.String(""),
		DomainName:     aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting CloudSearch Domain Service Access Policy: %s", d.Id())
	_, err := conn.UpdateServiceAccessPoliciesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudsearch.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudSearch Domain Service Access Policy (%s): %s", d.Id(), err)
	}

	_, err = waitAccessPolicyActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudSearch Domain Service Access Policy (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func FindAccessPolicyByName(ctx context.Context, conn *cloudsearch.CloudSearch, name string) (string, error) {
	output, err := findAccessPoliciesStatusByName(ctx, conn, name)

	if err != nil {
		return "", err
	}

	accessPolicy := aws.StringValue(output.Options)

	if accessPolicy == "" {
		return "", tfresource.NewEmptyResultError(name)
	}

	return accessPolicy, nil
}

func findAccessPoliciesStatusByName(ctx context.Context, conn *cloudsearch.CloudSearch, name string) (*cloudsearch.AccessPoliciesStatus, error) {
	input := &cloudsearch.DescribeServiceAccessPoliciesInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeServiceAccessPoliciesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudsearch.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func statusAccessPolicyState(ctx context.Context, conn *cloudsearch.CloudSearch, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAccessPoliciesStatusByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.State), nil
	}
}

func waitAccessPolicyActive(ctx context.Context, conn *cloudsearch.CloudSearch, name string, timeout time.Duration) (*cloudsearch.AccessPoliciesStatus, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudsearch.OptionStateProcessing},
		Target:  []string{cloudsearch.OptionStateActive},
		Refresh: statusAccessPolicyState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudsearch.AccessPoliciesStatus); ok {
		return output, err
	}

	return nil, err
}
