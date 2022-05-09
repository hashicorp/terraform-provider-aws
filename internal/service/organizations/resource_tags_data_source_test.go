package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccResourceTagsDataSource_basic(t *testing.T) {
	resourceName := "aws_organizations_policy.test"
	dataSourceName := "data.aws_organizations_resource_tags.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTagsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "tags.#", dataSourceName, "tags.#"),
				),
			},
		},
	})
}

func testAccResourceTagsDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_policy" "test" {
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

  depends_on = [aws_organizations_organization.test]

  tags = {
    TerraformProviderAwsTest = true
    Alpha                    = 1
  }
}

data "aws_organizations_resource_tags" "test" {
  resource_id = aws_organizations_policy.test.id
}
`, rName)
}
