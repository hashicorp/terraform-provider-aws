// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acm_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMCertificate_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_acm_certificate.test[0]"
	resourceName2 := "aws_acm_certificate.test[1]"

	privateKeyPEM := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificatePEM := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM, acctest.RandomDomain(t).String())

	arn1 := tfstatecheck.StateValue()
	arn2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM),
					"resource_count":         config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					arn1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrARN)),
					arn2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrARN)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM),
					"resource_count":         config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_acm_certificate.test", map[string]knownvalue.Check{
						names.AttrARN: arn1.ValueCheck(),
					}),

					querycheck.ExpectIdentity("aws_acm_certificate.test", map[string]knownvalue.Check{
						names.AttrARN: arn2.ValueCheck(),
					}),
				},
			},
		},
	})
}

func TestAccACMCertificate_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_acm_certificate.test[0]"

	privateKeyPEM := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	domain := acctest.RandomDomain(t).String()
	certificatePEM := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM, domain)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM),
					"resource_count":         config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM),
					"resource_count":         config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_acm_certificate.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_acm_certificate.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(domain)),
					querycheck.ExpectResourceKnownValues("aws_acm_certificate.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("certificate_authority_arn"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("certificate_body"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCertificateChain), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDomainName), knownvalue.StringExact(domain)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("domain_validation_options"), knownvalue.SetSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("early_renewal_duration"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("key_algorithm"), knownvalue.StringExact("RSA_2048")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("not_after"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("not_before"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("options"), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPrivateKey), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("renewal_eligibility"), knownvalue.StringExact("INELIGIBLE")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("renewal_summary"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("ISSUED")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("subject_alternative_names"), knownvalue.SetSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrType), knownvalue.StringExact("IMPORTED")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("validation_emails"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("validation_method"), knownvalue.StringExact("NONE")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					}),
				},
			},
		},
	})
}

func TestAccACMCertificate_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_acm_certificate.test[0]"
	resourceName2 := "aws_acm_certificate.test[1]"

	privateKeyPEM := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificatePEM := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM, acctest.RandomDomain(t).String())

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM),
					"resource_count":         config.IntegerVariable(2),
					"region":                 config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM),
					"resource_count":         config.IntegerVariable(2),
					"region":                 config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_acm_certificate.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_acm_certificate.test", identity2.Checks()),
				},
			},
		},
	})
}
