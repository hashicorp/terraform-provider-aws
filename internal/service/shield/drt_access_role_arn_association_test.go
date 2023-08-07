package shield_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/aws/aws-sdk-go/service/shield"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfshield "github.com/hashicorp/terraform-provider-aws/internal/service/shield"
)

// Acceptance test access AWS and cost money to run.
func TestAccShieldDRTAccessRoleArnAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var drtaccessrolearnassociation shield.DescribeDRTAccessOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_drt_access_role_arn_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRoleArn(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDRTAccessRoleArnAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDRTAccessRoleArnAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDRTAccessRoleArnAssociationExists(ctx, resourceName, &drtaccessrolearnassociation, t),
				),
			},
		},
	})
}

func TestAccShieldDRTAccessRoleArnAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var drtaccessrolearnassociation shield.DescribeDRTAccessOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_drt_access_role_arn_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRoleArn(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDRTAccessRoleArnAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDRTAccessRoleArnAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDRTAccessRoleArnAssociationExists(ctx, resourceName, &drtaccessrolearnassociation, t),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfshield.ResourceDRTAccessRoleArnAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDRTAccessRoleArnAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_shield_drt_access_role_arn_association" {
				continue
			}

			input := &shield.DescribeDRTAccessInput{}
			resp, err := conn.DescribeDRTAccessWithContext(ctx, input)

			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return nil
			}

			if resp != nil && (resp.RoleArn == nil || *resp.RoleArn == "") {
				return nil
			}

			return create.Error(names.Shield, create.ErrActionCheckingDestroyed, tfshield.ResNameDRTAccessRoleArnAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDRTAccessRoleArnAssociationNotExists(ctx context.Context, name string, drtaccessrolearnassociation *shield.DescribeDRTAccessOutput, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if ok {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameDRTAccessRoleArnAssociation, name, errors.New("found"))
		}
		if rs.Primary.ID != "" {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameDRTAccessRoleArnAssociation, name, errors.New("set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn(ctx)
		resp, err := conn.DescribeDRTAccessWithContext(ctx, &shield.DescribeDRTAccessInput{})
		if err != nil {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameDRTAccessRoleArnAssociation, rs.Primary.ID, err)
		}

		*drtaccessrolearnassociation = *resp

		return nil
	}
}

func testAccCheckDRTAccessRoleArnAssociationExists(ctx context.Context, name string, drtaccessrolearnassociation *shield.DescribeDRTAccessOutput, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameDRTAccessLogBucketAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameDRTAccessLogBucketAssociation, name, errors.New("not set"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn(ctx)
		resp, err := conn.DescribeDRTAccessWithContext(ctx, &shield.DescribeDRTAccessInput{})
		if err != nil {
			return create.Error(names.Shield, create.ErrActionCheckingExistence, tfshield.ResNameDRTAccessRoleArnAssociation, "testing", err)
		}

		*drtaccessrolearnassociation = *resp
		return nil
	}
}

func testAccPreCheckRoleArn(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldConn(ctx)

	input := &shield.DescribeDRTAccessInput{}
	_, err := conn.DescribeDRTAccessWithContext(ctx, input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDRTAccessRoleArnAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
	name = %[1]q
	assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        "Sid": "",
				"Effect": "Allow",
				"Principal": {
						"Service": "drt.shield.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
      },
    ]
  })
}

resource "aws_shield_protection_group" "test" {
  protection_group_id = %[1]q
  aggregation         = "MAX"
  pattern             = "ALL"
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSShieldDRTAccessPolicy"
}

resource "aws_shield_drt_access_role_arn_association" "test" {
  role_arn             = aws_iam_role.test.arn
}
`, rName)
}
