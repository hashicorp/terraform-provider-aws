package iam_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/pquerna/otp/totp"
)

func TestAccIAMUser_basic(t *testing.T) {
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	path1 := "/"
	path2 := "/path2/"
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(name1, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("aws_iam_user.user", &conf),
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
					testAccCheckUserExists("aws_iam_user.user", &conf),
					testAccCheckUserAttributes(&conf, name2, "/path2/"),
				),
			},
		},
	})
}

func TestAccIAMUser_disappears(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserDisappears(&user),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMUser_ForceDestroy_accessKey(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserCreatesAccessKey(&user),
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
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserCreatesLoginProfile(&user),
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
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserCreatesMFADevice(&user),
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
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserUploadsSSHKey(&user),
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
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserServiceSpecificCredential(&user),
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
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserUploadSigningCertificate(&user),
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
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	path := "/"
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(name1, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("aws_iam_user.user", &conf),
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
					testAccCheckUserExists("aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccIAMUser_pathChange(t *testing.T) {
	var conf iam.GetUserOutput

	name := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	path1 := "/"
	path2 := "/updated/"
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(name, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("aws_iam_user.user", &conf),
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
					testAccCheckUserExists("aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccIAMUser_permissionsBoundary(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.user"

	permissionsBoundary1 := fmt.Sprintf("arn:%s:iam::aws:policy/AdministratorAccess", acctest.Partition())
	permissionsBoundary2 := fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", acctest.Partition())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckUserPermissionsBoundary(&user, ""),
				),
			},
			// Test addition
			{
				Config: testAccUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckUserPermissionsBoundary(&user, permissionsBoundary1),
				),
			},
			// Test empty value
			{
				Config: testAccUserConfig_permissionsBoundary(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckUserPermissionsBoundary(&user, ""),
				),
			},
		},
	})
}

func TestAccIAMUser_tags(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test-Name"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-tag2"),
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
				Config: testAccUserConfig_tagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-tagUpdate"),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_user" {
			continue
		}

		// Try to get user
		_, err := conn.GetUser(&iam.GetUserInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		// Verify the error is what we want
		if !tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return err
		}
	}

	return nil
}

func testAccCheckUserExists(n string, res *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		resp, err := conn.GetUser(&iam.GetUserInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckUserAttributes(user *iam.GetUserOutput, name string, path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *user.User.UserName != name {
			return fmt.Errorf("Bad name: %s", *user.User.UserName)
		}

		if *user.User.Path != path {
			return fmt.Errorf("Bad path: %s", *user.User.Path)
		}

		return nil
	}
}

func testAccCheckUserDisappears(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		userName := aws.StringValue(getUserOutput.User.UserName)

		_, err := conn.DeleteUser(&iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
		if err != nil {
			return fmt.Errorf("error deleting user %q: %s", userName, err)
		}

		return nil
	}
}

func testAccCheckUserPermissionsBoundary(getUserOutput *iam.GetUserOutput, expectedPermissionsBoundaryArn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualPermissionsBoundaryArn := ""

		if getUserOutput.User.PermissionsBoundary != nil {
			actualPermissionsBoundaryArn = *getUserOutput.User.PermissionsBoundary.PermissionsBoundaryArn
		}

		if actualPermissionsBoundaryArn != expectedPermissionsBoundaryArn {
			return fmt.Errorf("PermissionsBoundary: '%q', expected '%q'.", actualPermissionsBoundaryArn, expectedPermissionsBoundaryArn)
		}

		return nil
	}
}

func testAccCheckUserCreatesAccessKey(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		input := &iam.CreateAccessKeyInput{
			UserName: getUserOutput.User.UserName,
		}

		if _, err := conn.CreateAccessKey(input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Access Key: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserCreatesLoginProfile(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		password, err := tfiam.GeneratePassword(32)
		if err != nil {
			return err
		}
		input := &iam.CreateLoginProfileInput{
			Password: aws.String(password),
			UserName: getUserOutput.User.UserName,
		}

		if _, err := conn.CreateLoginProfile(input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Login Profile: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserCreatesMFADevice(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		createVirtualMFADeviceInput := &iam.CreateVirtualMFADeviceInput{
			Path:                 getUserOutput.User.Path,
			VirtualMFADeviceName: getUserOutput.User.UserName,
		}

		createVirtualMFADeviceOutput, err := conn.CreateVirtualMFADevice(createVirtualMFADeviceInput)
		if err != nil {
			return fmt.Errorf("error creating IAM User (%s) Virtual MFA Device: %s", aws.StringValue(getUserOutput.User.UserName), err)
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
			UserName:            getUserOutput.User.UserName,
		}

		if _, err := conn.EnableMFADevice(enableVirtualMFADeviceInput); err != nil {
			return fmt.Errorf("error enabling IAM User (%s) Virtual MFA Device: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

// Creates an IAM User SSH Key outside of Terraform to verify that it is deleted when `force_destroy` is set
func testAccCheckUserUploadsSSHKey(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		publicKey, _, err := RandSSHKeyPairSize(2048, acctest.DefaultEmailAddress)
		if err != nil {
			return fmt.Errorf("error generating random SSH key: %w", err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		input := &iam.UploadSSHPublicKeyInput{
			UserName:         getUserOutput.User.UserName,
			SSHPublicKeyBody: aws.String(publicKey),
		}

		_, err = conn.UploadSSHPublicKey(input)
		if err != nil {
			return fmt.Errorf("error uploading IAM User (%s) SSH key: %w", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

// Creates an IAM User Service Specific Credential outside of Terraform to verify that it is deleted when `force_destroy` is set
func testAccCheckUserServiceSpecificCredential(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		input := &iam.CreateServiceSpecificCredentialInput{
			UserName:    getUserOutput.User.UserName,
			ServiceName: aws.String("codecommit.amazonaws.com"),
		}

		_, err := conn.CreateServiceSpecificCredential(input)
		if err != nil {
			return fmt.Errorf("error uploading IAM User (%s) Service Specifc Credential: %w", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserUploadSigningCertificate(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		key := acctest.TLSRSAPrivateKeyPEM(2048)
		certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

		input := &iam.UploadSigningCertificateInput{
			CertificateBody: aws.String(certificate),
			UserName:        getUserOutput.User.UserName,
		}

		if _, err := conn.UploadSigningCertificate(input); err != nil {
			return fmt.Errorf("error uploading IAM User (%s) Signing Certificate : %s", aws.StringValue(getUserOutput.User.UserName), err)
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

func testAccUserConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %q

  tags = {
    Name = "test-Name"
    tag2 = "test-tag2"
  }
}
`, rName)
}

func testAccUserConfig_tagsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %q

  tags = {
    tag2 = "test-tagUpdate"
  }
}
`, rName)
}
