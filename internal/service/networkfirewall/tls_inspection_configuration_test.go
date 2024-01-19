// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/networkfirewall/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

// TIP: ==== UNIT TESTS ====
// This is an example of a unit test. Its name is not prefixed with
// "TestAcc" like an acceptance test.
//
// Unlike acceptance tests, unit tests do not access AWS and are focused on a
// function (or method). Because of this, they are quick and cheap to run.
//
// In designing a resource's implementation, isolate complex bits from AWS bits
// so that they can be tested through a unit test. We encourage more unit tests
// in the provider.
//
// Cut and dry functions using well-used patterns, like typical flatteners and
// expanders, don't need unit testing. However, if they are complex or
// intricate, they should be unit tested.
// func TestTLSInspectionConfigurationExampleUnitTest(t *testing.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		TestName string
// 		Input    string
// 		Expected string
// 		Error    bool
// 	}{
// 		{
// 			TestName: "empty",
// 			Input:    "",
// 			Expected: "",
// 			Error:    true,
// 		},
// 		{
// 			TestName: "descriptive name",
// 			Input:    "some input",
// 			Expected: "some output",
// 			Error:    false,
// 		},
// 		{
// 			TestName: "another descriptive name",
// 			Input:    "more input",
// 			Expected: "more output",
// 			Error:    false,
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase
// 		t.Run(testCase.TestName, func(t *testing.T) {
// 			t.Parallel()
// 			got, err := tfnetworkfirewall.FunctionFromResource(testCase.Input)

// 			if err != nil && !testCase.Error {
// 				t.Errorf("got error (%s), expected no error", err)
// 			}

// 			if err == nil && testCase.Error {
// 				t.Errorf("got (%s) and no error, expected error", got)
// 			}

// 			if got != testCase.Expected {
// 				t.Errorf("got %s, expected %s", got, testCase.Expected)
// 			}
// 		})
// 	}
// }

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.

// NOTE: acceptance tests require environment variable ACM_CERTIFICATE_ARN
// to be set and the ACM certificate to be validated during testing.

func TestAccNetworkFirewallTLSInspectionConfiguration_egressBasic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	ca := os.Getenv("ACM_CA_CERTIFICATE_ARN")
	if ca == "" {
		t.Skipf("Environment variable %s is not set, skipping test", ca)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"
	commonName := "example.com"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 4096)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, commonName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.NetworkFirewall)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_egressBasic(rName, certificate, key, caCertificate, ca),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encryption_configuration.*", map[string]string{
						"key_id": "AWS_OWNED_KMS_KEY",
						"type":   "AWS_OWNED_KMS_KEY",
					}),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.certificate_authority_arn", ca),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.check_certificate_revocation_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.check_certificate_revocation_status.0.revoked_status_action", "REJECT"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.check_certificate_revocation_status.0.unknown_status_action", "PASS"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.destinations.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.destination_ports.0.to_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.protocols.0", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.sources.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.source_ports.0.from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.source_ports.0.to_port", "65535"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "network-firewall", regexache.MustCompile(`tls-configuration/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccNetworkFirewallTLSInspectionConfiguration_ingressBasic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	certificateArn := os.Getenv("ACM_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skipf("Environment variable %s is not set, skipping test", certificateArn)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.NetworkFirewall)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_ingressBasic(rName, certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encryption_configuration.*", map[string]string{
						"key_id": "AWS_OWNED_KMS_KEY",
						"type":   "AWS_OWNED_KMS_KEY",
					}),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.server_certificates.0.resource_arn", certificateArn),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.destinations.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.destination_ports.0.to_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.protocols.0", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.sources.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.source_ports.0.from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.source_ports.0.to_port", "65535"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.check_certificate_revocation_status.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", "0"),
					// resource.TestCheckResourceAttr(resourceName, "certificates.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "network-firewall", regexache.MustCompile(`tls-configuration/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccNetworkFirewallTLSInspectionConfiguration_ingressWithEncryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	certificateArn := os.Getenv("ACM_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skipf("Environment variable %s is not set, skipping test", certificateArn)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"
	keyName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.NetworkFirewall)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_ingressWithEncryptionConfiguration(rName, certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "encryption_configuration.0.key_id", keyName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.server_certificates.0.resource_arn", certificateArn),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.destinations.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.destination_ports.0.to_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.protocols.0", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.sources.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.source_ports.0.from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.scopes.0.source_ports.0.to_port", "65535"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configurations.0.check_certificate_revocation_status.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", "0"),
					// resource.TestCheckResourceAttr(resourceName, "certificates.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "network-firewall", regexache.MustCompile(`tls-configuration/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccNetworkFirewallTLSInspectionConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	domainName := acctest.ACMCertificateDomainFromEnv(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NetworkFirewall)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_ingressBasic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceTLSInspectionConfiguration = newResourceTLSInspectionConfiguration
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkfirewall.ResourceTLSInspectionConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTLSInspectionConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_tls_inspection_configuration" {
				continue
			}

			input := &networkfirewall.DescribeTLSInspectionConfigurationInput{
				TLSInspectionConfigurationArn: aws.String(rs.Primary.Attributes["arn"]),
			}
			_, err := conn.DescribeTLSInspectionConfigurationWithContext(ctx, input)
			if errs.IsA[*networkfirewall.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.NetworkFirewall, create.ErrActionCheckingDestroyed, tfnetworkfirewall.ResNameTLSInspectionConfiguration, rs.Primary.ID, err)
			}

			return create.Error(names.NetworkFirewall, create.ErrActionCheckingDestroyed, tfnetworkfirewall.ResNameTLSInspectionConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTLSInspectionConfigurationExists(ctx context.Context, name string, tlsinspectionconfiguration *networkfirewall.DescribeTLSInspectionConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.NetworkFirewall, create.ErrActionCheckingExistence, tfnetworkfirewall.ResNameTLSInspectionConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.NetworkFirewall, create.ErrActionCheckingExistence, tfnetworkfirewall.ResNameTLSInspectionConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn(ctx)
		resp, err := conn.DescribeTLSInspectionConfigurationWithContext(ctx, &networkfirewall.DescribeTLSInspectionConfigurationInput{
			TLSInspectionConfigurationArn: aws.String(rs.Primary.Attributes["arn"]),
		})

		if err != nil {
			return create.Error(names.NetworkFirewall, create.ErrActionCheckingExistence, tfnetworkfirewall.ResNameTLSInspectionConfiguration, rs.Primary.ID, err)
		}

		*tlsinspectionconfiguration = *resp

		return nil
	}
}

