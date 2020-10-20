package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccPreCheckAWSSSOInstance(t *testing.T) {
	ssoadminconn := testAccProvider.Meta().(*AWSClient).ssoadminconn

	input := &ssoadmin.ListInstancesInput{}

	_, err := ssoadminconn.ListInstances(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func TestAccDataSourceAwsSsoInstanceBasic(t *testing.T) {
	datasourceName := "data.aws_sso_instance.selected"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSSSOInstance(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSsoInstanceConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrAwsSsoARN(datasourceName, "arn", regexp.MustCompile("instance/ssoins-[a-zA-Z0-9-.]{16}")),
					resource.TestMatchResourceAttr(datasourceName, "identity_store_id", regexp.MustCompile("^[a-zA-Z0-9-]*")),
				),
			},
		},
	})
}

func testAccDataSourceAwsSsoInstanceConfigBasic() string {
	return `data "aws_sso_instance" "selected" {}`
}

func testAccMatchResourceAttrAwsSsoARN(resourceName, attributeName string, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arnRegexp := arn.ARN{
			Partition: testAccGetPartition(),
			Resource:  arnResourceRegexp.String(),
			Service:   "sso",
		}.String()

		attributeMatch, err := regexp.Compile(arnRegexp)

		if err != nil {
			return fmt.Errorf("Unable to compile ARN regexp (%s): %s", arnRegexp, err)
		}

		return resource.TestMatchResourceAttr(resourceName, attributeName, attributeMatch)(s)
	}
}
