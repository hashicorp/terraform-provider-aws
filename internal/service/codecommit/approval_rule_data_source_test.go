package codecommit_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codecommit"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCodeCommitApprovalRuleTemplateDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule.default"
	datasourceName := "data.aws_codecommit_approval_rule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, codecommit.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCodeCommitApprovalRuleTemplateDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "content", resourceName, "content"),
				),
			},
		},
	})
}

func testAccCheckCodeCommitApprovalRuleTemplateDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_approval_rule" "default" {
	name = "%s"
	content = <<EOF
{
	"DestinationReferences": ["refs/heads/master"],
	"Statements": [{
			"Type": "Approvers",
			"NumberOfApprovalsNeeded": 2,
			"ApprovalPoolMembers": ["arn:aws:sts::444194214726:assumed-role/CodeCommitReview/*"]}],
	"Version": "2018-11-08"
}
EOF
}

data "aws_codecommit_approval_rule" "default" {
  name = "${aws_codecommit_approval_rule.default.name}"
}
`, rName)
}
