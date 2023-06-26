package opensearchserverless

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_opensearchserverless_security_policy")
func DataSourceSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSecurityPolicyRead,

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceSecurityPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient(ctx)

	securityPolicyName := d.Get("name").(string)
	securityPolicyType := d.Get("type").(string)
	securityPolicy, err := FindSecurityPolicyByNameAndType(ctx, conn, securityPolicyName, securityPolicyType)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SecurityPolicy with name (%s) and type (%s): %s", securityPolicyName, securityPolicyType, err)
	}

	policyBytes, err := securityPolicy.Policy.MarshalSmithyDocument()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading JSON policy document for SecurityPolicy with name %s and type %s: %s", securityPolicyName, securityPolicyType, err)
	}

	d.SetId(aws.ToString(securityPolicy.Name))
	d.Set("description", securityPolicy.Description)
	d.Set("name", securityPolicy.Name)
	d.Set("policy", string(policyBytes))
	d.Set("policy_version", securityPolicy.PolicyVersion)
	d.Set("type", securityPolicy.Type)

	return diags
}
