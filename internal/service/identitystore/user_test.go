package identitystore_test

import (
	"context"
	"errors"
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfidentitystore "github.com/hashicorp/terraform-provider-aws/internal/service/identitystore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIdentityStoreUser_basic(t *testing.T) {
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "display_name", "Acceptance Test"),
					resource.TestCheckResourceAttr(resourceName, "emails.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "identity_store_id"),
					resource.TestCheckResourceAttr(resourceName, "name.0.family_name", "Doe"),
					resource.TestCheckResourceAttr(resourceName, "name.0.given_name", "John"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
					resource.TestCheckResourceAttr(resourceName, "user_name", rName),
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
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheckSSOAdminInstances(t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					acctest.CheckResourceDisappears(acctest.Provider, tfidentitystore.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIdentityStoreUser_Emails(t *testing.T) {
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	email1 := rName + "-1@example.com"
	email2 := rName + "-2@example.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_emails(rName, email1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.primary", "false"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.type", ""),
					resource.TestCheckResourceAttr(resourceName, "emails.0.value", email1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_emails(rName, email2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.primary", "false"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.type", ""),
					resource.TestCheckResourceAttr(resourceName, "emails.0.value", email2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_emailsTypeAndPrimary(rName, true, "test-type-1", email2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.primary", "true"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.type", "test-type-1"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.value", email2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_emailsTypeAndPrimary(rName, false, "test-type-2", email1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.primary", "false"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.type", "test-type-2"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.value", email1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserConfig_emails(rName, email2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.primary", "false"),
					resource.TestCheckResourceAttr(resourceName, "emails.0.type", ""),
					resource.TestCheckResourceAttr(resourceName, "emails.0.value", email2),
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
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "emails.#", "0"),
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

func TestAccIdentityStoreUser_NameFamilyName(t *testing.T) {
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameFamilyName(rName, "Doe"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.family_name", "Deer"),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameFormatted(t *testing.T) {
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameFormatted(rName, "JD1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.formatted", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameGivenName(t *testing.T) {
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameGivenName(rName, "John"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.given_name", "Jane"),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameHonorificPrefix(t *testing.T) {
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameHonorificPrefix(rName, "Dr."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.honorific_prefix", "Mr."),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameHonorificSuffix(t *testing.T) {
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameHonorificSuffix(rName, "M.D."),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.honorific_suffix", ""),
				),
			},
		},
	})
}

func TestAccIdentityStoreUser_NameMiddleName(t *testing.T) {
	var user identitystore.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_nameMiddleName(rName, "Howard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
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
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name.0.middle_name", ""),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreConn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_identitystore_user" {
			continue
		}

		_, err := conn.DescribeUser(ctx, &identitystore.DescribeUserInput{
			IdentityStoreId: aws.String(rs.Primary.Attributes["identity_store_id"]),
			UserId:          aws.String(rs.Primary.Attributes["user_id"]),
		})
		if err != nil {
			var nfe *types.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return create.Error(names.IdentityStore, create.ErrActionCheckingDestroyed, tfidentitystore.ResNameUser, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckUserExists(name string, user *identitystore.DescribeUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameUser, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameUser, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreConn
		ctx := context.Background()
		resp, err := conn.DescribeUser(ctx, &identitystore.DescribeUserInput{
			IdentityStoreId: aws.String(rs.Primary.Attributes["identity_store_id"]),
			UserId:          aws.String(rs.Primary.Attributes["user_id"]),
		})

		if err != nil {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameUser, rs.Primary.ID, err)
		}

		*user = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	//conn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreConn
	//ctx := context.Background()
	//
	//input := &identitystore.ListUsersInput{}
	//_, err := conn.ListUsers(ctx, input)
	//
	//if acctest.PreCheckSkipError(err) {
	//	t.Skipf("skipping acceptance testing: %s", err)
	//}
	//
	//if err != nil {
	//	t.Fatalf("unexpected PreCheck error: %s", err)
	//}
}

func testAccCheckUserNotRecreated(before, after *identitystore.DescribeUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.UserId), aws.ToString(after.UserId); before != after {
			return create.Error(names.IdentityStore, create.ErrActionCheckingNotRecreated, tfidentitystore.ResNameUser, before, errors.New("recreated"))
		}

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

func testAccUserConfig_emails(rName, emailValue string) string {
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
    value = %[2]q
  }
}
`, rName, emailValue)
}

func testAccUserConfig_emailsTypeAndPrimary(rName string, emailPrimary bool, emailType, emailValue string) string {
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
    primary = %[2]v
    type    = %[3]q
    value   = %[4]q
  }
}
`, rName, emailPrimary, emailType, emailValue)
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
