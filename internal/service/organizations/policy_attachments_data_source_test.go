package organizations_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccPolicyAttachmentsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy.test"
	dataSourceName := "data.aws_organizations_policy_attachments.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "policies.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "filter", dataSourceName, "policies.0.type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "filter", dataSourceName, "policies.1.type"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "policies.*", map[string]string{
						"aws_managed": "true",
						"id":          "p-FullAWSAccess",
						"name":        "FullAWSAccess",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "policies.*", map[string]string{
						"aws_managed": "false",
						"name":        rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "policies.*.arn", resourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "policies.*.id", resourceName, "id"),
				),
			},
		},
	})
}

func testAccPolicyAttachmentsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
%s

data "aws_organizations_policy_attachments" "test" {
  target_id = aws_organizations_organizational_unit.test.id
  filter    = "SERVICE_CONTROL_POLICY"

  depends_on = [aws_organizations_policy_attachment.test]
}
`, testAccPolicyAttachmentConfig_organizationalUnit(rName))
}
