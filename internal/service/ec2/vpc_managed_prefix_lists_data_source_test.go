package ec2_test

import (
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCManagedPrefixListsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ec2_managed_prefix_lists.aws", "ids.#", "4"),
				),
			},
		},
	})
}

const testAccVPCManagedPrefixListsDataSourceConfig_basic = `
data "aws_ec2_managed_prefix_lists" "aws" {
  filter {
    name   = "prefix-list-name"
    values = ["com.amazonaws.*"]
  }
}
`

func TestAccVPCManagedPrefixListsDataSource_filter_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tagKey := "key1"
	tagValue := "value1"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListConfig_tags1(rName, tagKey, tagValue),
			},
			{
				Config: testAccVPCManagedPrefixListsDataSourceConfig_filter_tags(tagKey, tagValue),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ec2_managed_prefix_lists.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccVPCManagedPrefixListsDataSourceConfig_filter_tags(tagKey string, tagValue string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_ec2_managed_prefix_lists" "test" {
  	tags = {
		%[1]q = %[2]q
	}
}
`, tagKey, tagValue))
}

func TestAccVPCManagedPrefixListsDataSource_noMatches(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListsDataSourceConfig_noMatches,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ec2_managed_prefix_lists.empty", "ids.#", "0"),
				),
			},
		},
	})
}

const testAccVPCManagedPrefixListsDataSourceConfig_noMatches = `
data "aws_ec2_managed_prefix_lists" "empty" {
  filter {
    name   = "prefix-list-name"
    values = ["no-match"]
  }
}
`
