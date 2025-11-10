// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{id}"),
					resource.TestCheckResourceAttr(resourceName, "auto_resolve_best_voices_enabled", acctest.CtTrue), //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttr(resourceName, "contact_flow_logs_enabled", acctest.CtFalse),       //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttr(resourceName, "contact_lens_enabled", acctest.CtTrue),             //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "early_media_enabled", acctest.CtTrue), //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", string(awstypes.DirectoryTypeConnectManaged)),
					resource.TestCheckResourceAttr(resourceName, "inbound_calls_enabled", acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexache.MustCompile(rName)),
					resource.TestCheckResourceAttr(resourceName, "multi_party_conference_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outbound_calls_enabled", acctest.CtTrue),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrServiceRole, "iam", regexache.MustCompile(`role/aws-service-role/connect.amazonaws.com/AWSServiceRoleForAmazonConnect_[A-Za-z0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.InstanceStatusActive)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_basicFlipped(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{id}"),
					resource.TestCheckResourceAttr(resourceName, "auto_resolve_best_voices_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "contact_flow_logs_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "contact_lens_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "early_media_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "inbound_calls_enabled", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexache.MustCompile(rName)),
					resource.TestCheckResourceAttr(resourceName, "multi_party_conference_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outbound_calls_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.InstanceStatusActive)),
				),
			},
		},
	})
}

func testAccInstance_directory(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.test"

	domainName := acctest.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_directory(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", string(awstypes.DirectoryTypeExistingDirectory)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.InstanceStatusActive)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"directory_id"},
			},
		},
	})
}

func testAccInstance_saml(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_saml(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", string(awstypes.DirectoryTypeSaml)),
					testAccCheckInstanceExists(ctx, resourceName, &v),
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

func testAccInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccInstanceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckInstanceExists(ctx context.Context, n string, v *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

		output, err := tfconnect.FindInstanceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_instance" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

			_, err := tfconnect.FindInstanceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInstanceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccInstanceConfig_tags1(rName string, tag, value string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag, value)
}

func testAccInstanceConfig_tags2(rName string, tag1, value1, tag2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1, value1, tag2, value2)
}

func testAccInstanceConfig_basicFlipped(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  auto_resolve_best_voices_enabled = false
  contact_flow_logs_enabled        = true
  contact_lens_enabled             = false
  early_media_enabled              = false
  identity_management_type         = "CONNECT_MANAGED"
  inbound_calls_enabled            = false
  instance_alias                   = %[1]q
  multi_party_conference_enabled   = false
  outbound_calls_enabled           = false
}
`, rName)
}

func testAccInstanceConfig_directory(rName, domain string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"
  tags = {
    Name = %[1]q
  }
}

resource "aws_directory_service_directory" "test" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}

resource "aws_connect_instance" "test" {
  directory_id             = aws_directory_service_directory.test.id
  identity_management_type = "EXISTING_DIRECTORY"
  instance_alias           = %[1]q
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true
}
`, rName, domain)
}

func testAccInstanceConfig_saml(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "SAML"
  instance_alias           = %[1]q
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true
}
`, rName)
}
