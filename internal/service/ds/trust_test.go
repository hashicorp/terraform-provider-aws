// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSTrust_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_basic(rName, domainName, domainNameOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^t-\w{10}`)),
					resource.TestCheckResourceAttr(resourceName, "conditional_forwarder_ip_addrs.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", "aws_directory_service_directory.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "remote_domain_name", domainNameOther),
					resource.TestCheckResourceAttr(resourceName, "selective_auth", string(awstypes.SelectiveAuthDisabled)),
					resource.TestCheckResourceAttr(resourceName, "trust_direction", string(awstypes.TrustDirectionTwoWay)),
					resource.TestCheckResourceAttr(resourceName, "trust_password", "Some0therPassword"),
					resource.TestCheckResourceAttr(resourceName, "trust_type", string(awstypes.TrustTypeForest)),
					resource.TestCheckResourceAttr(resourceName, "delete_associated_conditional_forwarder", acctest.CtFalse),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_date_time"),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_date_time"),
					resource.TestCheckResourceAttr(resourceName, "trust_state", string(awstypes.TrustStateVerifyFailed)),
					resource.TestCheckResourceAttrSet(resourceName, "trust_state_reason"),
					acctest.CheckResourceAttrRFC3339(resourceName, "state_last_updated_date_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
		},
	})
}

func TestAccDSTrust_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_basic(rName, domainName, domainNameOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfds.ResourceTrust, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDSTrust_Domain_TrailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_domain_trailingPeriod(rName, domainName, domainNameOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "remote_domain_name", fmt.Sprintf("%s.", domainNameOther)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
		},
	})
}

func TestAccDSTrust_twoWayBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	resourceOtherName := "aws_directory_service_trust.other"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_twoWayBasic(rName, domainName, domainNameOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "trust_state", string(awstypes.TrustStateVerified)),
					resource.TestCheckNoResourceAttr(resourceName, "trust_state_reason"),

					testAccCheckTrustExists(ctx, resourceOtherName, &v2),
					resource.TestCheckResourceAttr(resourceOtherName, "trust_state", string(awstypes.TrustStateVerified)),
					resource.TestCheckNoResourceAttr(resourceOtherName, "trust_state_reason"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
		},
	})
}

func TestAccDSTrust_oneWayBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	resourceOtherName := "aws_directory_service_trust.other"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_oneWay(rName, domainName, domainNameOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "trust_state", string(awstypes.TrustStateVerified)),
					resource.TestCheckNoResourceAttr(resourceName, "trust_state_reason"),

					testAccCheckTrustExists(ctx, resourceOtherName, &v2),
					resource.TestCheckResourceAttr(resourceOtherName, "trust_state", string(awstypes.TrustStateCreated)),
					resource.TestCheckNoResourceAttr(resourceOtherName, "trust_state_reason"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
		},
	})
}

func TestAccDSTrust_SelectiveAuth(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_SelectiveAuth(rName, domainName, domainNameOther, awstypes.SelectiveAuthEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "selective_auth", string(awstypes.SelectiveAuthEnabled)),
					resource.TestCheckResourceAttr(resourceName, "trust_state", string(awstypes.TrustStateVerifyFailed)), // Updating single-sided config
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
			{
				Config: testAccTrustConfig_SelectiveAuth(rName, domainName, domainNameOther, awstypes.SelectiveAuthDisabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "selective_auth", string(awstypes.SelectiveAuthDisabled)),
					resource.TestCheckResourceAttr(resourceName, "trust_state", string(awstypes.TrustStateVerifyFailed)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
		},
	})
}

func TestAccDSTrust_twoWaySelectiveAuth(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_twoWaySelectiveAuth(rName, domainName, domainNameOther, awstypes.SelectiveAuthEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "selective_auth", string(awstypes.SelectiveAuthEnabled)),
					resource.TestCheckResourceAttr(resourceName, "trust_state", string(awstypes.TrustStateVerified)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
			{
				Config: testAccTrustConfig_twoWaySelectiveAuth(rName, domainName, domainNameOther, awstypes.SelectiveAuthDisabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "selective_auth", string(awstypes.SelectiveAuthDisabled)),
					resource.TestCheckResourceAttr(resourceName, "trust_state", string(awstypes.TrustStateVerified)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
		},
	})
}

func TestAccDSTrust_TrustType(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_TrustType(rName, domainName, domainNameOther, awstypes.TrustTypeExternal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "trust_type", string(awstypes.TrustTypeExternal)),
					resource.TestCheckResourceAttr(resourceName, "trust_state", string(awstypes.TrustStateVerifyFailed)), // Updating single-sided config
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
		},
	})
}

func TestAccDSTrust_TrustTypeSpecifyDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_basic(rName, domainName, domainNameOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "trust_type", string(awstypes.TrustTypeForest)),
				),
			},
			{
				Config:   testAccTrustConfig_TrustType(rName, domainName, domainNameOther, awstypes.TrustTypeForest),
				PlanOnly: true,
			},
		},
	})
}

func TestAccDSTrust_ConditionalForwarderIPs(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_basic(rName, domainName, domainNameOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "conditional_forwarder_ip_addrs.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
			{
				Config: testAccTrustConfig_ConditionalForwarderIPs(rName, domainName, domainNameOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "conditional_forwarder_ip_addrs.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
		},
	})
}

func TestAccDSTrust_deleteAssociatedConditionalForwarder(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Trust
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_trust.test"
	domainName := acctest.RandomDomainName()
	domainNameOther := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustConfig_deleteAssociatedConditionalForwarder(rName, domainName, domainNameOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrustExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`^t-\w{10}`)),
					resource.TestCheckResourceAttr(resourceName, "conditional_forwarder_ip_addrs.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "delete_associated_conditional_forwarder", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccTrustStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"delete_associated_conditional_forwarder",
					"trust_password",
				},
			},
		},
	})
}

