// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestDiffPortSettings(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Old           []any
		New           []any
		ExpectedAuths []awstypes.IpPermission
		ExpectedRevs  []awstypes.IpPermission
	}{
		{ // No change
			Old: []any{
				map[string]any{
					"from_port":        8443,
					"ip_range":         "192.168.0.0/24",
					names.AttrProtocol: "TCP",
					"to_port":          8443,
				},
			},
			New: []any{
				map[string]any{
					"from_port":        8443,
					"ip_range":         "192.168.0.0/24",
					names.AttrProtocol: "TCP",
					"to_port":          8443,
				},
			},
			ExpectedAuths: nil,
			ExpectedRevs:  nil,
		},
		{ // Addition
			Old: []any{
				map[string]any{
					"from_port":        8443,
					"ip_range":         "192.168.0.0/24",
					names.AttrProtocol: "TCP",
					"to_port":          8443,
				},
			},
			New: []any{
				map[string]any{
					"from_port":        8443,
					"ip_range":         "192.168.0.0/24",
					names.AttrProtocol: "TCP",
					"to_port":          8443,
				},
				map[string]any{
					"from_port":        8888,
					"ip_range":         "192.168.0.0/24",
					names.AttrProtocol: "TCP",
					"to_port":          8888,
				},
			},
			ExpectedAuths: []awstypes.IpPermission{
				{
					FromPort: aws.Int32(8888),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: awstypes.IpProtocolTcp,
					ToPort:   aws.Int32(8888),
				},
			},
			ExpectedRevs: nil,
		},
		{ // Removal
			Old: []any{
				map[string]any{
					"from_port":        8443,
					"ip_range":         "192.168.0.0/24",
					names.AttrProtocol: "TCP",
					"to_port":          8443,
				},
			},
			New:           []any{},
			ExpectedAuths: nil,
			ExpectedRevs: []awstypes.IpPermission{
				{
					FromPort: aws.Int32(8443),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: awstypes.IpProtocolTcp,
					ToPort:   aws.Int32(8443),
				},
			},
		},
		{ // Removal + Addition
			Old: []any{
				map[string]any{
					"from_port":        8443,
					"ip_range":         "192.168.0.0/24",
					names.AttrProtocol: "TCP",
					"to_port":          8443,
				},
			},
			New: []any{
				map[string]any{
					"from_port":        8443,
					"ip_range":         "192.168.0.0/24",
					names.AttrProtocol: "UDP",
					"to_port":          8443,
				},
			},
			ExpectedAuths: []awstypes.IpPermission{
				{
					FromPort: aws.Int32(8443),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: awstypes.IpProtocolUdp,
					ToPort:   aws.Int32(8443),
				},
			},
			ExpectedRevs: []awstypes.IpPermission{
				{
					FromPort: aws.Int32(8443),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: awstypes.IpProtocolTcp,
					ToPort:   aws.Int32(8443),
				},
			},
		},
	}

	ignoreExportedOpts := cmpopts.IgnoreUnexported(
		awstypes.IpPermission{},
	)

	for _, tc := range testCases {
		a, r := tfgamelift.DiffPortSettings(tc.Old, tc.New)

		if diff := cmp.Diff(a, tc.ExpectedAuths, ignoreExportedOpts); diff != "" {
			t.Errorf("unexpected ExpectedAuths diff (+wanted, -got): %s", diff)
		}

		if diff := cmp.Diff(r, tc.ExpectedRevs, ignoreExportedOpts); diff != "" {
			t.Errorf("unexpected ExpectedRevs diff (+wanted, -got): %s", diff)
		}
	}
}

