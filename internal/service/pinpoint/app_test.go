package pinpoint_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccPinpointApp_basic(t *testing.T) {
	var application pinpoint.ApplicationResponse
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_withGeneratedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
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

func TestAccPinpointApp_campaignHookLambda(t *testing.T) {
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_CampaignHookLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "campaign_hook.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "campaign_hook.0.mode", "DELIVERY"),
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

func TestAccPinpointApp_limits(t *testing.T) {
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_Limits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "limits.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "limits.0.total", "100"),
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

func TestAccPinpointApp_quietTime(t *testing.T) {
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_QuietTime(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "quiet_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quiet_time.0.start", "00:00"),
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

func TestAccPinpointApp_tags(t *testing.T) {
	var application pinpoint.ApplicationResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpoint_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckApp(t) },
		ErrorCheck:        acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRAMResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_Tag1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppConfig_Tag2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppConfig_Tag1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccPreCheckApp(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

	input := &pinpoint.GetAppsInput{}

	_, err := conn.GetApps(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAppExists(n string, application *pinpoint.ApplicationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint app with that ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

		// Check if the app exists
		params := &pinpoint.GetAppInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetApp(params)

		if err != nil {
			return err
		}

		*application = *output.ApplicationResponse

		return nil
	}
}

const testAccAppConfig_withGeneratedName = `
resource "aws_pinpoint_app" "test" {}
`

func testAccAppConfig_CampaignHookLambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  campaign_hook {
    lambda_function_name = aws_lambda_function.test.arn
    mode                 = "DELIVERY"
  }

  depends_on = [aws_lambda_permission.test]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdapinpoint.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambdapinpoint.handler"
  runtime       = "nodejs12.x"
  publish       = true
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "aws" {}

data "aws_region" "current" {}

resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromPinpoint"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "pinpoint.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  source_arn    = "arn:${data.aws_partition.current.partition}:mobiletargeting:${data.aws_region.current.name}:${data.aws_caller_identity.aws.account_id}:/apps/*"
}
`, rName)
}

func testAccAppConfig_Limits(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  limits {
    daily               = 3
    maximum_duration    = 600
    messages_per_second = 50
    total               = 100
  }
}
`, rName)
}

func testAccAppConfig_QuietTime(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  quiet_time {
    start = "00:00"
    end   = "03:00"
  }
}
`, rName)
}

func testAccAppConfig_Tag1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppConfig_Tag2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCheckAppDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_app" {
			continue
		}

		// Check if the topic exists by fetching its attributes
		params := &pinpoint.GetAppInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetApp(params)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
				continue
			}
			return err
		}
		return fmt.Errorf("App exists when it should be destroyed!")
	}

	return nil
}

func testAccCheckRAMResourceShareDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_resource_share" {
			continue
		}

		request := &ram.GetResourceSharesInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.ID)},
			ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
		}

		output, err := conn.GetResourceShares(request)
		if err != nil {
			return err
		}

		if len(output.ResourceShares) > 0 {
			resourceShare := output.ResourceShares[0]
			if aws.StringValue(resourceShare.Status) != ram.ResourceShareStatusDeleted {
				return fmt.Errorf("RAM resource share (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}
