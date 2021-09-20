package aws

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
)

func TestAccAwsRoute53RecoveryReadinessReadinessCheck_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
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
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessReadinessCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessReadinessCheckConfig(rName, rSetName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessReadinessCheckExists(resourceName),
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

func TestAccAwsRoute53RecoveryReadinessReadinessCheck_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
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
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessReadinessCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessReadinessCheckConfig(rName, rSetName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessReadinessCheckExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsRoute53RecoveryReadinessReadinessCheck(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsRoute53RecoveryReadinessReadinessCheck_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53recoveryreadiness_readiness_check.test"
	cwArn := arn.ARN{
		AccountID: "123456789012",
		Partition: endpoints.AwsPartitionID,
		Region:    endpoints.EuWest1RegionID,
		Resource:  "alarm:zzzzzzzzz",
		Service:   "cloudwatch",
	}.String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessReadinessCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_Tags1(rName, cwArn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessReadinessCheckExists(resourceName),
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
				Config: testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_Tags2(rName, cwArn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessReadinessCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_Tags1(rName, cwArn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessReadinessCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsRoute53RecoveryReadinessReadinessCheck_timeout(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
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
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAwsRoute53RecoveryReadiness(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53recoveryreadiness.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53RecoveryReadinessReadinessCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_Timeout(rName, rSetName, cwArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53RecoveryReadinessReadinessCheckExists(resourceName),
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

func testAccCheckAwsRoute53RecoveryReadinessReadinessCheckDestroy(s *terraform.State) error {
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

func testAccCheckAwsRoute53RecoveryReadinessReadinessCheckExists(name string) resource.TestCheckFunc {
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

func testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_ResourceSet(rSetName, cwArn string) string {
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

func testAccAwsRoute53RecoveryReadinessReadinessCheckConfig(rName, rSetName, cwArn string) string {
	return acctest.ConfigCompose(testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_ResourceSet(rSetName, cwArn), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_readiness_check" "test" {
  readiness_check_name = %q
  resource_set_name    = aws_route53recoveryreadiness_resource_set.test.resource_set_name
}
`, rName))
}

func testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_Tags1(rName, cwArn, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_ResourceSet("resource-set-for-testing", cwArn), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_readiness_check" "test" {
  readiness_check_name = %[1]q
  resource_set_name    = aws_route53recoveryreadiness_resource_set.test.resource_set_name

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_Tags2(rName, cwArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_ResourceSet("resource-set-for-testing", cwArn), fmt.Sprintf(`
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

func testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_Timeout(rName, rSetName, cwArn string) string {
	return acctest.ConfigCompose(testAccAwsRoute53RecoveryReadinessReadinessCheckConfig_ResourceSet(rSetName, cwArn), fmt.Sprintf(`
resource "aws_route53recoveryreadiness_readiness_check" "test" {
  readiness_check_name = %q
  resource_set_name    = aws_route53recoveryreadiness_resource_set.test.resource_set_name

  timeouts {
    delete = "10m"
  }
}
`, rName))
}