func TestAccGameLiftFleet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.FleetAttributes

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)
	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "build_id", "aws_gamelift_build.test", names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`fleet/fleet-.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.0.certificate_type", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "default"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "NoProtection"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"runtime_configuration"},
			},
			{
				Config: testAccFleetConfig_basicUpdated(rNameUpdated, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "build_id", "aws_gamelift_build.test", names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`fleet/fleet-.+`)), resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "UpdatedGroup"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.new_game_sessions_per_creator", "2"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.policy_period_in_minutes", "15"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccGameLiftFleet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.FleetAttributes

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key
	launchPath := g.LaunchPath
	params := g.Parameters(33435)

	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basicTags1(rName, launchPath, params, bucketName, key, roleArn, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"runtime_configuration"},
			},
			{
				Config: testAccFleetConfig_basicTags2(rName, launchPath, params, bucketName, key, roleArn, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFleetConfig_basicTags1(rName, launchPath, params, bucketName, key, roleArn, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGameLiftFleet_allFields(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.FleetAttributes

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	desc := fmt.Sprintf("Terraform Acceptance Test %s", sdkacctest.RandString(8))

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := []string{
		g.Parameters(33435),
		g.Parameters(33436),
	}
	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_allFields(rName, desc, launchPath, params[0], bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "build_id", "aws_gamelift_build.test", names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`fleet/fleet-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "fleet_type", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, desc),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port":        "8080",
						"ip_range":         "8.8.8.8/32",
						names.AttrProtocol: "TCP",
						"to_port":          "8080",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port":        "8443",
						"ip_range":         "8.8.0.0/16",
						names.AttrProtocol: "TCP",
						"to_port":          "8443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port":        "60000",
						"ip_range":         "8.8.8.8/32",
						names.AttrProtocol: "UDP",
						"to_port":          "60000",
					}),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "TerraformAccTest"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2016"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.new_game_sessions_per_creator", "4"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.policy_period_in_minutes", "25"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.game_session_activation_timeout_seconds", "35"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.max_concurrent_game_session_activations", "99"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.parameters", params[0]),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"runtime_configuration"},
			},
			{
				Config: testAccFleetConfig_allFieldsUpdated(rNameUpdated, desc, launchPath, params[1], bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "build_id", "aws_gamelift_build.test", names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`fleet/fleet-.+`)), resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "fleet_type", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, desc),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port":        "8888",
						"ip_range":         "8.8.8.8/32",
						names.AttrProtocol: "TCP",
						"to_port":          "8888",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port":        "8443",
						"ip_range":         "8.4.0.0/16",
						names.AttrProtocol: "TCP",
						"to_port":          "8443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port":        "60000",
						"ip_range":         "8.8.8.8/32",
						names.AttrProtocol: "UDP",
						"to_port":          "60000",
					}),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "TerraformAccTest"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2016"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.new_game_sessions_per_creator", "4"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.policy_period_in_minutes", "25"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.game_session_activation_timeout_seconds", "35"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.max_concurrent_game_session_activations", "98"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.parameters", params[1]),
				),
			},
		},
	})
}

func TestAccGameLiftFleet_cert(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.FleetAttributes

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)
	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_cert(rName, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.0.certificate_type", "GENERATED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"runtime_configuration"},
			},
		},
	})
}

func TestAccGameLiftFleet_script(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.FleetAttributes

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_script(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "script_id", "aws_gamelift_script.test", names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`fleet/fleet-.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.0.certificate_type", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "default"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "NoProtection"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", "/local/game/lol"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"runtime_configuration"},
			},
		},
	})
}

func TestAccGameLiftFleet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf awstypes.FleetAttributes

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)
	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfgamelift.ResourceFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFleetExists(ctx context.Context, n string, v *awstypes.FleetAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftClient(ctx)

		output, err := tfgamelift.FindFleetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFleetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_gamelift_fleet" {
				continue
			}

			_, err := tfgamelift.FindFleetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("GameLift Fleet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccFleetConfig_basic(rName, launchPath, params, bucketName, key, roleArn string) string {
	return testAccFleetBasicTemplate(rName, bucketName, key, roleArn) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"
  name              = %[1]q

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = %[2]q
      parameters            = %[3]q
    }
  }
}
`, rName, launchPath, params)
}

