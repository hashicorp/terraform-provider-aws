package codecommit_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codecommit"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodecommit "github.com/hashicorp/terraform-provider-aws/internal/service/codecommit"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCodeCommitApprovalRuleTemplateAssociation_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule_template_association.test"
	repoResourceName := "aws_codecommit_repository.test"
	templateResourceName := "aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApprovalRuleTemplateAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "approval_rule_template_name", templateResourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "repository_name", repoResourceName, "repository_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodeCommitApprovalRuleTemplateAssociation_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule_template_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApprovalRuleTemplateAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodecommit.ResourceApprovalRuleTemplateAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeCommitApprovalRuleTemplateAssociation_Disappears_repository(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	repoResourceName := "aws_codecommit_repository.test"
	resourceName := "aws_codecommit_approval_rule_template_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApprovalRuleTemplateAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodecommit.ResourceRepository(), repoResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApprovalRuleTemplateAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn

		approvalTemplateName, repositoryName, err := tfcodecommit.ApprovalRuleTemplateAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		err = tfcodecommit.FindApprovalRuleTemplateAssociation(conn, approvalTemplateName, repositoryName)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckApprovalRuleTemplateAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codecommit_approval_rule_template_association" {
			continue
		}

		approvalTemplateName, repositoryName, err := tfcodecommit.ApprovalRuleTemplateAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		err = tfcodecommit.FindApprovalRuleTemplateAssociation(conn, approvalTemplateName, repositoryName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CodeCommit Approval Rule Template Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccApprovalRuleTemplateAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_codecommit_approval_rule_template" "test" {
  name = %[1]q

  content = <<EOF
  {
	  "Version": "2018-11-08",
	  "DestinationReferences": ["refs/heads/master"],
	  "Statements": [{
			  "Type": "Approvers",
			  "NumberOfApprovalsNeeded": 2,
			  "ApprovalPoolMembers": ["arn:${data.aws_partition.current.partition}:sts::${data.aws_caller_identity.current.account_id}:assumed-role/CodeCommitReview/*"]}]
  }
  EOF 
}

resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
}

resource "aws_codecommit_approval_rule_template_association" "test" {
  approval_rule_template_name = aws_codecommit_approval_rule_template.test.name
  repository_name             = aws_codecommit_repository.test.repository_name
}
`, rName)
}
