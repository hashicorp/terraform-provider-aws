// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "iot", fmt.Sprintf("policy/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
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

func TestAccIoTPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTPolicy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPolicyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPolicyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIoTPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// lintignore:AWSAT005
				Config: testAccPolicyConfig_resourceName(rName, fmt.Sprintf("arn:aws:iot:*:*:topic/%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// lintignore:AWSAT005
				Config: testAccPolicyConfig_resourceName(rName, fmt.Sprintf("arn:aws:iot:*:*:topic/%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccIoTPolicy_prune(t *testing.T) {
	ctx := acctest.Context(t)
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// lintignore:AWSAT005
				Config: testAccPolicyConfig_resourceName(rName, fmt.Sprintf("arn:aws:iot:*:*:topic/%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", acctest.Ct1),
					testAccCheckPolicyVersionIDs(ctx, resourceName, []string{acctest.Ct1}),
				),
			},
			{
				// lintignore:AWSAT005
				Config: testAccPolicyConfig_resourceName(rName, fmt.Sprintf("arn:aws:iot:*:*:topic/%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", acctest.Ct2),
					testAccCheckPolicyVersionIDs(ctx, resourceName, []string{acctest.Ct1, acctest.Ct2}),
				),
			},
			{
				// lintignore:AWSAT005
				Config: testAccPolicyConfig_resourceName(rName, fmt.Sprintf("arn:aws:iot:*:*:topic/%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", acctest.Ct3),
					testAccCheckPolicyVersionIDs(ctx, resourceName, []string{acctest.Ct1, acctest.Ct2, acctest.Ct3}),
				),
			},
			{
				// lintignore:AWSAT005
				Config: testAccPolicyConfig_resourceName(rName, fmt.Sprintf("arn:aws:iot:*:*:topic/%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", acctest.Ct4),
					testAccCheckPolicyVersionIDs(ctx, resourceName, []string{acctest.Ct1, acctest.Ct2, acctest.Ct3, acctest.Ct4}),
				),
			},
			{
				// lintignore:AWSAT005
				Config: testAccPolicyConfig_resourceName(rName, fmt.Sprintf("arn:aws:iot:*:*:topic/%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", "5"),
					testAccCheckPolicyVersionIDs(ctx, resourceName, []string{acctest.Ct1, acctest.Ct2, acctest.Ct3, acctest.Ct4, "5"}),
				),
			},
			{
				// lintignore:AWSAT005
				Config: testAccPolicyConfig_resourceName(rName, fmt.Sprintf("arn:aws:iot:*:*:topic/%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", "6"),
					testAccCheckPolicyVersionIDs(ctx, resourceName, []string{acctest.Ct2, acctest.Ct3, acctest.Ct4, "5", "6"}),
				),
			},
			{
				// lintignore:AWSAT005
				Config: testAccPolicyConfig_resourceName(rName, fmt.Sprintf("arn:aws:iot:*:*:topic/%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", "7"),
					testAccCheckPolicyVersionIDs(ctx, resourceName, []string{acctest.Ct3, acctest.Ct4, "5", "6", "7"}),
				),
			},
		},
	})
}

func testAccCheckPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_policy" {
				continue
			}

			_, err := tfiot.FindPolicyByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPolicyExists(ctx context.Context, n string, v *iot.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		output, err := tfiot.FindPolicyByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPolicyVersionIDs(ctx context.Context, n string, want []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		output, err := tfiot.FindPolicyVersionsByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		got := tfslices.ApplyToAll(output, func(v awstypes.PolicyVersion) string {
			return aws.ToString(v.VersionId)
		})

		if !cmp.Equal(got, want, cmpopts.SortSlices(func(i, j string) bool {
			return i < j
		})) {
			return fmt.Errorf("policy version IDs = %v, want = %v", got, want)
		}

		return nil
	}
}

func testAccPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccPolicyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }

  policy = <<EOF
  {
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Effect": "Allow",
		"Action": [
		  "iot:*"
		],
		"Resource": [
		  "*"
		]
	  }
	]
  }
  EOF  
}
`, rName, tagKey1, tagValue1)
}

func testAccPolicyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  policy = <<EOF
  {
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Effect": "Allow",
		"Action": [
		  "iot:*"
		],
		"Resource": [
		  "*"
		]
	  }
	]
  }
  EOF  
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPolicyConfig_resourceName(rName, resourceName string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        %[2]q
      ]
    }
  ]
}
EOF
}
`, rName, resourceName)
}
