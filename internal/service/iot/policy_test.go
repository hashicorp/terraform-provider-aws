package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
)

func TestAccIoTPolicy_basic(t *testing.T) {
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyInitialStateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("policy/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "default_version_id"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
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

func TestAccIoTPolicy_disappears(t *testing.T) {
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyInitialStateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfiot.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPolicyDestroy_basic(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_policy" {
			continue
		}

		// Try to find the Policy
		GetPolicyOpts := &iot.GetPolicyInput{
			PolicyName: aws.String(rs.Primary.Attributes["name"]),
		}

		resp, err := conn.GetPolicy(GetPolicyOpts)

		if err == nil {
			if resp.PolicyName != nil {
				return fmt.Errorf("IoT Policy still exists")
			}
		}

		// Verify the error is what we want
		if err != nil {
			iotErr, ok := err.(awserr.Error)
			if !ok || iotErr.Code() != "ResourceNotFoundException" {
				return err
			}
		}
	}

	return nil
}

func testAccCheckPolicyExists(n string, v *iot.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		resp, err := conn.GetPolicy(&iot.GetPolicyInput{
			PolicyName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPolicyInitialStateConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "test" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}
`, rName)
}
