package organizations_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOrganizationsPoliciesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	//Setting the two data sources on the same policy
	datasourceServiceControlPolicy := "data.aws_organizations_policy.test"
	dataSourceServiceControlPolicies := "data.aws_organizations_policies.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	serviceControlPolicyContent := `{"Version": "2012-10-17", "Statement": { "Effect": "Deny", "Action": "*", "Resource": "*"}}`

	//Set Up List calls for each policy type
	listServiceControlPolicies := "data.aws_organizations_policies.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPoliciesDataSourceConfig_ServiceControlPolicy(rName, organizations.PolicyTypeServiceControlPolicy, serviceControlPolicyContent),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(listServiceControlPolicies, "policies.#", "0"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceServiceControlPolicies, "policies.0.arn", datasourceServiceControlPolicy, "arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceServiceControlPolicies, "policies.0.aws_managed", datasourceServiceControlPolicy, "aws_managed"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceServiceControlPolicies, "policies.0.description", datasourceServiceControlPolicy, "description"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceServiceControlPolicies, "policies.0.name", datasourceServiceControlPolicy, "name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceServiceControlPolicies, "policies.0.type", datasourceServiceControlPolicy, "type"),
				),
			},
		},
	})
}

func testAccPoliciesDataSourceConfig_ServiceControlPolicy(rName, policyType, policyContent string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
  depends_on = [data.aws_organizations_organization.test]

  name    = "%s"
  type    = "%s"
  content = %s
}

data "aws_organizations_policies" "test" {
  depends_on=[aws_organizations_policy.test]
  filter="%s"
}

data "aws_organizations_policy" "test" {
  depends_on=[aws_organizations_policy.test]
  policy_id=data.aws_organizations_policies.test.policies[0].id
}

`, rName, policyType, strconv.Quote(policyContent), policyType)
}
