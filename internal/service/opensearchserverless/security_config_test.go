// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessSecurityConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var securityconfig types.SecurityConfigDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckSecurityConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfig_basic(rName, "test-fixtures/idp-metadata.xml"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "saml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_options.session_timeout"),
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

func TestAccOpenSearchServerlessSecurityConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	var securityconfig types.SecurityConfigDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckSecurityConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfig_update(rName, "test-fixtures/idp-metadata.xml", names.AttrDescription, 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "saml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.session_timeout", "60"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				Config: testAccSecurityConfig_update(rName, "test-fixtures/idp-metadata.xml", "description updated", 40),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, resourceName, &securityconfig),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "saml"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.session_timeout", "40"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description updated"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var securityconfig types.SecurityConfigDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckSecurityConfig(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfig_basic(rName, "test-fixtures/idp-metadata.xml"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, resourceName, &securityconfig),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfopensearchserverless.ResourceSecurityConfig, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSecurityConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_security_config" {
				continue
			}

			_, err := tfopensearchserverless.FindSecurityConfigByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingDestroyed, tfopensearchserverless.ResNameSecurityConfig, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSecurityConfigExists(ctx context.Context, name string, securityconfig *types.SecurityConfigDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityConfig, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityConfig, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)
		resp, err := tfopensearchserverless.FindSecurityConfigByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityConfig, rs.Primary.ID, err)
		}

		*securityconfig = *resp

		return nil
	}
}

func testAccPreCheckSecurityConfig(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.ListSecurityConfigsInput{
		Type: types.SecurityConfigTypeSaml,
	}
	_, err := conn.ListSecurityConfigs(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSecurityConfig_basic(rName string, samlOptions string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_config" "test" {
  name = %[1]q
  type = "saml"
  saml_options {
    metadata = file("%[2]s")
  }
}
`, rName, samlOptions)
}

func testAccSecurityConfig_update(rName, samlOptions, description string, sessionTimeout int) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_config" "test" {
  name        = %[1]q
  description = %[3]q
  type        = "saml"

  saml_options {
    metadata        = file("%[2]s")
    session_timeout = %[4]d
  }
}
`, rName, samlOptions, description, sessionTimeout)
}
