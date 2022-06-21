package route53recoveryreadiness_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53recoveryreadiness "github.com/hashicorp/terraform-provider-aws/internal/service/route53recoveryreadiness"
)

func TestAccRoute53RecoveryReadinessReadinessCheck_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rSetName := sdkacctest.RandomWithPrefix("tf-acc-test-set")
	resourceName := "aws_route53recoveryreadiness_readiness_check.test"
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReadinessCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReadinessCheckConfig_basic(rName, rSetName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReadinessCheckExists(resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`readiness-check/.+`)),
					resource.TestCheckResourceAttr(resourceName, "resource_set_name", rSetName),
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

func TestAccRoute53RecoveryReadinessReadinessCheck_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rSetName := sdkacctest.RandomWithPrefix("tf-acc-test-set")
	resourceName := "aws_route53recoveryreadiness_readiness_check.test"
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReadinessCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReadinessCheckConfig_basic(rName, rSetName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReadinessCheckExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53recoveryreadiness.ResourceReadinessCheck(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessReadinessCheck_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53recoveryreadiness_readiness_check.test"
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReadinessCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReadinessCheckConfig_tags1(rName, cwArn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReadinessCheckExists(resourceName),
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
				Config: testAccReadinessCheckConfig_tags2(rName, cwArn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReadinessCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccReadinessCheckConfig_tags1(rName, cwArn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReadinessCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRoute53RecoveryReadinessReadinessCheck_timeout(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rSetName := sdkacctest.RandomWithPrefix("tf-acc-test-set")
	resourceName := "aws_route53recoveryreadiness_readiness_check.test"
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReadinessCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReadinessCheckConfig_timeout(rName, rSetName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReadinessCheckExists(resourceName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "route53-recovery-readiness", regexp.MustCompile(`readiness-check/.+`)),
					resource.TestCheckResourceAttr(resourceName, "resource_set_name", rSetName),
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

func testAccCheckReadinessCheckDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53recoveryreadiness_readiness_check" {
			continue
		}

		input := &route53recoveryreadiness.GetReadinessCheckInput{
			ReadinessCheckName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetReadinessCheck(input)
		if err == nil {
			return fmt.Errorf("Route53RecoveryReadiness Readiness Check (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckReadinessCheckExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53RecoveryReadinessConn

		input := &route53recoveryreadiness.GetReadinessCheckInput{
			ReadinessCheckName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetReadinessCheck(input)

		return err
	}
}

func testAccReadinessCheckConfig_ResourceSet(rSetName, cwArn string) string {
	return fmt.Sprintf(`
resource "aws_route53recoveryreadiness_resource_set" "test" {
  resource_set_name = %[1]q
  resource_set_type = "AWS::CloudWatch::Alarm"

  resources {
    resource_arn = %[2]q
  }
}
`, rSetName, cwArn)
}

func testAccReadinessCheckConfig_basic(rName, rSetName, cwArn string) string {
	return acctest.ConfigCompose(testAccReadinessCheckConfig_ResourceSet(rSetName, cwArn), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_readiness_check" "test" {
  readiness_check_name = %q
  resource_set_name    = aws_route53recoveryreadiness_resource_set.test.resource_set_name
}
`, rName))
}

func testAccReadinessCheckConfig_tags1(rName, cwArn, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccReadinessCheckConfig_ResourceSet("resource-set-for-testing", cwArn), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_readiness_check" "test" {
  readiness_check_name = %[1]q
  resource_set_name    = aws_route53recoveryreadiness_resource_set.test.resource_set_name

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccReadinessCheckConfig_tags2(rName, cwArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccReadinessCheckConfig_ResourceSet("resource-set-for-testing", cwArn), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_readiness_check" "test" {
  readiness_check_name = %[1]q
  resource_set_name    = aws_route53recoveryreadiness_resource_set.test.resource_set_name

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccReadinessCheckConfig_timeout(rName, rSetName, cwArn string) string {
	return acctest.ConfigCompose(testAccReadinessCheckConfig_ResourceSet(rSetName, cwArn), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_readiness_check" "test" {
  readiness_check_name = %q
  resource_set_name    = aws_route53recoveryreadiness_resource_set.test.resource_set_name

  timeouts {
    delete = "10m"
  }
}
`, rName))
}
