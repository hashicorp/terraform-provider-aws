package comprehend_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/comprehend"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ComprehendClient()

	input := &comprehend.ListEntityRecognizersInput{}

	_, err := conn.ListEntityRecognizers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func configVPCWithSubnetsAndDNS(rName string, subnetCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = %[2]d

  vpc_id            = aws_vpc.test.id
  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName, subnetCount),
	)
}

func uniqueIDPattern() string {
	return prefixedUniqueIDPattern(resource.UniqueIdPrefix)
}

func prefixedUniqueIDPattern(prefix string) string {
	return fmt.Sprintf("%s[[:xdigit:]]{%d}", prefix, resource.UniqueIDSuffixLength)
}
