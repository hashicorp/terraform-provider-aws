package vpclattice

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpclattice_resource_policy")
func DataSourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourcePolicyRead,
		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameResourcePolicy = "Resource Policy Data Source"
)

func dataSourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	resourceArn := d.Id()

	policy, err := findResourcePolicyByID(ctx, conn, resourceArn)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameResourcePolicy, d.Id(), err)
	}

	if policy == nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameResourcePolicy, d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.ToString(policy.Policy))

	if err != nil {
		return diag.Errorf("setting policy %s: %s", aws.ToString(policy.Policy), err)
	}

	d.Set("policy", policyToSet)

	return nil
}
