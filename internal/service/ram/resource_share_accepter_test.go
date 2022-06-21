package ram_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccRAMResourceShareAccepter_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_ram_resource_share_accepter.test"
	principalAssociationResourceName := "aws_ram_principal_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckResourceShareAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAccepterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAccepterExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "share_arn", principalAssociationResourceName, "resource_share_arn"),
					acctest.MatchResourceAttrRegionalARNAccountID(resourceName, "invitation_arn", "ram", `\d{12}`, regexp.MustCompile(fmt.Sprintf("resource-share-invitation/%s$", verify.UUIDRegexPattern))),
					resource.TestMatchResourceAttr(resourceName, "share_id", regexp.MustCompile(fmt.Sprintf(`^rs-%s$`, verify.UUIDRegexPattern))),
					resource.TestCheckResourceAttr(resourceName, "status", ram.ResourceShareStatusActive),
					acctest.CheckResourceAttrAccountID(resourceName, "receiver_account_id"),
					resource.TestMatchResourceAttr(resourceName, "sender_account_id", regexp.MustCompile(`\d{12}`)),
					resource.TestCheckResourceAttr(resourceName, "share_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resources.%", "0"),
				),
			},
			{
				Config:            testAccResourceShareAccepterConfig_basic(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRAMResourceShareAccepter_disappears(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_ram_resource_share_accepter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckResourceShareAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAccepterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAccepterExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfram.ResourceResourceShareAccepter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRAMResourceShareAccepter_resourceAssociation(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_ram_resource_share_accepter.test"
	principalAssociationResourceName := "aws_ram_principal_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckResourceShareAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAccepterConfig_association(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAccepterExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "share_arn", principalAssociationResourceName, "resource_share_arn"),
					acctest.MatchResourceAttrRegionalARNAccountID(resourceName, "invitation_arn", "ram", `\d{12}`, regexp.MustCompile(fmt.Sprintf("resource-share-invitation/%s$", verify.UUIDRegexPattern))),
					resource.TestMatchResourceAttr(resourceName, "share_id", regexp.MustCompile(fmt.Sprintf(`^rs-%s$`, verify.UUIDRegexPattern))),
					resource.TestCheckResourceAttr(resourceName, "status", ram.ResourceShareStatusActive),
					acctest.CheckResourceAttrAccountID(resourceName, "receiver_account_id"),
					resource.TestMatchResourceAttr(resourceName, "sender_account_id", regexp.MustCompile(`\d{12}`)),
					resource.TestCheckResourceAttr(resourceName, "share_name", rName),
					resource.TestCheckResourceAttr(resourceName, "resources.%", "0"),
				),
			},
			{
				Config:            testAccResourceShareAccepterConfig_association(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckResourceShareAccepterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_resource_share_accepter" {
			continue
		}

		input := &ram.GetResourceSharesInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.Attributes["share_arn"])},
			ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
		}

		output, err := conn.GetResourceShares(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
				return nil
			}
			return fmt.Errorf("Error deleting RAM resource share: %s", err)
		}

		if len(output.ResourceShares) > 0 && aws.StringValue(output.ResourceShares[0].Status) != ram.ResourceShareStatusDeleted {
			return fmt.Errorf("RAM resource share invitation found, should be destroyed")
		}
	}

	return nil
}

func testAccCheckResourceShareAccepterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok || rs.Type != "aws_ram_resource_share_accepter" {
			return fmt.Errorf("RAM resource share invitation not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

		input := &ram.GetResourceSharesInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.Attributes["share_arn"])},
			ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
		}

		output, err := conn.GetResourceShares(input)
		if err != nil || len(output.ResourceShares) == 0 {
			return fmt.Errorf("Error finding RAM resource share: %s", err)
		}

		return nil
	}
}

func testAccResourceShareAccepterConfig_basic(rName string) string {
	return acctest.ConfigAlternateAccountProvider() + fmt.Sprintf(`
resource "aws_ram_resource_share_accepter" "test" {
  share_arn = aws_ram_principal_association.test.resource_share_arn
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name                      = %[1]q
  allow_external_principals = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_caller_identity.receiver.account_id
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_caller_identity" "receiver" {}
`, rName)
}

func testAccResourceShareAccepterConfig_association(rName string) string {
	return acctest.ConfigCompose(testAccResourceShareAccepterConfig_basic(rName), fmt.Sprintf(`
resource "aws_ram_resource_association" "test" {
  provider = "awsalternate"

  resource_arn       = aws_codebuild_project.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_codebuild_project" "test" {
  provider = "awsalternate"

  name         = %[1]q
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  provider = "awsalternate"

  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "codebuild.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy" "test" {
  provider = "awsalternate"

  name = %[1]q
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Resource = ["*"]
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    }]
  })
}
`, rName))
}
