package servicecatalog_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
)

// add sweeper to delete known test servicecat service actions

func TestAccServiceCatalogServiceAction_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_service_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "definition.0.name", "AWS-RestartEC2Instance"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
				},
			},
		},
	})
}

func TestAccServiceCatalogServiceAction_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_service_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfservicecatalog.ResourceServiceAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogServiceAction_update(t *testing.T) {
	resourceName := "aws_servicecatalog_service_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "definition.0.name", "AWS-RestartEC2Instance"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccServiceActionConfig_update(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.assume_role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", rName2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			}},
	})
}

func testAccCheckServiceActionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_service_action" {
			continue
		}

		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeServiceAction(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Service Action (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Service Action (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckServiceActionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeServiceAction(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Service Action (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccServiceActionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_service_action" "test" {
  accept_language = "en"
  description     = %[1]q
  name            = %[1]q

  definition {
    name    = "AWS-RestartEC2Instance"
    version = "1"
  }
}
`, rName)
}

func testAccServiceActionConfig_update(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    principals {
      type = "Service"

      identifiers = [
        "servicecatalog.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}",
      ]
    }

    actions = [
      "sts:AssumeRole",
    ]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_servicecatalog_service_action" "test" {
  description = %[1]q
  name        = %[1]q

  definition {
    assume_role = aws_iam_role.test.arn
    name        = "AWS-RestartEC2Instance"
    version     = "1"
  }
}
`, rName)
}
