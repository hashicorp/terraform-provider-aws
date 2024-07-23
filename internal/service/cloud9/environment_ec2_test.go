// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloud9_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cloud9/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloud9 "github.com/hashicorp/terraform-provider-aws/internal/service/cloud9"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloud9EnvironmentEC2_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Environment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Cloud9EndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Cloud9ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentEC2Destroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2Config_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cloud9", regexache.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "CONNECT_SSH"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrInstanceType, names.AttrSubnetID, "image_id"},
			},
		},
	})
}

func TestAccCloud9EnvironmentEC2_allFields(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Environment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	name1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	name2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	imageID := "ubuntu-22.04-x86_64"
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Cloud9EndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Cloud9ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentEC2Destroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2Config_allFields(rName, name1, description1, imageID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "automatic_stop_time_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "CONNECT_SSH"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description1),
					resource.TestCheckResourceAttr(resourceName, "image_id", imageID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name1),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "aws_iam_user.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, "aws_subnet.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ec2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"automatic_stop_time_minutes", "image_id", names.AttrInstanceType, names.AttrSubnetID},
			},
			{
				Config: testAccEnvironmentEC2Config_allFields(rName, name2, description2, imageID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name2),
				),
			},
		},
	})
}

func TestAccCloud9EnvironmentEC2_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Environment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Cloud9EndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Cloud9ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentEC2Destroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2Config_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrInstanceType, names.AttrSubnetID, "image_id"},
			},
			{
				Config: testAccEnvironmentEC2Config_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEnvironmentEC2Config_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCloud9EnvironmentEC2_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf types.Environment
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Cloud9EndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Cloud9ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentEC2Destroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentEC2Config_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentEC2Exists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloud9.ResourceEnvironmentEC2(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloud9.ResourceEnvironmentEC2(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEnvironmentEC2Exists(ctx context.Context, n string, v *types.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Cloud9Client(ctx)

		output, err := tfcloud9.FindEnvironmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEnvironmentEC2Destroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Cloud9Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloud9_environment_ec2" {
				continue
			}

			_, err := tfcloud9.FindEnvironmentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cloud9 Environment EC2 %s still exists.", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEnvironmentEC2Config_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}
`, rName))
}

func testAccEnvironmentEC2Config_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentEC2Config_base(rName), fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test[0].id
  image_id      = "amazonlinux-2023-x86_64"
}
`, rName))
}

func testAccEnvironmentEC2Config_allFields(rName, name, description, imageID string) string {
	return acctest.ConfigCompose(testAccEnvironmentEC2Config_base(rName), fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  automatic_stop_time_minutes = 60
  description                 = %[2]q
  instance_type               = "t2.micro"
  name                        = %[1]q
  owner_arn                   = aws_iam_user.test.arn
  subnet_id                   = aws_subnet.test[0].id
  connection_type             = "CONNECT_SSH"
  image_id                    = %[4]q
}

resource "aws_iam_user" "test" {
  name = %[3]q
}
`, name, description, rName, imageID))
}

func testAccEnvironmentEC2Config_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEnvironmentEC2Config_base(rName), fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test[0].id
  image_id      = "amazonlinux-2023-x86_64"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccEnvironmentEC2Config_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEnvironmentEC2Config_base(rName), fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test[0].id
  image_id      = "amazonlinux-2023-x86_64"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
