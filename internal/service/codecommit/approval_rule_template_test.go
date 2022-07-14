package codecommit_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodecommit "github.com/hashicorp/terraform-provider-aws/internal/service/codecommit"
)

func TestAccCodeCommitApprovalRuleTemplate_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApprovalRuleTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(resourceName),
					testAccCheckApprovalRuleTemplateContent(resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "approval_rule_template_id"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_user"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_content_sha256"),
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

func TestAccCodeCommitApprovalRuleTemplate_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApprovalRuleTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodecommit.ResourceApprovalRuleTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeCommitApprovalRuleTemplate_updateContentAndDescription(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApprovalRuleTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(resourceName),
					testAccCheckApprovalRuleTemplateContent(resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccApprovalRuleTemplateConfig_updateContentAndDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(resourceName),
					testAccCheckApprovalRuleTemplateContent(resourceName, 1),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test description"),
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

func TestAccCodeCommitApprovalRuleTemplate_updateName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_codecommit_approval_rule_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApprovalRuleTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccApprovalRuleTemplateConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovalRuleTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
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

func testAccCheckApprovalRuleTemplateContent(resourceName string, numApprovals int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedContent := fmt.Sprintf(`{"Version":"2018-11-08","DestinationReferences":["refs/heads/master"],"Statements":[{"Type":"Approvers","NumberOfApprovalsNeeded":%d,"ApprovalPoolMembers":["arn:%s:sts::%s:assumed-role/CodeCommitReview/*"]}]}`,
			numApprovals, acctest.Partition(), acctest.AccountID(),
		)
		return resource.TestCheckResourceAttr(resourceName, "content", expectedContent)(s)
	}
}

func testAccCheckApprovalRuleTemplateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn

		_, err := conn.GetApprovalRuleTemplate(&codecommit.GetApprovalRuleTemplateInput{
			ApprovalRuleTemplateName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckApprovalRuleTemplateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codecommit_approval_rule_template" {
			continue
		}

		_, err := conn.GetApprovalRuleTemplate(&codecommit.GetApprovalRuleTemplateInput{
			ApprovalRuleTemplateName: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, codecommit.ErrCodeApprovalRuleTemplateDoesNotExistException) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CodeCommit Approval Rule Template (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccApprovalRuleTemplateConfig_basic(rName string) string {
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
`, rName)
}

func testAccApprovalRuleTemplateConfig_updateContentAndDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_codecommit_approval_rule_template" "test" {
  name        = %[1]q
  description = "This is a test description"

  content = <<EOF
  {
	  "Version": "2018-11-08",
	  "DestinationReferences": ["refs/heads/master"],
	  "Statements": [{
			  "Type": "Approvers",
			  "NumberOfApprovalsNeeded": 1,
			  "ApprovalPoolMembers": ["arn:${data.aws_partition.current.partition}:sts::${data.aws_caller_identity.current.account_id}:assumed-role/CodeCommitReview/*"]}]
  }
  EOF 
}
`, rName)
}
