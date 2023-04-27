package vpclattice_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeTargets_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var targets vpclattice.ListTargetsOutput
	resourceName := "aws_vpclattice_register_targets.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargets_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTargetsExists(ctx, resourceName, &targets),
					// acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile("targetgroup/.+$")),
					// resource.TestCheckResourceAttr(resourceName, "config.#", "1"),
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

func testAccTargets_basic() string {
	return `
resource "aws_vpclattice_register_targets" "test" {
  target_group_identifier = "tg-00153386728e69d10"

  targets {
    id   = "i-081f98c4ef2ff21e3"
    port = 80
  	}
}
`
}

func testAccCheckTargetsExists(ctx context.Context, name string, targets *vpclattice.ListTargetsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameRegisterTargets, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameRegisterTargets, name, errors.New("not set"))
		}
		targetGroupIdentifier := rs.Primary.Attributes["target_group_identifier"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()
		resp, err := conn.ListTargets(ctx, &vpclattice.ListTargetsInput{
			TargetGroupIdentifier: aws.String(targetGroupIdentifier),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameRegisterTargets, rs.Primary.ID, err)
		}

		*targets = *resp

		return nil
	}
}

func testAccCheckTargetsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_register_targets" {
				continue
			}

			targetGroupIdentifier := rs.Primary.Attributes["target_group_identifier"]

			_, err := conn.ListTargets(ctx, &vpclattice.ListTargetsInput{
				TargetGroupIdentifier: aws.String(targetGroupIdentifier),
				Targets:               []types.Target{},
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameRegisterTargets, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}
