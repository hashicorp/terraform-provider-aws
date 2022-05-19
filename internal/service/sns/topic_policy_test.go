package sns_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/sns"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSNSTopicPolicy_basic(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists("aws_sns_topic.test", &attributes),
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_sns_topic.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
					acctest.CheckResourceAttrAccountID(resourceName, "owner"),
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

func TestAccSNSTopicPolicy_updated(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists("aws_sns_topic.test", &attributes),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists("aws_sns_topic.test", &attributes),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile("SNS:DeleteTopic")),
				),
			},
		},
	})
}

func TestAccSNSTopicPolicy_Disappears_topic(t *testing.T) {
	var attributes map[string]string
	topicResourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(topicResourceName, &attributes),
					acctest.CheckResourceDisappears(acctest.Provider, tfsns.ResourceTopic(), topicResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSNSTopicPolicy_disappears(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists("aws_sns_topic.test", &attributes),
					acctest.CheckResourceDisappears(acctest.Provider, tfsns.ResourceTopicPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSNSTopicPolicy_ignoreEquivalent(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyEquivalentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists("aws_sns_topic.test", &attributes),
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_sns_topic.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
					acctest.CheckResourceAttrAccountID(resourceName, "owner"),
				),
			},
			{
				Config:   testAccTopicPolicyEquivalent2Config(rName),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckTopicPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SNSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sns_topic_policy" {
			continue
		}

		_, err := tfsns.FindTopicAttributesByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SNS Topic Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccTopicPolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Id":"default",
  "Statement":[
    {
      "Sid":"%[1]s",
      "Effect":"Allow",
      "Principal":{
        "AWS":"*"
      },
      "Action":[
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission"
      ],
      "Resource":"${aws_sns_topic.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccTopicPolicyUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Id":"default",
  "Statement":[
    {
      "Sid":"%[1]s",
      "Effect":"Allow",
      "Principal":{
        "AWS":"*"
      },
      "Action":[
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission",
        "SNS:DeleteTopic"
      ],
      "Resource":"${aws_sns_topic.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccTopicPolicyEquivalentConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn = aws_sns_topic.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = %[1]q
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action = [
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission",
        "SNS:DeleteTopic",
      ]
      Resource = aws_sns_topic.test.arn
    }]
  })
}
`, rName)
}

func testAccTopicPolicyEquivalent2Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn = aws_sns_topic.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = %[1]q
      Effect = "Allow"
      Principal = {
        AWS = ["*"]
      }
      Action = [
        "SNS:SetTopicAttributes",
        "SNS:RemovePermission",
        "SNS:DeleteTopic",
        "SNS:AddPermission",
        "SNS:GetTopicAttributes",
      ]
      Resource = [aws_sns_topic.test.arn]
    }]
  })
}
`, rName)
}
