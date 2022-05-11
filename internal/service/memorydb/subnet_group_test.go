package memorydb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccPreCheck(t *testing.T) {
	// Checking for service fails in all partitions!
	// acctest.PreCheckPartitionHasService(memorydb.EndpointsID, t)

	if got, want := acctest.Partition(), endpoints.AwsUsGovPartitionID; got == want {
		t.Skipf("MemoryDB is not supported in %s partition", got)
	}
}

func TestAccMemoryDBSubnetGroup_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "memorydb", "subnetgroup/"+rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.1", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
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

func TestAccMemoryDBSubnetGroup_disappears(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfmemorydb.ResourceSubnetGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBSubnetGroup_nameGenerated(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_withNoName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
		},
	})
}

func TestAccMemoryDBSubnetGroup_namePrefix(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_withNamePrefix(rName, "tftest-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tftest-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tftest-"),
				),
			},
		},
	})
}

func TestAccMemoryDBSubnetGroup_update_description(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubnetGroupConfig_withDescription(rName, "Updated Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated Description"),
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

func TestAccMemoryDBSubnetGroup_update_subnetIds(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_withSubnetCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubnetGroupConfig_withSubnetCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "3"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.1", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.2", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubnetGroupConfig_withSubnetCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_ids.*", "aws_subnet.test.1", "id"),
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

func TestAccMemoryDBSubnetGroup_update_tags(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_withTags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubnetGroupConfig_withTags2(rName, "Key1", "value1", "Key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubnetGroupConfig_withTags1(rName, "Key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSubnetGroupConfig_withTags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
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

func testAccCheckSubnetGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_memorydb_subnet_group" {
			continue
		}

		_, err := tfmemorydb.FindSubnetGroupByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("MemoryDB Subnet Group %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSubnetGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB Subnet Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

		_, err := tfmemorydb.FindSubnetGroupByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccSubnetGroupConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test.*.id

  tags = {
    Test = "test"
  }
}
`, rName),
	)
}

func testAccSubnetGroupConfig_withNoName(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		`
resource "aws_memorydb_subnet_group" "test" {
  subnet_ids = aws_subnet.test.*.id
}
`,
	)
}

func testAccSubnetGroupConfig_withNamePrefix(rName, rNamePrefix string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  name_prefix = %[1]q
  subnet_ids  = aws_subnet.test.*.id
}
`, rNamePrefix),
	)
}

func testAccSubnetGroupConfig_withDescription(rName, description string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  name        = %[1]q
  subnet_ids  = aws_subnet.test.*.id
  description = %[2]q
}
`, rName, description),
	)
}

func testAccSubnetGroupConfig_withSubnetCount(rName string, subnetCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, subnetCount),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test.*.id
}
`, rName),
	)
}

func testAccSubnetGroupConfig_withTags0(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test.*.id
}
`, rName),
	)
}

func testAccSubnetGroupConfig_withTags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test.*.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value),
	)
}

func testAccSubnetGroupConfig_withTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_memorydb_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test.*.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value),
	)
}
