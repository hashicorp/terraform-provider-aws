// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/connect"
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
	var v connect.Instance
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
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "connect", regexache.MustCompile(`instance/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_resolve_best_voices_enabled", acctest.CtTrue), //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttr(resourceName, "contact_flow_logs_enabled", acctest.CtFalse),       //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttr(resourceName, "contact_lens_enabled", acctest.CtTrue),             //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "early_media_enabled", acctest.CtTrue), //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeConnectManaged),
					resource.TestCheckResourceAttr(resourceName, "inbound_calls_enabled", acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexache.MustCompile(rName)),
					resource.TestCheckResourceAttr(resourceName, "multi_party_conference_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outbound_calls_enabled", acctest.CtTrue),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrServiceRole, "iam", regexache.MustCompile(`role/aws-service-role/connect.amazonaws.com/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.InstanceStatusActive),
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
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "connect", regexache.MustCompile(`instance/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_resolve_best_voices_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "contact_flow_logs_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "contact_lens_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "early_media_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "inbound_calls_enabled", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexache.MustCompile(rName)),
					resource.TestCheckResourceAttr(resourceName, "multi_party_conference_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outbound_calls_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.InstanceStatusActive),
				),
			},
		},
	})
}

func testAccInstance_directory(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.Instance
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
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeExistingDirectory),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.InstanceStatusActive),
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
	var v connect.Instance
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
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeSaml),
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

func testAccCheckInstanceExists(ctx context.Context, n string, v *connect.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Connect Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

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

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

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
    Name = "terraform-testacc-directory-service-directory-tags"
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-test"
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
