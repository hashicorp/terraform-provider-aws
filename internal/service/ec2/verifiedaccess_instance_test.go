package ec2_test

import (
	"context"
	"errors"
	"fmt"
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

func TestAccVerifiedAccessInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var verifiedaccessinstance ec2.VerifiedAccessInstance
	description := sdkacctest.RandString(100)
	resourceName := "aws_verifiedaccess_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessInstanceConfig_basic(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceExists(ctx, resourceName, &verifiedaccessinstance),
					resource.TestCheckResourceAttr(resourceName, "description", description),
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

func TestAccVerifiedAccessInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var verifiedaccessinstance ec2.VerifiedAccessInstance
	description := sdkacctest.RandString(100)
	resourceName := "aws_verifiedaccess_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessInstanceConfig_tags1(description, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceExists(ctx, resourceName, &verifiedaccessinstance),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccVerifiedAccessInstanceConfig_tags2(description, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceExists(ctx, resourceName, &verifiedaccessinstance),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVerifiedAccessInstanceConfig_tags1(description, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceExists(ctx, resourceName, &verifiedaccessinstance),
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

func TestAccVerifiedAccessInstance_disappears(t *testing.T) {
	ctx := context.Background()
	var verifiedaccessinstance ec2.VerifiedAccessInstance
	description := sdkacctest.RandString(100)
	resourceName := "aws_verifiedaccess_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessInstanceConfig_basic(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceExists(ctx, resourceName, &verifiedaccessinstance),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVerifiedAccessInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVerifiedAccessInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_verifiedaccess_instance" {
			continue
		}

		_, err := tfec2.FindVerifiedAccessInstanceByID(ctx, conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVerifiedAccessInstance, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckVerifiedAccessInstanceExists(ctx context.Context, name string, verifiedaccessinstance *ec2.VerifiedAccessInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessInstance, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessInstance, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()

		output, err := tfec2.FindVerifiedAccessInstanceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, tfec2.ResNameVerifiedAccessInstance, rs.Primary.ID, err)
		}

		*verifiedaccessinstance = *output

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn()
	ctx := context.Background()

	input := &ec2.DescribeVerifiedAccessInstancesInput{}
	_, err := conn.DescribeVerifiedAccessInstancesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVerifiedAccessInstanceConfig_basic(description string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_instance" "test" {
  description = %[1]q
}
`, description)
}

func testAccVerifiedAccessInstanceConfig_tags1(description, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_instance" "test" {
  description = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, description, tagKey1, tagValue1)
}

func testAccVerifiedAccessInstanceConfig_tags2(description, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_instance" "test" {
  description = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, description, tagKey1, tagValue1, tagKey2, tagValue2)
}
