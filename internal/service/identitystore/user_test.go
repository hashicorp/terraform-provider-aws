// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfidentitystore "github.com/hashicorp/terraform-provider-aws/internal/service/identitystore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIdentityStoreUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "Acceptance Test"),
					resource.TestCheckResourceAttr(resourceName, "emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "external_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "identity_store_id"),
					resource.TestCheckResourceAttr(resourceName, "locale", ""),
					resource.TestCheckResourceAttr(resourceName, "name.0.family_name", "Doe"),
					resource.TestCheckResourceAttr(resourceName, "name.0.formatted", ""),
					resource.TestCheckResourceAttr(resourceName, "name.0.given_name", "John"),
					resource.TestCheckResourceAttr(resourceName, "name.0.honorific_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "name.0.honorific_suffix", ""),
					resource.TestCheckResourceAttr(resourceName, "name.0.middle_name", ""),
					resource.TestCheckResourceAttr(resourceName, "nickname", ""),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "preferred_language", ""),
					resource.TestCheckResourceAttr(resourceName, "profile_url", ""),
					resource.TestCheckResourceAttr(resourceName, "timezone", ""),
					resource.TestCheckResourceAttr(resourceName, "title", ""),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, rName),
					resource.TestCheckResourceAttr(resourceName, "user_type", ""),
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

func TestAccIdentityStoreUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfidentitystore.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIdentityStoreUser_Addresses(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_addresses2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.country", "US"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.formatted", "Formatted Address 1"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.locality", "The Locality 1"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.postal_code", "AAA BBB 1"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.primary", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.region", "The Region 1"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.street_address", "The Street Address 1"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.type", "The Type 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_addresses3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.country", "GB"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.formatted", "Formatted Address 2"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.locality", "The Locality 2"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.postal_code", "AAA BBB 2"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.primary", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.region", "The Region 2"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.street_address", "The Street Address 2"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.type", "The Type 2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_addresses1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.country", "US"),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.formatted", ""),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.locality", ""),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.postal_code", ""),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.primary", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.region", ""),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.street_address", ""),
					resource.TestCheckResourceAttr(resourceName, "addresses.0.type", "Home"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct0),
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

