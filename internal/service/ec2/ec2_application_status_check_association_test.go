// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2ApplicationStatusCheckAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_application_status_check_association.test"
	checkResourceName := "aws_ec2_application_status_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationStatusCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationStatusCheckAssociationConfig_tag("Name", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_status_check_id", checkResourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttr(resourceName, "target_tag_key", "Name"),
					resource.TestCheckResourceAttr(resourceName, "target_tag_value", rName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccApplicationStatusCheckAssociationImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "application_status_check_id",
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheckAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_application_status_check_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationStatusCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationStatusCheckAssociationConfig_tag("Name", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceApplicationStatusCheckAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheckAssociation_disappears_ApplicationStatusCheck(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_application_status_check_association.test"
	checkResourceName := "aws_ec2_application_status_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationStatusCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationStatusCheckAssociationConfig_tag("Name", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceApplicationStatusCheck, checkResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(checkResourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(checkResourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheckAssociation_instance(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_application_status_check_association.test"
	checkResourceName := "aws_ec2_application_status_check.test"
	instanceResourceName := "aws_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationStatusCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationStatusCheckAssociationConfig_instance(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_status_check_id", checkResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, instanceResourceName, names.AttrID),
					resource.TestCheckNoResourceAttr(resourceName, "target_tag_key"),
					resource.TestCheckNoResourceAttr(resourceName, "target_tag_value"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccApplicationStatusCheckAssociationImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "application_status_check_id",
			},
		},
	})
}

func TestAccEC2ApplicationStatusCheckAssociation_replaceTarget(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ec2_application_status_check_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckApplicationStatusCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationStatusCheckAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationStatusCheckAssociationConfig_tag("Name", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_tag_key", "Name"),
					resource.TestCheckResourceAttr(resourceName, "target_tag_value", rName),
				),
			},
			{
				Config: testAccApplicationStatusCheckAssociationConfig_tag("Environment", "production"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationStatusCheckAssociationExists(ctx, t, resourceName),
					testAccCheckApplicationStatusCheckAssociationNotFound(ctx, t, resourceName, "", "Name", rName),
					resource.TestCheckResourceAttr(resourceName, "target_tag_key", "Environment"),
					resource.TestCheckResourceAttr(resourceName, "target_tag_value", "production"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccCheckApplicationStatusCheckAssociationNotFound(ctx context.Context, t *testing.T, name, instanceID, tagKey, tagValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Application Status Check Association", name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)
		checkID := rs.Primary.Attributes["application_status_check_id"]
		_, err := tfec2.FindApplicationStatusCheckAssociationByKey(ctx, conn, checkID, instanceID, tagKey, tagValue)
		if retry.NotFound(err) {
			return nil
		}
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Application Status Check Association", checkID, err)
		}

		return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "Application Status Check Association", checkID, errors.New("not destroyed"))
	}
}

func testAccCheckApplicationStatusCheckAssociationExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Application Status Check Association", name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)
		attributes := rs.Primary.Attributes

		_, err := tfec2.FindApplicationStatusCheckAssociationByKey(ctx, conn, attributes["application_status_check_id"], attributes[names.AttrInstanceID], attributes["target_tag_key"], attributes["target_tag_value"])
		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Application Status Check Association", attributes["application_status_check_id"], err)
		}

		return nil
	}
}

func testAccCheckApplicationStatusCheckAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_application_status_check_association" {
				continue
			}

			attributes := rs.Primary.Attributes
			_, err := tfec2.FindApplicationStatusCheckAssociationByKey(ctx, conn, attributes["application_status_check_id"], attributes[names.AttrInstanceID], attributes["target_tag_key"], attributes["target_tag_value"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "Application Status Check Association", attributes["application_status_check_id"], err)
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "Application Status Check Association", attributes["application_status_check_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccApplicationStatusCheckAssociationImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		attributes := rs.Primary.Attributes
		checkID := attributes["application_status_check_id"]
		if instanceID := attributes[names.AttrInstanceID]; instanceID != "" {
			return fmt.Sprintf("%s,instance-id,%s", checkID, instanceID), nil
		}

		return fmt.Sprintf("%s,tag,%s,%s", checkID, url.QueryEscape(attributes["target_tag_key"]), url.QueryEscape(attributes["target_tag_value"])), nil
	}
}

func testAccApplicationStatusCheckAssociationConfig_tag(tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_ec2_application_status_check" "test" {
  protocol = "http"
  port     = 80
}

resource "aws_ec2_application_status_check_association" "test" {
  application_status_check_id = aws_ec2_application_status_check.test.id
  target_tag_key              = %[1]q
  target_tag_value            = %[2]q
}
`, tagKey, tagValue)
}

func testAccApplicationStatusCheckAssociationConfig_instance(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigVPCWithSubnets(rName, 1),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_application_status_check" "test" {
  protocol = "http"
  port     = 80
}

resource "aws_ec2_application_status_check_association" "test" {
  application_status_check_id = aws_ec2_application_status_check.test.id
  instance_id                 = aws_instance.test.id
}
`, rName))
}
