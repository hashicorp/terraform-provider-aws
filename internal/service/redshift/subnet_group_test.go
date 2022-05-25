package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftSubnetGroup_basic(t *testing.T) {
	var v redshift.ClusterSubnetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description"},
			},
		},
	})
}

func TestAccRedshiftSubnetGroup_disappears(t *testing.T) {
	var clusterSubnetGroup redshift.ClusterSubnetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &clusterSubnetGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceSubnetGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftSubnetGroup_updateDescription(t *testing.T) {
	var v redshift.ClusterSubnetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
			{
				Config: testAccSubnetGroupConfig_updateDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "test description updated"),
				),
			},
		},
	})
}

func TestAccRedshiftSubnetGroup_updateSubnetIDs(t *testing.T) {
	var v redshift.ClusterSubnetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
			{
				Config: testAccSubnetGroupConfig_updateIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "3"),
				),
			},
		},
	})
}

func TestAccRedshiftSubnetGroup_tags(t *testing.T) {
	var v redshift.ClusterSubnetGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-redshift-subnetgroup"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description"},
			},
			{
				Config: testAccSubnetGroupConfig_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-redshift-subnetgroup"),
					resource.TestCheckResourceAttr(resourceName, "tags.test", "test2"),
				),
			},
		},
	})
}

func testAccCheckSubnetGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_subnet_group" {
			continue
		}

		_, err := tfredshift.FindSubnetGroupByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Subent Group %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSubnetGroupExists(n string, v *redshift.ClusterSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Subent Group is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		resp, err := tfredshift.FindSubnetGroupByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccSubnetGroupConfigBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccSubnetGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSubnetGroupConfigBase(rName), fmt.Sprintf(`
resource "aws_redshift_subnet_group" "test" {
  name        = %[1]q
  description = "test description"
  subnet_ids  = [aws_subnet.test.id, aws_subnet.test2.id]
}
`, rName))
}

func testAccSubnetGroupConfig_updateDescription(rName string) string {
	return acctest.ConfigCompose(testAccSubnetGroupConfigBase(rName), fmt.Sprintf(`
resource "aws_redshift_subnet_group" "test" {
  name        = %[1]q
  description = "test description updated"
  subnet_ids  = [aws_subnet.test.id, aws_subnet.test2.id]
}
`, rName))
}

func testAccSubnetGroupConfig_tags(rName string) string {
	return acctest.ConfigCompose(testAccSubnetGroupConfigBase(rName), fmt.Sprintf(`
resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.id, aws_subnet.test2.id]

  tags = {
    Name = "tf-redshift-subnetgroup"
  }
}
`, rName))
}

func testAccSubnetGroupConfig_tagsUpdated(rName string) string {
	return acctest.ConfigCompose(testAccSubnetGroupConfigBase(rName), fmt.Sprintf(`
resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.id, aws_subnet.test2.id]

  tags = {
    Name        = "tf-redshift-subnetgroup"
    environment = "production"
    test        = "test2"
  }
}
`, rName))
}

func testAccSubnetGroupConfig_updateIDs(rName string) string {
	return acctest.ConfigCompose(testAccSubnetGroupConfigBase(rName), fmt.Sprintf(`
resource "aws_subnet" "testtest2" {
  cidr_block        = "10.1.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test.id, aws_subnet.test2.id, aws_subnet.testtest2.id]
}
`, rName))
}