// func testAccPreCheck(ctx context.Context, t *testing.T) {
// 	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn(ctx)

// 	input := &networkfirewall.ListTLSInspectionConfigurationsInput{}
// 	_, err := conn.ListTLSInspectionConfigurationsWithContext(ctx, input)

// 	if acctest.PreCheckSkipError(err) {
// 		t.Skipf("skipping acceptance testing: %s", err)
// 	}
// 	if err != nil {
// 		t.Fatalf("unexpected PreCheck error: %s", err)
// 	}
// }

// func testAccCheckTLSInspectionConfigurationNotRecreated(before, after *networkfirewall.DescribeTLSInspectionConfigurationOutput) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.TLSInspectionConfigurationResponse.TLSInspectionConfigurationId), aws.ToString(after.TLSInspectionConfigurationResponse.TLSInspectionConfigurationId); before != after {
// 			return create.Error(names.NetworkFirewall, create.ErrActionCheckingNotRecreated, tfnetworkfirewall.ResNameTLSInspectionConfiguration, aws.ToString(before.TLSInspectionConfigurationId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccTLSInspectionConfigurationConfig_ingressBasic(rName, certificateARN string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name = %[1]q
  description = "test"
  encryption_configuration {
    key_id = "AWS_OWNED_KMS_KEY"
    type = "AWS_OWNED_KMS_KEY"
  }
  tls_inspection_configuration {
    server_certificate_configurations {
      server_certificates {
        resource_arn = %[2]q
      }
      scopes {
        protocols = [ 6 ]
        destination_ports {
            from_port = 443
            to_port = 443
        }
        destinations {
          address_definition = "0.0.0.0/0"
        }  
        source_ports {
          from_port = 0
          to_port = 65535
        }
        sources {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }
}
`, rName, certificateARN)
}

func testAccTLSInspectionConfigurationConfig_encryptionConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}

func testAccTLSInspectionConfigurationConfig_ingressWithEncryptionConfiguration(rName, certificateARN string) string {
	return acctest.ConfigCompose(testAccTLSInspectionConfigurationConfig_encryptionConfiguration(rName), fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name = %[1]q
  description = "test"
  encryption_configuration {
    key_id = aws_kms_key.test.arn
    type = "CUSTOMER_KMS"
  }
  tls_inspection_configuration {
    server_certificate_configurations {
      server_certificates {
        resource_arn = %[2]q
      }
      scopes {
        protocols = [ 6 ]
        destination_ports {
          from_port = 443
          to_port = 443
        }
        destinations {
        address_definition = "0.0.0.0/0"
        }  
        source_ports {
        from_port = 0
        to_port = 65535
        }
        sources {
        address_definition = "0.0.0.0/0"
        }
      }
    }
  }
}
`, rName, certificateARN))
}

func testAccTLSInspectionConfigurationConfig_egressBasic(rName, certificate, privateKey, chain, ca string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body  = "%[2]s"
  private_key       = "%[3]s"
  certificate_chain = "%[4]s"
}

resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name = %[1]q
  description = "test"
  encryption_configuration {
    key_id = "AWS_OWNED_KMS_KEY"
    type = "AWS_OWNED_KMS_KEY"
  }
  tls_inspection_configuration {
    server_certificate_configurations {
      certificate_authority_arn = %[5]q
      check_certificate_revocation_status {
        revoked_status_action = "REJECT"
        unknown_status_action = "PASS"
      }
      scopes {
        protocols = [ 6 ]
        destination_ports {
          from_port = 443
          to_port = 443
        }
        destinations {
          address_definition = "0.0.0.0/0"
        }  
        source_ports {
          from_port = 0
          to_port = 65535
        }
        sources {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(privateKey), acctest.TLSPEMEscapeNewlines(chain), ca)
}
