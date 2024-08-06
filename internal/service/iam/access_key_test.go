// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/vault/helper/pgpkeys"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMAccessKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.AccessKeyMetadata
	resourceName := "aws_iam_access_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessKeyExists(ctx, resourceName, &conf),
					testAccCheckAccessKeyAttributes(&conf, "Active"),
					acctest.CheckResourceAttrRFC3339(resourceName, "create_date"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
					resource.TestCheckNoResourceAttr(resourceName, "encrypted_secret"),
					resource.TestCheckNoResourceAttr(resourceName, "key_fingerprint"),
					resource.TestCheckResourceAttrSet(resourceName, "ses_smtp_password_v4"),
					resource.TestCheckNoResourceAttr(resourceName, "encrypted_ses_smtp_password_v4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"encrypted_secret", "key_fingerprint", "pgp_key", "secret", "ses_smtp_password_v4", "encrypted_ses_smtp_password_v4"},
			},
		},
	})
}

func TestAccIAMAccessKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.AccessKeyMetadata
	resourceName := "aws_iam_access_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessKeyExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceAccessKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMAccessKey_encrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.AccessKeyMetadata
	resourceName := "aws_iam_access_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessKeyConfig_encrypted(rName, testPubKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessKeyExists(ctx, resourceName, &conf),
					testAccCheckAccessKeyAttributes(&conf, "Active"),
					testDecryptSecretKeyAndTest(resourceName, testPrivKey1),
					resource.TestCheckNoResourceAttr(resourceName, "secret"),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_secret"),
					resource.TestCheckResourceAttrSet(resourceName, "key_fingerprint"),
					resource.TestCheckNoResourceAttr(resourceName, "ses_smtp_password_v4"),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_ses_smtp_password_v4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"encrypted_secret", "key_fingerprint", "pgp_key", "secret", "ses_smtp_password_v4", "encrypted_ses_smtp_password_v4"},
			},
		},
	})
}

func TestAccIAMAccessKey_status(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.AccessKeyMetadata
	resourceName := "aws_iam_access_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessKeyConfig_status(rName, string(awstypes.StatusTypeInactive)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessKeyExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StatusTypeInactive)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"encrypted_secret", "key_fingerprint", "pgp_key", "secret", "ses_smtp_password_v4", "encrypted_ses_smtp_password_v4"},
			},
			{
				Config: testAccAccessKeyConfig_status(rName, string(awstypes.StatusTypeActive)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessKeyExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StatusTypeActive)),
				),
			},
			{
				Config: testAccAccessKeyConfig_status(rName, string(awstypes.StatusTypeInactive)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessKeyExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.StatusTypeInactive)),
				),
			},
		},
	})
}

func testAccCheckAccessKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_access_key" {
				continue
			}

			_, err := tfiam.FindAccessKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["user"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Access Key (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessKeyExists(ctx context.Context, n string, v *awstypes.AccessKeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		output, err := tfiam.FindAccessKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["user"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAccessKeyAttributes(accessKeyMetadata *awstypes.AccessKeyMetadata, status string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.Contains(*accessKeyMetadata.UserName, acctest.ResourcePrefix) {
			return fmt.Errorf("Bad username: %s", *accessKeyMetadata.UserName)
		}

		if string(accessKeyMetadata.Status) != status {
			return fmt.Errorf("Bad status: %s", string(accessKeyMetadata.Status))
		}

		return nil
	}
}

func testDecryptSecretKeyAndTest(nAccessKey, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keyResource, ok := s.RootModule().Resources[nAccessKey]
		if !ok {
			return fmt.Errorf("Not found: %s", nAccessKey)
		}

		secret, ok := keyResource.Primary.Attributes["encrypted_secret"]
		if !ok {
			return errors.New("No secret in state")
		}

		password, ok := keyResource.Primary.Attributes["encrypted_ses_smtp_password_v4"]
		if !ok {
			return errors.New("No password in state")
		}

		// We can't verify that the decrypted secret or password is correct, because we don't
		// have it. We can verify that decrypting it does not error
		_, err := pgpkeys.DecryptBytes(secret, key)
		if err != nil {
			return fmt.Errorf("Error decrypting secret: %s", err)
		}
		_, err = pgpkeys.DecryptBytes(password, key)
		if err != nil {
			return fmt.Errorf("Error decrypting password: %s", err)
		}

		return nil
	}
}

func testAccAccessKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_access_key" "test" {
  user = aws_iam_user.test.name
}
`, rName)
}

func testAccAccessKeyConfig_encrypted(rName, key string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_access_key" "test" {
  user = aws_iam_user.test.name

  pgp_key = <<EOF
%[2]s
EOF
}
`, rName, key)
}

func testAccAccessKeyConfig_status(rName string, status string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_access_key" "test" {
  user   = aws_iam_user.test.name
  status = %[2]q
}
`, rName, status)
}

func TestSESSMTPPasswordFromSecretKeySigV4(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Region   string
		Input    string
		Expected string
	}{
		{names.EUCentral1RegionID, "some+secret+key", "BMXhUYlu5Z3gSXVQORxlVa7XPaz91aGWdfHxvkOZdWZ2"},
		{names.EUCentral1RegionID, "another+secret+key", "BBbphbrQmrKMx42d1N6+C7VINYEBGI5v9VsZeTxwskfh"},
		{names.USWest1RegionID, "some+secret+key", "BH+jbMzper5WwlwUar9E1ySBqHa9whi0GPo+sJ0mVYJj"},
		{names.USWest1RegionID, "another+secret+key", "BKVmjjMDFk/qqw8EROW99bjCS65PF8WKvK5bSr4Y6EqF"},
	}

	for _, tc := range cases {
		actual, err := tfiam.SESSMTPPasswordFromSecretKeySigV4(&tc.Input, tc.Region)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if actual != tc.Expected {
			t.Fatalf("%q: expected %q, got %q", tc.Input, tc.Expected, actual)
		}
	}
}
