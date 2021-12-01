package codecommit_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/codecommit"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAWSCodeCommitApprovalRule_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, codecommit.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCodeCommitApprovalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeCommitApprovalRule_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitApprovalRuleExists(resourceName),
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

func TestAccAWSCodeCommitApprovalRule_withChanges(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_approval_rule.test"

	testContent, _ := structure.NormalizeJsonString(fmt.Sprintf(`
	{
		"Version": "2018-11-08",
		"DestinationReferences": ["refs/heads/master"],
		"Statements": [{
				"Type": "Approvers",
				"NumberOfApprovalsNeeded": 1,
				"ApprovalPoolMembers": ["arn:aws:sts::444194214726:assumed-role/CodeCommitReview/*"]}]
	}`))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, codecommit.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCodeCommitApprovalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeCommitApprovalRule_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitApprovalRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodeCommitApprovalRule_withChanges(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeCommitApprovalRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test description - with changes"),
					resource.TestCheckResourceAttr(resourceName, "content", testContent),
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

func testAccCheckCodeCommitApprovalRuleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn
		out, err := conn.GetApprovalRuleTemplate(&codecommit.GetApprovalRuleTemplateInput{
			ApprovalRuleTemplateName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if out.ApprovalRuleTemplate.ApprovalRuleTemplateId == nil {
			return fmt.Errorf("No CodeCommit Approval Rule Template Found")
		}

		if *out.ApprovalRuleTemplate.ApprovalRuleTemplateName != rs.Primary.ID {
			return fmt.Errorf("CodeCommit Approval Rule Template mismatch - existing: %q, state: %q",
				*out.ApprovalRuleTemplate.ApprovalRuleTemplateName, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCodeCommitApprovalRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codecommit_approval_rule" {
			continue
		}

		_, err := conn.GetApprovalRuleTemplate(&codecommit.GetApprovalRuleTemplateInput{
			ApprovalRuleTemplateName: aws.String(rs.Primary.ID),
		})

		if ae, ok := err.(awserr.Error); ok && ae.Code() == "ApprovalRuleTemplateDoesNotExistException" {
			continue
		}
		if err == nil {
			return fmt.Errorf("Approval rule template still exists: %s", rs.Primary.ID)
		}
		return err
	}

	return nil
}

func testAccCodeCommitApprovalRule_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_approval_rule" "test" {
  name = %[1]q

  description = "This is a test description"

  content = <<EOF
  {
	  "Version": "2018-11-08",
	  "DestinationReferences": ["refs/heads/master"],
	  "Statements": [{
			  "Type": "Approvers",
			  "NumberOfApprovalsNeeded": 2,
			  "ApprovalPoolMembers": ["arn:aws:sts::444194214726:assumed-role/CodeCommitReview/*"]}]
  }
  EOF 
}
`, rName)
}

func testAccCodeCommitApprovalRule_withChanges(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_approval_rule" "test" {
  name        = %[1]q
  description = "This is a test description - with changes"

  content = <<EOF
  {
	  "Version": "2018-11-08",
	  "DestinationReferences": ["refs/heads/master"],
	  "Statements": [{
			  "Type": "Approvers",
			  "NumberOfApprovalsNeeded": 1,
			  "ApprovalPoolMembers": ["arn:aws:sts::444194214726:assumed-role/CodeCommitReview/*"]}]
  }
  EOF 
}
`, rName)
}
