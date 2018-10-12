package aws

import (
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func dataSourceAwsIAMPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIAMPolicyRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"arn"},
			},
			"path_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "/",
				ConflictsWith: []string{"arn"},
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsIAMPolicyRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	var iamPolicyARN string
	if v, ok := d.GetOk("name"); ok {
		input := &iam.ListPoliciesInput{
			PathPrefix: aws.String(d.Get("path_prefix").(string)),
		}

		policyArns := make([]string, 0)
		if err := iamconn.ListPoliciesPages(input, func(res *iam.ListPoliciesOutput, lastPage bool) bool {
			for _, policy := range res.Policies {
				if v.(string) == aws.StringValue(policy.PolicyName) {
					policyArns = append(policyArns, aws.StringValue(policy.Arn))
				}
			}
			return !lastPage
		}); err != nil {
			return err
		}

		if len(policyArns) == 0 {
			return fmt.Errorf("no matching IAM policy found")
		}
		if len(policyArns) > 1 {
			return fmt.Errorf("multiple IAM policies matched; use additional constraints to reduce matches to a single IAM policy")
		}

		iamPolicyARN = policyArns[0]

	} else if v, ok := d.GetOk("arn"); ok {
		iamPolicyARN = v.(string)
	} else {
		return fmt.Errorf(
			"ARN or name must be set to query IAM policy")
	}

	d.SetId(iamPolicyARN)
	return resourceAwsIamPolicyRead(d, meta)
}