func testAccFleetConfig_basicTags1(rName, launchPath, params, bucketName, key, roleArn, tagKey1, tagValue1 string) string {
	return testAccFleetBasicTemplate(rName, bucketName, key, roleArn) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"
  name              = %[1]q

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = %[2]q
      parameters            = %[3]q
    }
  }

  tags = {
    %[4]q = %[5]q
  }
}
`, rName, launchPath, params, tagKey1, tagValue1)
}

func testAccFleetConfig_basicTags2(rName, launchPath, params, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccFleetBasicTemplate(rName, bucketName, key, roleArn) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"
  name              = %[1]q

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = %[2]q
      parameters            = %[3]q
    }
  }

  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }
}
`, rName, launchPath, params, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccFleetConfig_basicUpdated(rName, launchPath, params, bucketName, key, roleArn string) string {
	return testAccFleetBasicTemplate(rName, bucketName, key, roleArn) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id                           = aws_gamelift_build.test.id
  ec2_instance_type                  = "c4.large"
  description                        = %[1]q
  name                               = %[1]q
  metric_groups                      = ["UpdatedGroup"]
  new_game_session_protection_policy = "FullProtection"

  resource_creation_limit_policy {
    new_game_sessions_per_creator = 2
    policy_period_in_minutes      = 15
  }

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = %[2]q
      parameters            = "%[3]s"
    }
  }
}
`, rName, launchPath, params)
}

func testAccFleetConfig_allFields(rName, desc, launchPath, params, bucketName, key, roleArn string) string {
	return testAccFleetBasicTemplate(rName, bucketName, key, roleArn) +
		testAccFleetIAMRole(rName) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"
  name              = "%s"
  description       = "%s"
  instance_role_arn = aws_iam_role.test.arn
  fleet_type        = "ON_DEMAND"

  ec2_inbound_permission {
    from_port = 8080
    ip_range  = "8.8.8.8/32"
    protocol  = "TCP"
    to_port   = 8080
  }

  ec2_inbound_permission {
    from_port = 8443
    ip_range  = "8.8.0.0/16"
    protocol  = "TCP"
    to_port   = 8443
  }

  ec2_inbound_permission {
    from_port = 60000
    ip_range  = "8.8.8.8/32"
    protocol  = "UDP"
    to_port   = 60000
  }

  metric_groups                      = ["TerraformAccTest"]
  new_game_session_protection_policy = "FullProtection"

  resource_creation_limit_policy {
    new_game_sessions_per_creator = 4
    policy_period_in_minutes      = 25
  }

  runtime_configuration {
    game_session_activation_timeout_seconds = 35
    max_concurrent_game_session_activations = 99

    server_process {
      concurrent_executions = 1
      launch_path           = %q
      parameters            = "%s"
    }
  }
}
`, rName, desc, launchPath, params)
}

func testAccFleetConfig_allFieldsUpdated(rName, desc, launchPath, params, bucketName, key, roleArn string) string {
	return testAccFleetBasicTemplate(rName, bucketName, key, roleArn) +
		testAccFleetIAMRole(rName) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"

  name              = "%s"
  description       = "%s"
  instance_role_arn = aws_iam_role.test.arn
  fleet_type        = "ON_DEMAND"

  ec2_inbound_permission {
    from_port = 8888
    ip_range  = "8.8.8.8/32"
    protocol  = "TCP"
    to_port   = 8888
  }

  ec2_inbound_permission {
    from_port = 8443
    ip_range  = "8.4.0.0/16"
    protocol  = "TCP"
    to_port   = 8443
  }

  ec2_inbound_permission {
    from_port = 60000
    ip_range  = "8.8.8.8/32"
    protocol  = "UDP"
    to_port   = 60000
  }

  metric_groups                      = ["TerraformAccTest"]
  new_game_session_protection_policy = "FullProtection"

  resource_creation_limit_policy {
    new_game_sessions_per_creator = 4
    policy_period_in_minutes      = 25
  }

  runtime_configuration {
    game_session_activation_timeout_seconds = 35
    max_concurrent_game_session_activations = 98

    server_process {
      concurrent_executions = 1
      launch_path           = %q
      parameters            = "%s"
    }
  }
}
`, rName, desc, launchPath, params)
}

func testAccFleetBasicTemplate(rName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2016"

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }
}
`, rName, bucketName, key, roleArn)
}

func testAccFleetIAMRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "gamelift.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  path        = "/"
  description = "GameLift Fleet PassRole Policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:PassRole",
        "sts:AssumeRole"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "test" {
  name       = %[1]q
  roles      = [aws_iam_role.test.name]
  policy_arn = aws_iam_policy.test.arn
}
`, rName)
}

func testAccFleetConfig_cert(rName, launchPath, params, bucketName, key, roleArn string) string {
	return testAccFleetBasicTemplate(rName, bucketName, key, roleArn) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"
  name              = %[1]q

  certificate_configuration {
    certificate_type = "GENERATED"
  }

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = %[2]q
      parameters            = %[3]q
    }
  }
}
`, rName, launchPath, params)
}

func testAccFleetConfig_script(rName string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name     = %[1]q
  zip_file = "test-fixtures/script.zip"
}

resource "aws_gamelift_fleet" "test" {
  script_id         = aws_gamelift_script.test.id
  ec2_instance_type = "t2.micro"
  name              = %[1]q

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = "/local/game/lol"
    }
  }
}
`, rName)
}
