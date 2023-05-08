package vpclattice

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
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

	resourceArn := d.Get("resource_arn").(string)

	out, err := findDataSourceResourcePolicyById(ctx, conn, resourceArn)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameResourcePolicy, d.Id(), err)
	}

	if out == nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameResourcePolicy, d.Id(), err)
	}

	d.Set("resource_arn", resourceArn)
	d.SetId(resourceArn)
	d.Set("policy", aws.ToString(out.Policy))

	// TIP: Setting a JSON string to avoid errorneous diffs.
	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))

	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, DSNameAuthPolicy, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameAuthPolicy, d.Id(), err)
	}

	d.Set("policy", p)

	return nil
}

func findDataSourceResourcePolicyById(ctx context.Context, conn *vpclattice.Client, resource_arn string) (*vpclattice.GetResourcePolicyOutput, error) {
	in := &vpclattice.GetResourcePolicyInput{
		ResourceArn: aws.String(resource_arn),
	}

	out, err := conn.GetResourcePolicy(ctx, in)

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	return out, nil
}
