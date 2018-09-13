package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsAmiIds_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAmiIdsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID("data.aws_ami_ids.ubuntu"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAmiIds_sorted(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAmiIdsConfig_sorted1(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_ami_from_instance.a", "id"),
					resource.TestCheckResourceAttrSet("aws_ami_from_instance.b", "id"),
				),
			},
			{
				Config: testAccDataSourceAwsAmiIdsConfig_sorted2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEbsSnapshotDataSourceID("data.aws_ami_ids.test"),
					resource.TestCheckResourceAttr("data.aws_ami_ids.test", "ids.#", "2"),
					resource.TestCheckResourceAttrPair(
						"data.aws_ami_ids.test", "ids.0",
						"aws_ami_from_instance.b", "id"),
					resource.TestCheckResourceAttrPair(
						"data.aws_ami_ids.test", "ids.1",
						"aws_ami_from_instance.a", "id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAmiIds_empty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAmiIdsConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID("data.aws_ami_ids.empty"),
					resource.TestCheckResourceAttr("data.aws_ami_ids.empty", "ids.#", "0"),
				),
			},
		},
	})
}

const testAccDataSourceAwsAmiIdsConfig_basic = `
data "aws_ami_ids" "ubuntu" {
    owners = ["099720109477"]

    filter {
        name   = "name"
        values = ["ubuntu/images/ubuntu-*-*-amd64-server-*"]
    }
}
`

func testAccDataSourceAwsAmiIdsConfig_sorted1(rName string) string {
	return fmt.Sprintf(`
resource "aws_instance" "test" {
    ami           = "ami-efd0428f"
    instance_type = "m3.medium"

    count = 2
}

resource "aws_ami_from_instance" "a" {
    name                    = "%s-a"
    source_instance_id      = "${aws_instance.test.*.id[0]}"
    snapshot_without_reboot = true
}

resource "aws_ami_from_instance" "b" {
    name                    = "%s-b"
    source_instance_id      = "${aws_instance.test.*.id[1]}"
    snapshot_without_reboot = true

    // We want to ensure that 'aws_ami_from_instance.a.creation_date' is less
    // than 'aws_ami_from_instance.b.creation_date' so that we can ensure that
    // the images are being sorted correctly.
    depends_on = ["aws_ami_from_instance.a"]
}
`, rName, rName)
}

func testAccDataSourceAwsAmiIdsConfig_sorted2(rName string) string {
	return testAccDataSourceAwsAmiIdsConfig_sorted1(rName) + fmt.Sprintf(`
data "aws_ami_ids" "test" {
  owners     = ["self"]
  name_regex = "^%s-"
}
`, rName)
}

const testAccDataSourceAwsAmiIdsConfig_empty = `
data "aws_ami_ids" "empty" {
  filter {
    name   = "name"
    values = []
  }
}
`
