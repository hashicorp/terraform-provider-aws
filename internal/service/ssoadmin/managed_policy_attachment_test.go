package ssoadmin_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
)

func TestAccSSOAdminManagedPolicyAttachment_basic(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(resourceName),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "managed_policy_arn", regexp.MustCompile(`policy/AlexaForBusinessDeviceSetup`)),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_name", "AlexaForBusinessDeviceSetup"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_set_arn", permissionSetResourceName, "arn"),
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

func TestAccSSOAdminManagedPolicyAttachment_forceNew(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(resourceName),
				),
			},
			{
				Config: testAccManagedPolicyAttachmentConfig_forceNew(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(resourceName),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "managed_policy_arn", regexp.MustCompile(`policy/AmazonCognitoReadOnly`)),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_name", "AmazonCognitoReadOnly"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_set_arn", permissionSetResourceName, "arn"),
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

func TestAccSSOAdminManagedPolicyAttachment_disappears(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssoadmin.ResourceManagedPolicyAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminManagedPolicyAttachment_Disappears_permissionSet(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssoadmin.ResourcePermissionSet(), permissionSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminManagedPolicyAttachment_multipleManagedPolicies(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	otherResourceName := "aws_ssoadmin_managed_policy_attachment.other"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPolicyAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(resourceName),
				),
			},
			{
				Config: testAccManagedPolicyAttachmentConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPolicyAttachmentExists(resourceName),
					testAccCheckManagedPolicyAttachmentExists(otherResourceName),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(otherResourceName, "managed_policy_arn", regexp.MustCompile(`policy/AmazonDynamoDBReadOnlyAccess`)),
					resource.TestCheckResourceAttr(otherResourceName, "managed_policy_name", "AmazonDynamoDBReadOnlyAccess"),
					resource.TestCheckResourceAttrPair(otherResourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(otherResourceName, "permission_set_arn", permissionSetResourceName, "arn"),
				),
			},
			{
				ResourceName:      otherResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckManagedPolicyAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssoadmin_managed_policy_attachment" {
			continue
		}

		managedPolicyArn, permissionSetArn, instanceArn, err := tfssoadmin.ParseManagedPolicyAttachmentID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing SSO Managed Policy Attachment ID (%s): %w", rs.Primary.ID, err)
		}

		policy, err := tfssoadmin.FindManagedPolicy(conn, managedPolicyArn, permissionSetArn, instanceArn)

		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if policy == nil {
			continue
		}

		return fmt.Errorf("Managed Policy (%s) for SSO Permission Set (%s) still exists", managedPolicyArn, permissionSetArn)

	}

	return nil
}

func testAccCheckManagedPolicyAttachmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		managedPolicyArn, permissionSetArn, instanceArn, err := tfssoadmin.ParseManagedPolicyAttachmentID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Managed Policy Attachment ID (%s): %w", rs.Primary.ID, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn

		policy, err := tfssoadmin.FindManagedPolicy(conn, managedPolicyArn, permissionSetArn, instanceArn)

		if err != nil {
			return err
		}

		if policy == nil {
			return fmt.Errorf("Managed Policy (%s) for SSO Permission Set (%s) not found", managedPolicyArn, permissionSetArn)
		}

		return nil
	}
}

func testAccManagedPolicyAttachmentBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccManagedPolicyAttachmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccManagedPolicyAttachmentBaseConfig(rName),
		`
resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AlexaForBusinessDeviceSetup"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`)
}

func testAccManagedPolicyAttachmentConfig_forceNew(rName string) string {
	return acctest.ConfigCompose(
		testAccManagedPolicyAttachmentBaseConfig(rName),
		`
resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonCognitoReadOnly"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`)
}

func testAccManagedPolicyAttachmentConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccManagedPolicyAttachmentConfig_basic(rName),
		`
resource "aws_ssoadmin_managed_policy_attachment" "other" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonDynamoDBReadOnlyAccess"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`)
}
