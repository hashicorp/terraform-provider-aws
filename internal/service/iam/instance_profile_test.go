// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMInstanceProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.InstanceProfile
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("instance-profile/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, "aws_iam_role.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccIAMInstanceProfile_withoutRole(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.InstanceProfile
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_noRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrRole),
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

func TestAccIAMInstanceProfile_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.InstanceProfile
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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

func TestAccIAMInstanceProfile_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.InstanceProfile
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccIAMInstanceProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.InstanceProfile
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceInstanceProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMInstanceProfile_Disappears_role(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.InstanceProfile
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceRole(), "aws_iam_role.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMInstanceProfile_launchConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.InstanceProfile
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfileConfig_launchConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("instance-profile/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, "aws_iam_role.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceProfileConfig_launchConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("instance-profile/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, "aws_iam_role.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccInstanceProfileConfig_launchConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("instance-profile/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, "aws_iam_role.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccInstanceProfileConfig_launchConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("instance-profile/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, "aws_iam_role.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceProfileConfig_launchConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceProfileExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("instance-profile/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRole, "aws_iam_role.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func testAccCheckInstanceProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_instance_profile" {
				continue
			}

			_, err := tfiam.FindInstanceProfileByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Instance Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckInstanceProfileExists(ctx context.Context, n string, v *awstypes.InstanceProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM Instance Profile ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		output, err := tfiam.FindInstanceProfileByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccInstanceProfileConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%[1]s-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccInstanceProfileConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccInstanceProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}
`, rName))
}

func testAccInstanceProfileConfig_noRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = %[1]q
}
`, rName)
}

func testAccInstanceProfileConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(testAccInstanceProfileConfig_base(rName), `
resource "aws_iam_instance_profile" "test" {
  role = aws_iam_role.test.name
}
`)
}

func testAccInstanceProfileConfig_namePrefix(rName, namePrefix string) string {
	return acctest.ConfigCompose(testAccInstanceProfileConfig_base(rName), fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name_prefix = %[1]q
  role        = aws_iam_role.test.name
}
`, namePrefix))
}

func testAccInstanceProfileConfig_launchConfiguration(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceProfileConfig_base(rName),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name                 = %[1]q
  iam_instance_profile = aws_iam_instance_profile.test.name
  image_id             = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type        = "t2.micro"
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}
`, rName))
}
