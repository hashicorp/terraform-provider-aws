// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCManagedResourceVisibility_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccVPCManagedResourceVisibility_basic,
		"update":        testAccVPCManagedResourceVisibility_update,
		"Identity":      testAccVPCManagedResourceVisibility_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccVPCManagedResourceVisibility_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_resource_visibility.test"
	visibility := string(awstypes.ManagedResourceDefaultVisibilityHidden)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckManagedResourceVisibility(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedResourceVisibilityConfig_basic(visibility),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("default_visibility"), knownvalue.StringExact(visibility)),
				},
			},
		},
	})
}

func testAccVPCManagedResourceVisibility_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_resource_visibility.test"
	visibility1 := string(awstypes.ManagedResourceDefaultVisibilityHidden)
	visibility2 := string(awstypes.ManagedResourceDefaultVisibilityVisible)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckManagedResourceVisibility(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedResourceVisibilityConfig_basic(visibility1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("default_visibility"), knownvalue.StringExact(visibility1)),
				},
			},
			{
				Config: testAccManagedResourceVisibilityConfig_basic(visibility2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("default_visibility"), knownvalue.StringExact(visibility2)),
				},
			},
		},
	})
}

func testAccPreCheckManagedResourceVisibility(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

	_, err := conn.GetManagedResourceVisibility(ctx, &ec2.GetManagedResourceVisibilityInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccManagedResourceVisibilityConfig_basic(visibility string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_resource_visibility" "test" {
  default_visibility = %[1]q
}
`, visibility)
}
