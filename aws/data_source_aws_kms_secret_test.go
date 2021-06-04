package aws

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/atest"
)

func TestAccAWSKmsSecretDataSource_removed(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { atest.PreCheck(t) },
		ErrorCheck: atest.ErrorCheck(t, kms.EndpointsID),
		Providers:  atest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsKmsSecretDataSourceConfig,
				ExpectError: regexp.MustCompile(dataSourceAwsKmsSecretRemovedMessage),
			},
		},
	})
}

const testAccAwsKmsSecretDataSourceConfig = `
data "aws_kms_secret" "testing" {
  secret {
    name    = "secret_name"
    payload = "data-source-removed"

    context = {
      name = "value"
    }
  }
}
`
