// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/pquerna/otp/totp"
)

func TestAccIAMUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.User

	name1 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	path1 := "/"
	path2 := "/path2/"
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(name1, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, "aws_iam_user.user", &conf),
					testAccCheckUserAttributes(&conf, name1, "/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
			{
				Config: testAccUserConfig_basic(name2, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, "aws_iam_user.user", &conf),
					testAccCheckUserAttributes(&conf, name2, "/path2/"),
				),
			},
		},
	})
}

func TestAccIAMUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var user iam.User

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_accessKey(t *testing.T) {
	ctx := acctest.Context(t)
	var user iam.User

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					testAccCheckUserCreatesAccessKey(ctx, &user),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_loginProfile(t *testing.T) {
	ctx := acctest.Context(t)
	var user iam.User

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					testAccCheckUserCreatesLoginProfile(ctx, &user),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_mfaDevice(t *testing.T) {
	ctx := acctest.Context(t)
	var user iam.User

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					testAccCheckUserCreatesMFADevice(ctx, &user),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_sshKey(t *testing.T) {
	ctx := acctest.Context(t)
	var user iam.User

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					testAccCheckUserUploadsSSHKey(ctx, &user),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_serviceSpecificCred(t *testing.T) {
	ctx := acctest.Context(t)
	var user iam.User

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					testAccCheckUserServiceSpecificCredential(ctx, &user),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_signingCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var user iam.User

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					testAccCheckUserUploadSigningCertificate(ctx, t, &user),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccIAMUser_nameChange(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.User

	name1 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	path := "/"
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(name1, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, "aws_iam_user.user", &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
			{
				Config: testAccUserConfig_basic(name2, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, "aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccIAMUser_pathChange(t *testing.T) {
	ctx := acctest.Context(t)
	var conf iam.User

	name := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	path1 := "/"
	path2 := "/updated/"
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(name, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, "aws_iam_user.user", &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
			{
				Config: testAccUserConfig_basic(name, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, "aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccIAMUser_permissionsBoundary(t *testing.T) {
	ctx := acctest.Context(t)
	var user iam.User

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.user"

	permissionsBoundary1 := fmt.Sprintf("arn:%s:iam::aws:policy/AdministratorAccess", acctest.Partition())
	permissionsBoundary2 := fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", acctest.Partition())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckUserPermissionsBoundary(&user, permissionsBoundary1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
			// Test update
			{
				Config: testAccUserConfig_permissionsBoundary(rName, permissionsBoundary2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary2),
					testAccCheckUserPermissionsBoundary(&user, permissionsBoundary2),
				),
			},
			// Test import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			// Test removal
			{
				Config: testAccUserConfig_basic(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckUserPermissionsBoundary(&user, ""),
				),
			},
			// Test addition
			{
				Config: testAccUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckUserPermissionsBoundary(&user, permissionsBoundary1),
				),
			},
			// Test drift detection
			{
				PreConfig: func() {
					// delete the boundary manually
					conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)
					input := &iam.DeleteUserPermissionsBoundaryInput{
						UserName: user.UserName,
					}
					_, err := conn.DeleteUserPermissionsBoundaryWithContext(ctx, input)
					if err != nil {
						t.Fatalf("Failed to delete permission_boundary from user (%s): %s", aws.StringValue(user.UserName), err)
					}
				},
				Config: testAccUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				// check the boundary was restored
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckUserPermissionsBoundary(&user, permissionsBoundary1),
				),
			},
			// Test empty value
			{
				Config: testAccUserConfig_permissionsBoundary(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckUserPermissionsBoundary(&user, ""),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_user" {
				continue
			}

			_, err := tfiam.FindUserByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM User %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserExists(ctx context.Context, n string, v *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM User ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		output, err := tfiam.FindUserByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUserAttributes(user *iam.User, name string, path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *user.UserName != name {
			return fmt.Errorf("Bad name: %s", *user.UserName)
		}

		if *user.Path != path {
			return fmt.Errorf("Bad path: %s", *user.Path)
		}

		return nil
	}
}

func testAccCheckUserPermissionsBoundary(user *iam.User, expectedPermissionsBoundaryArn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualPermissionsBoundaryArn := ""

		if user.PermissionsBoundary != nil {
			actualPermissionsBoundaryArn = *user.PermissionsBoundary.PermissionsBoundaryArn
		}

		if actualPermissionsBoundaryArn != expectedPermissionsBoundaryArn {
			return fmt.Errorf("PermissionsBoundary: '%q', expected '%q'.", actualPermissionsBoundaryArn, expectedPermissionsBoundaryArn)
		}

		return nil
	}
}

func testAccCheckUserCreatesAccessKey(ctx context.Context, user *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		input := &iam.CreateAccessKeyInput{
			UserName: user.UserName,
		}

		if _, err := conn.CreateAccessKeyWithContext(ctx, input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Access Key: %s", aws.StringValue(user.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserCreatesLoginProfile(ctx context.Context, user *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)
		password, err := tfiam.GeneratePassword(32)
		if err != nil {
			return err
		}
		input := &iam.CreateLoginProfileInput{
			Password: aws.String(password),
			UserName: user.UserName,
		}

		if _, err := conn.CreateLoginProfileWithContext(ctx, input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Login Profile: %s", aws.StringValue(user.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserCreatesMFADevice(ctx context.Context, user *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		createVirtualMFADeviceInput := &iam.CreateVirtualMFADeviceInput{
			Path:                 user.Path,
			VirtualMFADeviceName: user.UserName,
		}

		createVirtualMFADeviceOutput, err := conn.CreateVirtualMFADeviceWithContext(ctx, createVirtualMFADeviceInput)
		if err != nil {
			return fmt.Errorf("error creating IAM User (%s) Virtual MFA Device: %s", aws.StringValue(user.UserName), err)
		}

		secret := string(createVirtualMFADeviceOutput.VirtualMFADevice.Base32StringSeed)
		authenticationCode1, err := totp.GenerateCode(secret, time.Now().Add(-30*time.Second))
		if err != nil {
			return fmt.Errorf("error generating Virtual MFA Device authentication code 1: %s", err)
		}
		authenticationCode2, err := totp.GenerateCode(secret, time.Now())
		if err != nil {
			return fmt.Errorf("error generating Virtual MFA Device authentication code 2: %s", err)
		}

		enableVirtualMFADeviceInput := &iam.EnableMFADeviceInput{
			AuthenticationCode1: aws.String(authenticationCode1),
			AuthenticationCode2: aws.String(authenticationCode2),
			SerialNumber:        createVirtualMFADeviceOutput.VirtualMFADevice.SerialNumber,
			UserName:            user.UserName,
		}

		if _, err := conn.EnableMFADeviceWithContext(ctx, enableVirtualMFADeviceInput); err != nil {
			return fmt.Errorf("error enabling IAM User (%s) Virtual MFA Device: %s", aws.StringValue(user.UserName), err)
		}

		return nil
	}
}

// Creates an IAM User SSH Key outside of Terraform to verify that it is deleted when `force_destroy` is set
func testAccCheckUserUploadsSSHKey(ctx context.Context, user *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		publicKey, _, err := RandSSHKeyPairSize(2048, acctest.DefaultEmailAddress)
		if err != nil {
			return fmt.Errorf("error generating random SSH key: %w", err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		input := &iam.UploadSSHPublicKeyInput{
			UserName:         user.UserName,
			SSHPublicKeyBody: aws.String(publicKey),
		}

		_, err = conn.UploadSSHPublicKeyWithContext(ctx, input)
		if err != nil {
			return fmt.Errorf("error uploading IAM User (%s) SSH key: %w", aws.StringValue(user.UserName), err)
		}

		return nil
	}
}

// Creates an IAM User Service Specific Credential outside of Terraform to verify that it is deleted when `force_destroy` is set
func testAccCheckUserServiceSpecificCredential(ctx context.Context, user *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		input := &iam.CreateServiceSpecificCredentialInput{
			UserName:    user.UserName,
			ServiceName: aws.String("codecommit.amazonaws.com"),
		}

		_, err := conn.CreateServiceSpecificCredentialWithContext(ctx, input)
		if err != nil {
			return fmt.Errorf("error uploading IAM User (%s) Service Specifc Credential: %w", aws.StringValue(user.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserUploadSigningCertificate(ctx context.Context, t *testing.T, user *iam.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn(ctx)

		key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
		certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

		input := &iam.UploadSigningCertificateInput{
			CertificateBody: aws.String(certificate),
			UserName:        user.UserName,
		}

		if _, err := conn.UploadSigningCertificateWithContext(ctx, input); err != nil {
			return fmt.Errorf("error uploading IAM User (%s) Signing Certificate : %s", aws.StringValue(user.UserName), err)
		}

		return nil
	}
}

func testAccUserConfig_basic(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name = %q
  path = %q
}
`, rName, path)
}

func testAccUserConfig_permissionsBoundary(rName, permissionsBoundary string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name                 = %q
  permissions_boundary = %q
}
`, rName, permissionsBoundary)
}

func testAccUserConfig_forceDestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  force_destroy = true
  name          = %q
}
`, rName)
}

func testAccUserConfig_tags0(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}
`, rName)
}

func testAccUserConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccUserConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccUserConfig_tagsNull(rName, tagKey1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q

  tags = {
    %[2]q = null
  }
}
`, rName, tagKey1)
}
