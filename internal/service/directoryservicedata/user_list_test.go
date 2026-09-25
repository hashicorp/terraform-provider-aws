// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directoryservicedata_test

import (
	"fmt"
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

func TestAccDirectoryServiceDataUser_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_directoryservicedata_user.test[0]"
	resourceName2 := "aws_directoryservicedata_user.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	samAccountNamePrefix := fmt.Sprintf(
		"%s%s",
		testAccUserPrefix,
		acctest.RandStringFromCharSet(t, 20-len(testAccUserPrefix)-2, acctest.CharSetAlphaNum),
	)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectoryServiceDataServiceID),
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/User/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:        config.StringVariable(rName),
					"resource_count":       config.IntegerVariable(2),
					"directoryDomain":      config.StringVariable(domainName),
					"samAccountNamePrefix": config.StringVariable(samAccountNamePrefix),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("sam_account_name"), knownvalue.StringExact(samAccountNamePrefix+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("sam_account_name"), knownvalue.StringExact(samAccountNamePrefix+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/User/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:        config.StringVariable(rName),
					"resource_count":       config.IntegerVariable(2),
					"directoryDomain":      config.StringVariable(domainName),
					"samAccountNamePrefix": config.StringVariable(samAccountNamePrefix),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_directoryservicedata_user.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_directoryservicedata_user.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(samAccountNamePrefix+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_directoryservicedata_user.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_directoryservicedata_user.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_directoryservicedata_user.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(samAccountNamePrefix+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_directoryservicedata_user.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccDirectoryServiceDataUser_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_directoryservicedata_user.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	emailAddress := acctest.RandomEmailAddress(domainName)
	samAccountNamePrefix := fmt.Sprintf(
		"%s%s",
		testAccUserPrefix,
		acctest.RandStringFromCharSet(t, 20-len(testAccUserPrefix)-2, acctest.CharSetAlphaNum),
	)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectoryServiceDataServiceID),
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/User/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:        config.StringVariable(rName),
					"resource_count":       config.IntegerVariable(1),
					"directoryDomain":      config.StringVariable(domainName),
					"emailAddress":         config.StringVariable(emailAddress),
					"samAccountNamePrefix": config.StringVariable(samAccountNamePrefix),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("sam_account_name"), knownvalue.StringExact(samAccountNamePrefix+"-0")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/User/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:        config.StringVariable(rName),
					"resource_count":       config.IntegerVariable(1),
					"directoryDomain":      config.StringVariable(domainName),
					"emailAddress":         config.StringVariable(emailAddress),
					"samAccountNamePrefix": config.StringVariable(samAccountNamePrefix),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_directoryservicedata_user.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_directoryservicedata_user.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(samAccountNamePrefix+"-0")),
					querycheck.ExpectResourceKnownValues("aws_directoryservicedata_user.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("email_address"), knownvalue.StringExact(emailAddress)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEnabled), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("directory_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("distinguished_name"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("given_name"), knownvalue.StringExact(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("realm"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("sam_account_name"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("sid"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("user_principal_name"), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccDirectoryServiceDataUser_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_directoryservicedata_user.test[0]"
	resourceName2 := "aws_directoryservicedata_user.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	samAccountNamePrefix := fmt.Sprintf(
		"%s%s",
		testAccUserPrefix,
		acctest.RandStringFromCharSet(t, 20-len(testAccUserPrefix)-2, acctest.CharSetAlphaNum),
	)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckDirectoryService(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectoryServiceDataServiceID),
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/User/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:        config.StringVariable(rName),
					"resource_count":       config.IntegerVariable(2),
					"region":               config.StringVariable(acctest.AlternateRegion()),
					"directoryDomain":      config.StringVariable(domainName),
					"samAccountNamePrefix": config.StringVariable(samAccountNamePrefix),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("sam_account_name"), knownvalue.StringExact(samAccountNamePrefix+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("sam_account_name"), knownvalue.StringExact(samAccountNamePrefix+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/User/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:        config.StringVariable(rName),
					"resource_count":       config.IntegerVariable(2),
					"region":               config.StringVariable(acctest.AlternateRegion()),
					"directoryDomain":      config.StringVariable(domainName),
					"samAccountNamePrefix": config.StringVariable(samAccountNamePrefix),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_directoryservicedata_user.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_directoryservicedata_user.test", identity2.Checks()),
				},
			},
		},
	})
}
