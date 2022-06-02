package emrserverless_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/emrserverless"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemrserverless "github.com/hashicorp/terraform-provider-aws/internal/service/emrserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEMRApplication_basic(t *testing.T) {
	var application emrserverless.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emrserverless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "emr-serverless", regexp.MustCompile(`/applications/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "hive"),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-6.6.0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccEMRApplication_disappears(t *testing.T) {
	var application emrserverless.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emrserverless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					acctest.CheckResourceDisappears(acctest.Provider, tfemrserverless.ResourceApplication(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfemrserverless.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRApplication_tags(t *testing.T) {
	var application emrserverless.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emrserverless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
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
				Config: testAccApplicationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckApplicationExists(resourceName string, application *emrserverless.Application) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRServerlessConn

		output, err := tfemrserverless.FindApplicationByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("EMR Serverless Application (%s) not found", rs.Primary.ID)
		}

		*application = *output

		return nil
	}
}

func testAccCheckApplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EMRServerlessConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emrserverless_application" {
			continue
		}

		_, err := tfemrserverless.FindApplicationByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EMR Serverless Application %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccApplicationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"
}
`, rName)
}

func testAccApplicationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  auth_mode                   = "SSO"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  engine_security_group_id    = aws_security_group.test.id
  name                        = %[1]q
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = [aws_subnet.test.id]
  user_role                   = aws_iam_role.test.arn
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccApplicationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  auth_mode                   = "SSO"
  default_s3_location         = "s3://${aws_s3_bucket.test.bucket}/test"
  engine_security_group_id    = aws_security_group.test.id
  name                        = %[1]q
  service_role                = aws_iam_role.test.arn
  subnet_ids                  = [aws_subnet.test.id]
  user_role                   = aws_iam_role.test.arn
  vpc_id                      = aws_vpc.test.id
  workspace_security_group_id = aws_security_group.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
