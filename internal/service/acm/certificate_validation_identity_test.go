// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acm_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Identity tests are manual because the resource's `id` attribute is a timestamp,
// not the ARN, so the generated test's id-identity comparison does not apply.

func TestAccACMCertificateValidation_Identity_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_acm_certificate_validation.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domainName := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/CertificateValidation/basic/"),
				ConfigVariables: config.Variables{
					"domainName": config.StringVariable(domainName),
					"rootDomain": config.StringVariable(rootDomain),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateValidationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrCertificateARN: knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrCertificateARN)),
				},
			},
		},
	})
}

func TestAccACMCertificateValidation_Identity_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_acm_certificate_validation.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domainName := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup in alternate region
			{
				ConfigDirectory: config.StaticDirectory("testdata/CertificateValidation/region_override/"),
				ConfigVariables: config.Variables{
					"domainName": config.StringVariable(domainName),
					"rootDomain": config.StringVariable(rootDomain),
					"region":     config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrCertificateARN: knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrCertificateARN)),
				},
			},
		},
	})
}

func TestAccACMCertificateValidation_Identity_ExistingResource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_acm_certificate_validation.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domainName := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ACMServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/CertificateValidation/basic_v6.42.0/"),
				ConfigVariables: config.Variables{
					"domainName": config.StringVariable(domainName),
					"rootDomain": config.StringVariable(rootDomain),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateValidationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			// Step 2: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/CertificateValidation/basic/"),
				ConfigVariables: config.Variables{
					"domainName": config.StringVariable(domainName),
					"rootDomain": config.StringVariable(rootDomain),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateValidationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrCertificateARN: knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrCertificateARN)),
				},
			},
		},
	})
}

func TestAccACMCertificateValidation_Identity_ExistingResource_noRefreshNoChange(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_acm_certificate_validation.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domainName := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ACMServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ConfigDirectory: config.StaticDirectory("testdata/CertificateValidation/basic_v6.42.0/"),
				ConfigVariables: config.Variables{
					"domainName": config.StringVariable(domainName),
					"rootDomain": config.StringVariable(rootDomain),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateValidationExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			// Step 2: Current version, no refresh
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/CertificateValidation/basic/"),
				ConfigVariables: config.Variables{
					"domainName": config.StringVariable(domainName),
					"rootDomain": config.StringVariable(rootDomain),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}
