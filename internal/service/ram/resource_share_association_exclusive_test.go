// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	nftypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRAMResourceShareAssociationExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_resource_share_association_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRAMSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceShareAssociationExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAssociationExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAssociationExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_share_arn", "aws_ram_resource_share.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_arns.#", "1"),
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

func TestAccRAMResourceShareAssociationExclusive_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_resource_share_association_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRAMSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceShareAssociationExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAssociationExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAssociationExclusiveExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfram.ResourceResourceShareAssociationExclusive, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRAMResourceShareAssociationExclusive_exclusiveManagement(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_resource_share_association_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var resourceShareARN string
	var injectedFirewallPolicyARN string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRAMSharingWithOrganizationEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			if err := testAccCheckResourceShareAssociationExclusiveDestroy(ctx)(s); err != nil {
				return err
			}
			// Clean up the injected firewall policy
			if injectedFirewallPolicyARN != "" {
				nfConn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)
				_, _ = nfConn.DeleteFirewallPolicy(ctx, &networkfirewall.DeleteFirewallPolicyInput{
					FirewallPolicyArn: aws.String(injectedFirewallPolicyARN),
				})
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAssociationExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAssociationExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_arns.#", "1"),
					// Capture the resource share ARN for use in PreConfig
					testAccCaptureResourceShareARN(resourceName, &resourceShareARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Inject a Network Firewall policy outside of Terraform
				PreConfig: func() {
					ramConn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)
					nfConn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

					// Create a new firewall policy outside of Terraform
					policyName := rName + "-injected"
					createPolicyOutput, err := nfConn.CreateFirewallPolicy(ctx, &networkfirewall.CreateFirewallPolicyInput{
						FirewallPolicyName: aws.String(policyName),
						FirewallPolicy: &nftypes.FirewallPolicy{
							StatelessDefaultActions:         []string{"aws:drop"},
							StatelessFragmentDefaultActions: []string{"aws:drop"},
						},
					})
					if err != nil {
						t.Fatalf("creating injected firewall policy: %s", err)
					}
					injectedFirewallPolicyARN = aws.ToString(createPolicyOutput.FirewallPolicyResponse.FirewallPolicyArn)

					// Associate the injected firewall policy with the resource share
					_, err = ramConn.AssociateResourceShare(ctx, &ram.AssociateResourceShareInput{
						ResourceShareArn: aws.String(resourceShareARN),
						ResourceArns:     []string{injectedFirewallPolicyARN},
					})
					if err != nil {
						t.Fatalf("associating injected firewall policy with resource share: %s", err)
					}

					// Wait for resource association to be complete
					_, err = tfram.WaitResourceAssociationCreated(ctx, ramConn, resourceShareARN, injectedFirewallPolicyARN)
					if err != nil {
						t.Fatalf("waiting for injected resource association: %s", err)
					}
				},
				// Re-run the same config - the exclusive resource should remove the injected association
				Config: testAccResourceShareAssociationExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAssociationExclusiveExists(ctx, resourceName),
					// Verify the resource count is back to 1 (injected resource removed)
					resource.TestCheckResourceAttr(resourceName, "principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_arns.#", "1"),
					// Verify only 1 resource association exists in AWS
					testAccCheckResourceShareAssociationExclusiveResourceCount(ctx, &resourceShareARN, 1),
				),
			},
		},
	})
}

func TestAccRAMResourceShareAssociationExclusive_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_resource_share_association_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRAMSharingWithOrganizationEnabled(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourceShareAssociationExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAssociationExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAssociationExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_arns.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceShareAssociationExclusiveConfig_updateResources(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAssociationExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_arns.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceShareAssociationExclusiveConfig_updatePrincipals(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAssociationExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "principals.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "resource_arns.#", "2"),
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

func TestAccRAMResourceShareAssociationExclusive_servicePrincipalWithSources(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_resource_share_association_exclusive.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceShareAssociationExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAssociationExclusiveConfig_servicePrincipalWithSources(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareAssociationExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_share_arn", "aws_ram_resource_share.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sources.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "principals.*", "pca-connector-ad.amazonaws.com"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "resource_arns.*", "aws_acmpca_certificate_authority.test", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "sources.*", "data.aws_caller_identity.current", names.AttrAccountID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sources"},
			},
		},
	})
}

func testAccCheckResourceShareAssociationExclusiveExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)
		resourceShareARN := rs.Primary.Attributes["resource_share_arn"]

		// Check if resource share exists
		_, err := tfram.FindResourceShareOwnerSelfByARN(ctx, conn, resourceShareARN)
		if err != nil {
			return err
		}

		// Fetch all associations
		principals, resources, err := tfram.FindAssociationsForResourceShare(ctx, conn, resourceShareARN)
		if err != nil {
			return err
		}

		// Verify principals exist if configured
		if rs.Primary.Attributes["principals.#"] != "0" {
			if len(principals) == 0 {
				return fmt.Errorf("no principal associations found for resource share %s", resourceShareARN)
			}
		}

		// Verify resources exist if configured
		if rs.Primary.Attributes["resource_arns.#"] != "0" {
			if len(resources) == 0 {
				return fmt.Errorf("no resource associations found for resource share %s", resourceShareARN)
			}
		}

		return nil
	}
}

func testAccCheckResourceShareAssociationExclusiveDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ram_resource_share_association_exclusive" {
				continue
			}

			resourceShareARN := rs.Primary.Attributes["resource_share_arn"]

			// Fetch all associations
			principals, resources, err := tfram.FindAssociationsForResourceShare(ctx, conn, resourceShareARN)
			if err != nil && !retry.NotFound(err) {
				return err
			}

			// Check if principals still exist
			if rs.Primary.Attributes["principals.#"] != "0" {
				if len(principals) > 0 {
					return fmt.Errorf("RAM Resource Share Association Exclusive principals still exist for %s", resourceShareARN)
				}
			}

			// Check if resource associations still exist
			if rs.Primary.Attributes["resource_arns.#"] != "0" {
				if len(resources) > 0 {
					return fmt.Errorf("RAM Resource Share Association Exclusive resources still exist for %s", resourceShareARN)
				}
			}
		}

		return nil
	}
}

func testAccResourceShareAssociationExclusiveConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_ram_resource_share" "test" {
  allow_external_principals = false
  name                      = %[1]q
}

resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr        = "10.0.0.0/8"
    description = "Test entry"
  }
}

resource "aws_ram_resource_share_association_exclusive" "test" {
  resource_share_arn = aws_ram_resource_share.test.arn
  principals         = [data.aws_organizations_organization.test.arn]
  resource_arns      = [aws_ec2_managed_prefix_list.test.arn]
}
`, rName)
}

func testAccResourceShareAssociationExclusiveConfig_updateResources(rName string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_ram_resource_share" "test" {
  allow_external_principals = false
  name                      = %[1]q
}

resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr        = "10.0.0.0/8"
    description = "Test entry"
  }
}

resource "aws_ec2_managed_prefix_list" "test2" {
  name           = "%[1]s-2"
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr        = "172.16.0.0/12"
    description = "Test entry 2"
  }
}

resource "aws_ram_resource_share_association_exclusive" "test" {
  resource_share_arn = aws_ram_resource_share.test.arn
  principals         = [data.aws_organizations_organization.test.arn]
  resource_arns      = [aws_ec2_managed_prefix_list.test.arn, aws_ec2_managed_prefix_list.test2.arn]
}
`, rName)
}

func testAccResourceShareAssociationExclusiveConfig_updatePrincipals(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

data "aws_caller_identity" "receiver" {
  provider = "awsalternate"
}

resource "aws_ram_resource_share" "test" {
  allow_external_principals = true
  name                      = %[1]q
}

resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr        = "10.0.0.0/8"
    description = "Test entry"
  }
}

resource "aws_ec2_managed_prefix_list" "test2" {
  name           = "%[1]s-2"
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr        = "172.16.0.0/12"
    description = "Test entry 2"
  }
}

resource "aws_ram_resource_share_association_exclusive" "test" {
  resource_share_arn = aws_ram_resource_share.test.arn
  principals         = [data.aws_organizations_organization.test.arn, data.aws_caller_identity.receiver.account_id]
  resource_arns      = [aws_ec2_managed_prefix_list.test.arn, aws_ec2_managed_prefix_list.test2.arn]
}
`, rName))
}

// testAccCaptureResourceShareARN captures the resource share ARN from terraform state for use in PreConfig.
func testAccCaptureResourceShareARN(resourceName string, target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		*target = rs.Primary.Attributes["resource_share_arn"]
		return nil
	}
}

// testAccCheckResourceShareAssociationExclusiveResourceCount verifies the number of resource associations in AWS.
func testAccCheckResourceShareAssociationExclusiveResourceCount(ctx context.Context, resourceShareARN *string, expectedCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		_, resources, err := tfram.FindAssociationsForResourceShare(ctx, conn, *resourceShareARN)
		if err != nil {
			return err
		}

		if len(resources) != expectedCount {
			return fmt.Errorf("expected %d resource associations, got %d", expectedCount, len(resources))
		}

		return nil
	}
}

func testAccResourceShareAssociationExclusiveConfig_servicePrincipalWithSources(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

resource "aws_ram_resource_share" "test" {
  allow_external_principals = true
  name                      = %[1]q
}

resource "aws_ram_resource_share_association_exclusive" "test" {
  resource_share_arn = aws_ram_resource_share.test.arn
  principals         = ["pca-connector-ad.amazonaws.com"]
  resource_arns      = [aws_acmpca_certificate_authority.test.arn]
  sources            = [data.aws_caller_identity.current.account_id]

  depends_on = [aws_acmpca_certificate_authority_certificate.test]
}
`, rName)
}
