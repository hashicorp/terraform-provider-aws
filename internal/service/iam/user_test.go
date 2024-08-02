// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
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
	var conf awstypes.User

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
					names.AttrForceDestroy},
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
	var user awstypes.User

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
	var user awstypes.User

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
					names.AttrForceDestroy},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_loginProfile(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User

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
					names.AttrForceDestroy},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_mfaDevice(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User

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
					names.AttrForceDestroy},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_sshKey(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User

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
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_serviceSpecificCred(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User

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
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_signingCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User

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
					names.AttrForceDestroy},
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_policyAttached(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User

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
					testAccCheckUserAttachPolicy(ctx, &user), // externally attach a policy
				),
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_policyInline(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User

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
					testAccCheckUserInlinePolicy(ctx, &user), // externally put an inline policy
				),
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_policyInlineAttached(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User

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
					testAccCheckUserInlinePolicy(ctx, &user), // externally put an inline policy
					testAccCheckUserAttachPolicy(ctx, &user), // externally attach a policy
				),
			},
		},
	})
}

func TestAccIAMUser_nameChange(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.User

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
					names.AttrForceDestroy},
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
	var conf awstypes.User

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
					names.AttrForceDestroy},
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
	var user awstypes.User

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
					names.AttrForceDestroy},
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
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
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
					conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)
					input := &iam.DeleteUserPermissionsBoundaryInput{
						UserName: user.UserName,
					}
					_, err := conn.DeleteUserPermissionsBoundary(ctx, input)
					if err != nil {
						t.Fatalf("Failed to delete permission_boundary from user (%s): %s", aws.ToString(user.UserName), err)
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

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

func testAccCheckUserExists(ctx context.Context, n string, v *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM User ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		output, err := tfiam.FindUserByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUserAttributes(user *awstypes.User, name string, path string) resource.TestCheckFunc {
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

func testAccCheckUserPermissionsBoundary(user *awstypes.User, expectedPermissionsBoundaryArn string) resource.TestCheckFunc {
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

func testAccCheckUserCreatesAccessKey(ctx context.Context, user *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		input := &iam.CreateAccessKeyInput{
			UserName: user.UserName,
		}

		if _, err := conn.CreateAccessKey(ctx, input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Access Key: %s", aws.ToString(user.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserCreatesLoginProfile(ctx context.Context, user *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)
		password, err := tfiam.GeneratePassword(32)
		if err != nil {
			return err
		}
		input := &iam.CreateLoginProfileInput{
			Password: aws.String(password),
			UserName: user.UserName,
		}

		if _, err := conn.CreateLoginProfile(ctx, input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Login Profile: %s", aws.ToString(user.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserCreatesMFADevice(ctx context.Context, user *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		createVirtualMFADeviceInput := &iam.CreateVirtualMFADeviceInput{
			Path:                 user.Path,
			VirtualMFADeviceName: user.UserName,
		}

		createVirtualMFADeviceOutput, err := conn.CreateVirtualMFADevice(ctx, createVirtualMFADeviceInput)
		if err != nil {
			return fmt.Errorf("error creating IAM User (%s) Virtual MFA Device: %s", aws.ToString(user.UserName), err)
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

		if _, err := conn.EnableMFADevice(ctx, enableVirtualMFADeviceInput); err != nil {
			return fmt.Errorf("error enabling IAM User (%s) Virtual MFA Device: %s", aws.ToString(user.UserName), err)
		}

		return nil
	}
}

// Creates an IAM User SSH Key outside of Terraform to verify that it is deleted when `force_destroy` is set
func testAccCheckUserUploadsSSHKey(ctx context.Context, user *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		publicKey, _, err := RandSSHKeyPairSize(2048, acctest.DefaultEmailAddress)
		if err != nil {
			return fmt.Errorf("error generating random SSH key: %w", err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		input := &iam.UploadSSHPublicKeyInput{
			UserName:         user.UserName,
			SSHPublicKeyBody: aws.String(publicKey),
		}

		_, err = conn.UploadSSHPublicKey(ctx, input)
		if err != nil {
			return fmt.Errorf("error uploading IAM User (%s) SSH key: %w", aws.ToString(user.UserName), err)
		}

		return nil
	}
}

// Creates an IAM User Service Specific Credential outside of Terraform to verify that it is deleted when `force_destroy` is set
func testAccCheckUserServiceSpecificCredential(ctx context.Context, user *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		input := &iam.CreateServiceSpecificCredentialInput{
			UserName:    user.UserName,
			ServiceName: aws.String("codecommit.amazonaws.com"),
		}

		_, err := conn.CreateServiceSpecificCredential(ctx, input)
		if err != nil {
			return fmt.Errorf("error uploading IAM User (%s) Service Specifc Credential: %w", aws.ToString(user.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserUploadSigningCertificate(ctx context.Context, t *testing.T, user *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
		certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

		input := &iam.UploadSigningCertificateInput{
			CertificateBody: aws.String(certificate),
			UserName:        user.UserName,
		}

		if _, err := conn.UploadSigningCertificate(ctx, input); err != nil {
			return fmt.Errorf("error uploading IAM User (%s) Signing Certificate : %s", aws.ToString(user.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserAttachPolicy(ctx context.Context, user *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		doc := `{"Version":"2012-10-17","Statement":[{"Action":["iam:ChangePassword"],"Resource":"*","Effect":"Allow"}]}`

		input := &iam.CreatePolicyInput{
			PolicyDocument: aws.String(doc),
			PolicyName:     user.UserName,
		}

		output, err := conn.CreatePolicy(ctx, input)
		if err != nil {
			return fmt.Errorf("externally creating IAM Policy (%s): %s", aws.ToString(user.UserName), err)
		}

		_, err = tfresource.RetryWhenNewResourceNotFound(ctx, 2*time.Minute, func() (interface{}, error) {
			return tfiam.FindPolicyByARN(ctx, conn, aws.ToString(output.Policy.Arn))
		}, true)
		if err != nil {
			return fmt.Errorf("waiting for external creation of IAM Policy (%s): %s", aws.ToString(user.UserName), err)
		}

		if err := tfiam.AttachPolicyToUser(ctx, conn, aws.ToString(user.UserName), aws.ToString(output.Policy.Arn)); err != nil {
			return fmt.Errorf("externally attaching IAM User (%s) to policy (%s): %s", aws.ToString(user.UserName), aws.ToString(output.Policy.Arn), err)
		}

		return nil
	}
}

func testAccCheckUserInlinePolicy(ctx context.Context, user *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		doc := `{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":["ec2:DescribeElasticGpus","ec2:DescribeFastSnapshotRestores","ec2:DescribeScheduledInstances","ec2:DescribeScheduledInstanceAvailability"],"Resource":"*"}}`

		input := &iam.PutUserPolicyInput{
			PolicyDocument: aws.String(doc),
			PolicyName:     user.UserName,
			UserName:       user.UserName,
		}

		_, err := conn.PutUserPolicy(ctx, input)
		if err != nil {
			return fmt.Errorf("externally putting IAM User (%s) policy: %s", aws.ToString(user.UserName), err)
		}

		_, err = tfresource.RetryWhenNotFound(ctx, 2*time.Minute, func() (interface{}, error) {
			return tfiam.FindUserPolicyByTwoPartKey(ctx, conn, aws.ToString(user.UserName), aws.ToString(user.UserName))
		})
		if err != nil {
			return fmt.Errorf("waiting for external creation of inline IAM User Policy (%s): %s", aws.ToString(user.UserName), err)
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
