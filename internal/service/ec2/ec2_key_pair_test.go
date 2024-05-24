// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2KeyPair_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var keyPair types.KeyPairInfo
	resourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_basic(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(ctx, resourceName, &keyPair),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", fmt.Sprintf("key-pair/%s", rName)),
					resource.TestMatchResourceAttr(resourceName, "fingerprint", regexache.MustCompile(`[0-9a-f]{2}(:[0-9a-f]{2}){15}`)),
					resource.TestCheckResourceAttr(resourceName, "key_name", rName),
					resource.TestCheckResourceAttr(resourceName, "key_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPublicKey, publicKey),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPublicKey},
			},
		},
	})
}

func TestAccEC2KeyPair_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var keyPair types.KeyPairInfo
	resourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_tags1(rName, publicKey, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(ctx, resourceName, &keyPair),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPublicKey},
			},
			{
				Config: testAccKeyPairConfig_tags2(rName, publicKey, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(ctx, resourceName, &keyPair),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccKeyPairConfig_tags1(rName, publicKey, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(ctx, resourceName, &keyPair),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2KeyPair_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var keyPair types.KeyPairInfo
	resourceName := "aws_key_pair.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_nameGenerated(publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(ctx, resourceName, &keyPair),
					acctest.CheckResourceAttrNameGenerated(resourceName, "key_name"),
					resource.TestCheckResourceAttr(resourceName, "key_name_prefix", "terraform-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPublicKey},
			},
		},
	})
}

func TestAccEC2KeyPair_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var keyPair types.KeyPairInfo
	resourceName := "aws_key_pair.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_namePrefix("tf-acc-test-prefix-", publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(ctx, resourceName, &keyPair),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "key_name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "key_name_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPublicKey},
			},
		},
	})
}

func TestAccEC2KeyPair_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var keyPair types.KeyPairInfo
	resourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyPairDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairConfig_basic(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyPairExists(ctx, resourceName, &keyPair),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceKeyPair(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckKeyPairDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_key_pair" {
				continue
			}

			_, err := tfec2.FindKeyPairByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Key Pair %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKeyPairExists(ctx context.Context, n string, v *types.KeyPairInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindKeyPairByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccKeyPairConfig_basic(rName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}
`, rName, publicKey)
}

func testAccKeyPairConfig_tags1(rName, publicKey, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, publicKey, tagKey1, tagValue1)
}

func testAccKeyPairConfig_tags2(rName, publicKey, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, publicKey, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccKeyPairConfig_nameGenerated(publicKey string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  public_key = %[1]q
}
`, publicKey)
}

func testAccKeyPairConfig_namePrefix(namePrefix, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name_prefix = %[1]q
  public_key      = %[2]q
}
`, namePrefix, publicKey)
}
