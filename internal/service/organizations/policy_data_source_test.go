package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOrganizationPolicyDataSource_UnattachedPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	dataSourceName := "data.aws_organizations_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationPolicyDataSourceConfig_UnattachedPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "policy_id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "content", dataSourceName, "content"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
				),
			},
		},
	})
}

func testAccOrganizationPolicyDataSourceConfig_UnattachedPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
	feature_set = "ALL"
	enabled_policy_types = "ALL"
}

resource "aws_organizations_policy" "test" {
	depends_on = [aws_organizations_organization.test]
	content = <<EOF
{
	"Version": "2012-10-17",
	"Statement": {
	"Effect": "Allow",
	"Action": "*",
	"Resource": "*"
	}
}
EOF

name = %[1]q
}


data "aws_organizations_policy" "test" {
	policy_id = aws_organizations_policy.test.id
}
`, rName)
}
