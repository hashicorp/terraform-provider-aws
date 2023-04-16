package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedAccessGroup_basic(t *testing.T) {
	ctx := context.Background()
	var verifiedaccessgroup ec2.VerifiedAccessGroup
	description := sdkacctest.RandString(100)
	policyDocument := "permit(principal, action, resource) \nwhen {\n    context.http_request.method == \"GET\"\n};"
	resourceName := "aws_verifiedaccess_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_basic(description, policyDocument),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &verifiedaccessgroup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`verified-access-group/+.`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDocument),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccVerifiedAccessGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var verifiedaccessgroup ec2.VerifiedAccessGroup
	description := sdkacctest.RandString(100)
	policyDocument := "permit(principal, action, resource) \nwhen {\n    context.http_request.method == \"GET\"\n};"
	resourceName := "aws_verifiedaccess_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_tags1(description, policyDocument, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &verifiedaccessgroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccVerifiedAccessGroupConfig_tags2(description, policyDocument, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &verifiedaccessgroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVerifiedAccessGroupConfig_tags1(description, policyDocument, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &verifiedaccessgroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccVerifiedAccessGroup_disappears(t *testing.T) {
	ctx := context.Background()
	var verifiedaccessgroup ec2.VerifiedAccessGroup
	description := sdkacctest.RandString(100)
	policyDocument := "permit(principal, action, resource) \nwhen {\n    context.http_request.method == \"GET\"\n};"
	resourceName := "aws_verifiedaccess_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/sso.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_basic(description, policyDocument),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &verifiedaccessgroup),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedAccessGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVerifiedAccessGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_verifiedaccess_group" {
			continue
		}

		_, err := tfec2.FindVerifiedAccessGroupByID(ctx, conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVerifiedAccessGroup, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckVerifiedAccessGroupExists(ctx context.Context, name string, verifiedaccessgroup *ec2.VerifiedAccessGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessGroup, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessGroup, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

		resp, err := tfec2.FindVerifiedAccessGroupByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessGroup, rs.Primary.ID, err)
		}

		*verifiedaccessgroup = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()
	ctx := context.Background()

	input := &ec2.DescribeVerifiedAccessGroupsInput{}
	_, err := conn.DescribeVerifiedAccessGroupsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVerifiedAccessGroupConfig_baseConfig(description string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_instance" "test" {
  description = %[1]q
}

resource "aws_verifiedaccess_trust_provider" "test" {
  policy_reference_name    = "test"
  trust_provider_type      = "user"
  user_trust_provider_type = "iam-identity-center"
}

resource "aws_verifiedaccess_trust_provider_attachment" "test" {
  verified_access_instance_id       = aws_verifiedaccess_instance.test.id
  verified_access_trust_provider_id = aws_verifiedaccess_trust_provider.test.id
}
`, description)
}

func testAccVerifiedAccessGroupConfig_basic(description, policyDocument string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessGroupConfig_baseConfig(description), fmt.Sprintf(`
resource "aws_verifiedaccess_group" "test" {
  description                 = %[1]q
  verified_access_instance_id = aws_verifiedaccess_instance.test.id

  policy_document = %[2]q

  depends_on = [
    aws_verifiedaccess_trust_provider_attachment.test
  ]
}
`, description, policyDocument))
}

func testAccVerifiedAccessGroupConfig_tags1(description, policyDocument, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessGroupConfig_baseConfig(description), fmt.Sprintf(`
resource "aws_verifiedaccess_group" "test" {
  description                 = %[1]q
  verified_access_instance_id = aws_verifiedaccess_instance.test.id

  policy_document = %[2]q

  tags = {
    %[3]q = %[4]q
  }

  depends_on = [
    aws_verifiedaccess_trust_provider_attachment.test
  ]
}
`, description, policyDocument, tagKey1, tagValue1))
}

func testAccVerifiedAccessGroupConfig_tags2(description, policyDocument, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVerifiedAccessGroupConfig_baseConfig(description), fmt.Sprintf(`
resource "aws_verifiedaccess_group" "test" {
  description                 = %[1]q
  verified_access_instance_id = aws_verifiedaccess_instance.test.id

  policy_document = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
  depends_on = [
    aws_verifiedaccess_trust_provider_attachment.test
  ]
}
`, description, policyDocument, tagKey1, tagValue1, tagKey2, tagValue2))
}