func testAccCheckTrustExists(ctx context.Context, n string, v *awstypes.Trust) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Directory Service Trust ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSClient(ctx)

		output, err := tfds.FindTrustByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTrustDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_trust" {
				continue
			}

			_, err := tfds.FindTrustByTwoPartKey(ctx, conn, rs.Primary.Attributes["directory_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Service Trust %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTrustStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["directory_id"], rs.Primary.Attributes["remote_domain_name"]), nil
	}
}

func testAccTrustConfig_basic(rName, domain, domainOther string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_trust" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = aws_directory_service_directory.other.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.other.dns_ip_addresses
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "other" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_directory_service_directory.test.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.other.security_group_id
}

resource "aws_security_group_rule" "other" {
  security_group_id = aws_directory_service_directory.other.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.test.security_group_id
}
`, domain, domainOther),
	)
}

func testAccTrustConfig_domain_trailingPeriod(rName, domain, domainOther string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_trust" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = "${aws_directory_service_directory.other.name}."
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.other.dns_ip_addresses
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "other" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_directory_service_directory.test.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.other.security_group_id
}

resource "aws_security_group_rule" "other" {
  security_group_id = aws_directory_service_directory.other.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.test.security_group_id
}
`, domain, domainOther),
	)
}

func testAccTrustConfig_twoWayBasic(rName, domain, domainOther string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_trust" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = aws_directory_service_directory.other.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.other.dns_ip_addresses
}

resource "aws_directory_service_trust" "other" {
  directory_id = aws_directory_service_directory.other.id

  remote_domain_name = aws_directory_service_directory.test.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.test.dns_ip_addresses
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "other" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_directory_service_directory.test.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.other.security_group_id
}

resource "aws_security_group_rule" "other" {
  security_group_id = aws_directory_service_directory.other.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.test.security_group_id
}
`, domain, domainOther),
	)
}

func testAccTrustConfig_oneWay(rName, domain, domainOther string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_trust" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = aws_directory_service_directory.other.name
  trust_direction    = "One-Way: Outgoing"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.other.dns_ip_addresses
}

resource "aws_directory_service_trust" "other" {
  directory_id = aws_directory_service_directory.other.id

  remote_domain_name = aws_directory_service_directory.test.name
  trust_direction    = "One-Way: Incoming"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.test.dns_ip_addresses
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "other" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_directory_service_directory.test.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.other.security_group_id
}

resource "aws_security_group_rule" "other" {
  security_group_id = aws_directory_service_directory.other.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.test.security_group_id
}
`, domain, domainOther),
	)
}

func testAccTrustConfig_SelectiveAuth(rName, domain, domainOther string, selectiveAuth awstypes.SelectiveAuth) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_trust" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = aws_directory_service_directory.other.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.other.dns_ip_addresses

  selective_auth = %[3]q
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "other" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_directory_service_directory.test.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.other.security_group_id
}

resource "aws_security_group_rule" "other" {
  security_group_id = aws_directory_service_directory.other.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.test.security_group_id
}
`, domain, domainOther, selectiveAuth),
	)
}

func testAccTrustConfig_twoWaySelectiveAuth(rName, domain, domainOther string, selectiveAuth awstypes.SelectiveAuth) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_trust" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = aws_directory_service_directory.other.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.other.dns_ip_addresses

  selective_auth = %[3]q
}

resource "aws_directory_service_trust" "other" {
  directory_id = aws_directory_service_directory.other.id

  remote_domain_name = aws_directory_service_directory.test.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.test.dns_ip_addresses
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "other" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_directory_service_directory.test.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.other.security_group_id
}

resource "aws_security_group_rule" "other" {
  security_group_id = aws_directory_service_directory.other.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.test.security_group_id
}
`, domain, domainOther, selectiveAuth),
	)
}

func testAccTrustConfig_TrustType(rName, domain, domainOther string, trustType awstypes.TrustType) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_trust" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = aws_directory_service_directory.other.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = aws_directory_service_directory.other.dns_ip_addresses

  trust_type = %[3]q
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "other" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_directory_service_directory.test.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.other.security_group_id
}

resource "aws_security_group_rule" "other" {
  security_group_id = aws_directory_service_directory.other.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.test.security_group_id
}
`, domain, domainOther, trustType),
	)
}

func testAccTrustConfig_ConditionalForwarderIPs(rName, domain, domainOther string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_trust" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = aws_directory_service_directory.other.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs = toset(slice(tolist(aws_directory_service_directory.other.dns_ip_addresses), 0, 1))
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "other" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_directory_service_directory.test.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.other.security_group_id
}

resource "aws_security_group_rule" "other" {
  security_group_id = aws_directory_service_directory.other.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.test.security_group_id
}
`, domain, domainOther),
	)
}

func testAccTrustConfig_deleteAssociatedConditionalForwarder(rName, domain, domainOther string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_trust" "test" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = aws_directory_service_directory.other.name
  trust_direction    = "Two-Way"
  trust_password     = "Some0therPassword"

  conditional_forwarder_ip_addrs          = aws_directory_service_directory.other.dns_ip_addresses
  delete_associated_conditional_forwarder = true
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "other" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_directory_service_directory.test.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.other.security_group_id
}

resource "aws_security_group_rule" "other" {
  security_group_id = aws_directory_service_directory.other.security_group_id

  type                     = "egress"
  protocol                 = "all"
  from_port                = 0
  to_port                  = 65535
  source_security_group_id = aws_directory_service_directory.test.security_group_id
}
`, domain, domainOther),
	)
}
