package ssoadmin_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
)

func TestAccSSOAdminPermissionSet_basic(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT1H"),
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

func TestAccSSOAdminPermissionSet_tags(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_tagsSingle(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPermissionSetConfig_tagsMultiple(rName, "key1", "updatedvalue1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updatedvalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPermissionSetConfig_tagsSingle(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccSSOAdminPermissionSet_updateDescription(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccPermissionSetConfig_updateDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
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

func TestAccSSOAdminPermissionSet_updateRelayState(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", ""),
				),
			},
			{
				Config: testAccPermissionSetConfig_updateRelayState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
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

func TestAccSSOAdminPermissionSet_updateSessionDuration(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
				),
			},
			{
				Config: testAccPermissionSetConfig_updateSessionDuration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT2H"),
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

// TestAccSSOAdminPermissionSet_RelayState_updateSessionDuration validates
// the resource's unchanged values (primarily relay_state) after updating the session_duration argument
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17411
func TestAccSSOAdminPermissionSet_RelayState_updateSessionDuration(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_relayState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT1H"),
				),
			},
			{
				Config: testAccPermissionSetConfig_relayStateUpdateSessionDuration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT2H"),
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

func TestAccSSOAdminPermissionSet_mixedPolicyAttachments(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
				),
			},
			{
				Config: testAccPermissionSetConfig_mixedPolicyAttachments(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(resourceName),
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

func testAccCheckPermissionSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssoadmin_permission_set" {
			continue
		}

		arn, instanceArn, err := tfssoadmin.ParseResourceID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Permission Set ID (%s): %w", rs.Primary.ID, err)
		}

		input := &ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(arn),
		}

		_, err = conn.DescribePermissionSet(input)

		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SSO Permission Set (%s) still exists", arn)
	}

	return nil
}

func testAccCheckSOAdminPermissionSetExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminConn

		arn, instanceArn, err := tfssoadmin.ParseResourceID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Permission Set ID (%s): %w", rs.Primary.ID, err)
		}

		_, err = conn.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(arn),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccPermissionSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccPermissionSetConfig_updateDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  description  = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccPermissionSetConfig_updateRelayState(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state  = "https://example.com"
}
`, rName)
}

func testAccPermissionSetConfig_updateSessionDuration(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  session_duration = "PT2H"
}
`, rName)
}

func testAccPermissionSetConfig_relayState(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  description      = %[1]q
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state      = "https://example.com"
  session_duration = "PT1H"
}
`, rName)
}

func testAccPermissionSetConfig_relayStateUpdateSessionDuration(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  description      = %[1]q
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state      = "https://example.com"
  session_duration = "PT2H"
}
`, rName)
}

func testAccPermissionSetConfig_tagsSingle(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPermissionSetConfig_tagsMultiple(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPermissionSetConfig_mixedPolicyAttachments(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AlexaForBusinessDeviceSetup"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }
}
resource "aws_ssoadmin_permission_set_inline_policy" "test" {
  inline_policy      = data.aws_iam_policy_document.test.json
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`, rName)
}
