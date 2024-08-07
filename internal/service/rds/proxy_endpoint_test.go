// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSProxyEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyEndpointPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "db_proxy_endpoint_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "db_proxy_name", "aws_db_proxy.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "target_role", "READ_WRITE"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(`db-proxy-endpoint:.+`)),
					resource.TestCheckResourceAttr(resourceName, "vpc_subnet_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.1", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
					resource.TestMatchResourceAttr(resourceName, names.AttrEndpoint, regexache.MustCompile(`^[\w\-\.]+\.rds\.amazonaws\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "is_default", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "tags.#", acctest.Ct0),
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

func TestAccRDSProxyEndpoint_targetRole(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyEndpointPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyEndpointConfig_targetRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyEndpointExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "target_role", "READ_ONLY"),
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

func TestAccRDSProxyEndpoint_vpcSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyEndpointPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyEndpointConfig_vpcSecurityGroupIDs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyEndpointExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test.0", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProxyEndpointConfig_vpcSecurityGroupIDs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyEndpointExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test.1", names.AttrID),
				),
			},
		},
	})
}

func TestAccRDSProxyEndpoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbProxy types.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyEndpointPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyEndpointConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyEndpointExists(ctx, resourceName, &dbProxy),
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
				Config: testAccProxyEndpointConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyEndpointExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProxyEndpointConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyEndpointExists(ctx, resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRDSProxyEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyEndpointPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyEndpointExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceProxyEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSProxyEndpoint_Disappears_proxy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v types.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccDBProxyEndpointPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyEndpointExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceProxy(), "aws_db_proxy.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDBProxyEndpointPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

	_, err := conn.DescribeDBProxyEndpoints(ctx, &rds.DescribeDBProxyEndpointsInput{})

	if tfawserr.ErrCodeEquals(err, tfrds.ErrCodeInvalidAction) {
		t.Skipf("skipping acceptance test, RDS Proxy Endpoint not supported: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckProxyEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_db_proxy_endpoint" {
				continue
			}

			_, err := tfrds.FindDBProxyEndpointByTwoPartKey(ctx, conn, rs.Primary.Attributes["db_proxy_name"], rs.Primary.Attributes["db_proxy_endpoint_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS DB Proxy Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProxyEndpointExists(ctx context.Context, n string, v *types.DBProxyEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBProxyEndpointByTwoPartKey(ctx, conn, rs.Primary.Attributes["db_proxy_name"], rs.Primary.Attributes["db_proxy_endpoint_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccProxyEndpointConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "{\"username\":\"db_user\",\"password\":\"db_user_password\"}"
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

data "aws_iam_policy_document" "assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["rds.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "secretsmanager:GetRandomPassword",
      "secretsmanager:CreateSecret",
      "secretsmanager:ListSecrets",
    ]
    resources = ["*"]
  }

  statement {
    actions   = ["secretsmanager:*"]
    resources = [aws_secretsmanager_secret.test.arn]
  }
}

resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  debug_logging          = false
  engine_family          = "MYSQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = aws_security_group.test[*].id
  vpc_subnet_ids         = aws_subnet.test[*].id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName))
}

func testAccProxyEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProxyEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test[*].id
}
`, rName))
}

func testAccProxyEndpointConfig_targetRole(rName string) string {
	return acctest.ConfigCompose(testAccProxyEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test[*].id
  target_role            = "READ_ONLY"
}
`, rName))
}

func testAccProxyEndpointConfig_vpcSecurityGroupIDs1(rName string) string {
	return acctest.ConfigCompose(testAccProxyEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = [aws_security_group.test[0].id]
}
`, rName))
}

func testAccProxyEndpointConfig_vpcSecurityGroupIDs2(rName string) string {
	return acctest.ConfigCompose(testAccProxyEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test[*].id
  vpc_security_group_ids = aws_security_group.test[*].id
}
`, rName))
}

func testAccProxyEndpointConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccProxyEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test[*].id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccProxyEndpointConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccProxyEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test[*].id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}