func TestAccIdentityStoreUser_Emails(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	email1 := acctest.RandomEmailAddress(acctest.RandomDomainName())
	email2 := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_emails1(rName, email1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "emails.0.primary", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "emails.0.type", "The Type 1"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.value", email1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_emails2(rName, email2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "emails.0.primary", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "emails.0.type", "The Type 2"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.value", email2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_emails3(rName, email2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "emails.0.primary", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "emails.0.type", ""),
					resource.TestCheckResourceAttr(resourceName, "emails.0.value", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", acctest.Ct0),
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

func TestAccIdentityStoreUser_Locale(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_locale(rName, "en-US"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "locale", "en-US"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_locale(rName, "en-GB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "locale", "en-GB"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "locale", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameFamilyName(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameFamilyName(rName, "Doe"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.family_name", "Doe"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_nameFamilyName(rName, "Deer"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.family_name", "Deer"),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameFormatted(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameFormatted(rName, "JD1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.formatted", "JD1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_nameFormatted(rName, "JD2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.formatted", "JD2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.formatted", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameGivenName(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameGivenName(rName, "John"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.given_name", "John"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_nameGivenName(rName, "Jane"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.given_name", "Jane"),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameHonorificPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameHonorificPrefix(rName, "Dr."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.honorific_prefix", "Dr."),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_nameHonorificPrefix(rName, "Mr."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.honorific_prefix", "Mr."),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameHonorificSuffix(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameHonorificSuffix(rName, "M.D."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.honorific_suffix", "M.D."),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_nameHonorificSuffix(rName, "MSc"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.honorific_suffix", "MSc"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.honorific_suffix", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameMiddleName(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameMiddleName(rName, "Howard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.middle_name", "Howard"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_nameMiddleName(rName, "Ben"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.middle_name", "Ben"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.middle_name", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NickName(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nickName(rName, "JD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "nickname", "JD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_nickName(rName, "Johnny"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "nickname", "Johnny"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "nickname", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_PhoneNumbers(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_phoneNumbers1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.0.primary", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.0.type", "The Type 1"),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.0.value", "111111"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_phoneNumbers2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.0.primary", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.0.type", "The Type 2"),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.0.value", "2222222"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_phoneNumbers3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.0.primary", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.0.type", ""),
					resource.TestCheckResourceAttr(resourceName, "phone_numbers.0.value", "2222222"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", acctest.Ct0),
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

func TestAccIdentityStoreUser_PreferredLanguage(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_preferredLanguage(rName, "EN"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "preferred_language", "EN"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_preferredLanguage(rName, "ET"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "preferred_language", "ET"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "preferred_language", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_ProfileURL(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_profileURL(rName, "http://example.com/1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "profile_url", "http://example.com/1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_profileURL(rName, "http://example.com/2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "profile_url", "http://example.com/2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "profile_url", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_Timezone(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_timezone(rName, "UTC"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_timezone(rName, "Europe/London"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "timezone", "Europe/London"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "timezone", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_Title(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_title(rName, "Mr"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "title", "Mr"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_title(rName, "Ms"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "title", "Ms"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "title", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_UserType(t *testing.T) {
	ctx := acctest.Context(t)
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IdentityStoreEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_userType(rName, "Member"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_type", "Member"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_userType(rName, "Admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_type", "Admin"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_type", ""),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_identitystore_user" {
				continue
			}

			_, err := tfidentitystore.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes["identity_store_id"], rs.Primary.Attributes["user_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IdentityStore User %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckUserExists(ctx context.Context, n string, v *identitystore.DescribeUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameUser, n, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameUser, n, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreClient(ctx)

		output, err := tfidentitystore.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes["identity_store_id"], rs.Primary.Attributes["user_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccUserConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}
`, rName)
}

func testAccUserConfig_addresses1(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }

  addresses {
    country = "US"
    type    = "Home"
  }
}
`, rName)
}

func testAccUserConfig_addresses2(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }

  addresses {
    country        = "US"
    formatted      = "Formatted Address 1"
    locality       = "The Locality 1"
    postal_code    = "AAA BBB 1"
    primary        = true
    region         = "The Region 1"
    street_address = "The Street Address 1"
    type           = "The Type 1"
  }
}
`, rName)
}

func testAccUserConfig_addresses3(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }

  addresses {
    country        = "GB"
    formatted      = "Formatted Address 2"
    locality       = "The Locality 2"
    postal_code    = "AAA BBB 2"
    primary        = false
    region         = "The Region 2"
    street_address = "The Street Address 2"
    type           = "The Type 2"
  }
}
`, rName)
}

func testAccUserConfig_emails1(rName, email string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "John"
    given_name  = "Doe"
  }

  emails {
    primary = true
    type    = "The Type 1"
    value   = %[2]q
  }
}
`, rName, email)
}

func testAccUserConfig_emails2(rName, email string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "John"
    given_name  = "Doe"
  }

  emails {
    primary = false
    type    = "The Type 2"
    value   = %[2]q
  }
}
`, rName, email)
}

func testAccUserConfig_emails3(rName, email string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "John"
    given_name  = "Doe"
  }

  emails {}
}
`, rName, email)
}

func testAccUserConfig_locale(rName, locale string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  locale = %[2]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}
`, rName, locale)
}

func testAccUserConfig_nameFamilyName(rName, familyName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = %[2]q
    given_name  = "John"
  }
}
`, rName, familyName)
}

func testAccUserConfig_nameFormatted(rName, formatted string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "Doe"
    formatted   = %[2]q
    given_name  = "John"
  }
}
`, rName, formatted)
}

func testAccUserConfig_nameGivenName(rName, givenName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "Doe"
    given_name  = %[2]q
  }
}
`, rName, givenName)
}

func testAccUserConfig_nameHonorificPrefix(rName, honorificPrefix string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name      = "Doe"
    given_name       = "John"
    honorific_prefix = %[2]q
  }
}
`, rName, honorificPrefix)
}

func testAccUserConfig_nameHonorificSuffix(rName, honorificSuffix string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name      = "Doe"
    given_name       = "John"
    honorific_suffix = %[2]q
  }
}
`, rName, honorificSuffix)
}

func testAccUserConfig_nameMiddleName(rName, middleName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "Doe"
    given_name  = "John"
    middle_name = %[2]q
  }
}
`, rName, middleName)
}

func testAccUserConfig_nickName(rName, nickName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  nickname = %[2]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}
`, rName, nickName)
}

func testAccUserConfig_phoneNumbers1(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "John"
    given_name  = "Doe"
  }

  phone_numbers {
    primary = true
    type    = "The Type 1"
    value   = "111111"
  }
}
`, rName)
}

func testAccUserConfig_phoneNumbers2(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "John"
    given_name  = "Doe"
  }

  phone_numbers {
    primary = false
    type    = "The Type 2"
    value   = "2222222"
  }
}
`, rName)
}

func testAccUserConfig_phoneNumbers3(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "John"
    given_name  = "Doe"
  }

  phone_numbers {
    value = "2222222"
  }
}
`, rName)
}

func testAccUserConfig_preferredLanguage(rName, preferredLanguage string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  preferred_language = %[2]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}
`, rName, preferredLanguage)
}

func testAccUserConfig_profileURL(rName, profileUrl string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  profile_url = %[2]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}
`, rName, profileUrl)
}

func testAccUserConfig_timezone(rName, timezone string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  timezone = %[2]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}
`, rName, timezone)
}

func testAccUserConfig_title(rName, title string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  title = %[2]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}
`, rName, title)
}

func testAccUserConfig_userType(rName, userType string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  user_type = %[2]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}
`, rName, userType)
}
