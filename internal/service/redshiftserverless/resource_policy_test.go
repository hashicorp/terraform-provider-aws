package redshiftserverless_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftServerlessResourcePolicy_basic(t *testing.T) {
	resourceName := "aws_redshiftserverless_resource_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_redshiftserverless_workgroup.test", "arn"),
					// resource.TestCheckResourceAttr(resourceName, "amount", "60"),
					// resource.TestCheckResourceAttr(resourceName, "usage_type", "serverless-compute"),
					// resource.TestCheckResourceAttr(resourceName, "breach_action", "log"),
					// resource.TestCheckResourceAttr(resourceName, "period", "monthly"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// {
			// 	Config: testAccResourcePolicyConfig_basic(rName, 120),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckResourcePolicyExists(resourceName),
			// 		resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_redshiftserverless_workgroup.test", "arn"),
			// 		resource.TestCheckResourceAttr(resourceName, "amount", "120"),
			// 	),
			// },
		},
	})
}

func TestAccRedshiftServerlessResourcePolicy_disappears(t *testing.T) {
	resourceName := "aws_redshiftserverless_resource_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, redshiftserverless.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshiftserverless.ResourceResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshiftserverless_resource_policy" {
			continue
		}
		_, err := tfredshiftserverless.FindResourcePolicyByArn(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Serverless Resource Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckResourcePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Serverless Resource Policy is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftServerlessConn

		_, err := tfredshiftserverless.FindResourcePolicyByArn(conn, rs.Primary.ID)

		return err
	}
}

func testAccResourcePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_resource_policy" "test" {
  resource_arn = aws_redshiftserverless_namespace.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = %[1]q
    Statement = [{
      Sid    = %[1]q
      Effect = "Allow"
      Principal = {
        AWS = ["*"]
      }
      Action = [
        "redshift-serverless:RestoreFromSnapshot",
      ]
      Resource = [aws_redshiftserverless_namespace.test.arn]
    }]
  })
}
`, rName)
}
