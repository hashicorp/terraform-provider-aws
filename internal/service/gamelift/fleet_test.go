package gamelift_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestDiffPortSettings(t *testing.T) {
	testCases := []struct {
		Old           []interface{}
		New           []interface{}
		ExpectedAuths []*gamelift.IpPermission
		ExpectedRevs  []*gamelift.IpPermission
	}{
		{ // No change
			Old: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			ExpectedAuths: []*gamelift.IpPermission{},
			ExpectedRevs:  []*gamelift.IpPermission{},
		},
		{ // Addition
			Old: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
				map[string]interface{}{
					"from_port": 8888,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8888,
				},
			},
			ExpectedAuths: []*gamelift.IpPermission{
				{
					FromPort: aws.Int64(8888),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: aws.String("TCP"),
					ToPort:   aws.Int64(8888),
				},
			},
			ExpectedRevs: []*gamelift.IpPermission{},
		},
		{ // Removal
			Old: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			New:           []interface{}{},
			ExpectedAuths: []*gamelift.IpPermission{},
			ExpectedRevs: []*gamelift.IpPermission{
				{
					FromPort: aws.Int64(8443),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: aws.String("TCP"),
					ToPort:   aws.Int64(8443),
				},
			},
		},
		{ // Removal + Addition
			Old: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "UDP",
					"to_port":   8443,
				},
			},
			ExpectedAuths: []*gamelift.IpPermission{
				{
					FromPort: aws.Int64(8443),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: aws.String("UDP"),
					ToPort:   aws.Int64(8443),
				},
			},
			ExpectedRevs: []*gamelift.IpPermission{
				{
					FromPort: aws.Int64(8443),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: aws.String("TCP"),
					ToPort:   aws.Int64(8443),
				},
			},
		},
	}

	for _, tc := range testCases {
		a, r := tfgamelift.DiffPortSettings(tc.Old, tc.New)

		authsString := fmt.Sprintf("%+v", a)
		expectedAuths := fmt.Sprintf("%+v", tc.ExpectedAuths)
		if authsString != expectedAuths {
			t.Fatalf("Expected authorizations: %+v\nGiven: %+v", tc.ExpectedAuths, a)
		}

		revString := fmt.Sprintf("%+v", r)
		expectedRevs := fmt.Sprintf("%+v", tc.ExpectedRevs)
		if revString != expectedRevs {
			t.Fatalf("Expected authorizations: %+v\nGiven: %+v", tc.ExpectedRevs, r)
		}
	}
}

func TestAccGameLiftFleet_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf gamelift.FleetAttributes

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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "build_id", "aws_gamelift_build.test", "id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`fleet/fleet-.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.0.certificate_type", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "default"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "NoProtection"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "build_id", "aws_gamelift_build.test", "id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`fleet/fleet-.+`)), resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", rNameUpdated),
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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccGameLiftFleet_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf gamelift.FleetAttributes

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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basicTags1(rName, launchPath, params, bucketName, key, roleArn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"runtime_configuration"},
			},
			{
				Config: testAccFleetConfig_basicTags2(rName, launchPath, params, bucketName, key, roleArn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFleetConfig_basicTags1(rName, launchPath, params, bucketName, key, roleArn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGameLiftFleet_allFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf gamelift.FleetAttributes

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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_allFields(rName, desc, launchPath, params[0], bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "build_id", "aws_gamelift_build.test", "id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`fleet/fleet-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "fleet_type", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port": "8080",
						"ip_range":  "8.8.8.8/32",
						"protocol":  "TCP",
						"to_port":   "8080",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port": "8443",
						"ip_range":  "8.8.0.0/16",
						"protocol":  "TCP",
						"to_port":   "8443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port": "60000",
						"ip_range":  "8.8.8.8/32",
						"protocol":  "UDP",
						"to_port":   "60000",
					}),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "TerraformAccTest"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
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
					testAccCheckFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "build_id", "aws_gamelift_build.test", "id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`fleet/fleet-.+`)), resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "fleet_type", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port": "8888",
						"ip_range":  "8.8.8.8/32",
						"protocol":  "TCP",
						"to_port":   "8888",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port": "8443",
						"ip_range":  "8.4.0.0/16",
						"protocol":  "TCP",
						"to_port":   "8443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_inbound_permission.*", map[string]string{
						"from_port": "60000",
						"ip_range":  "8.8.8.8/32",
						"protocol":  "UDP",
						"to_port":   "60000",
					}),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "TerraformAccTest"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf gamelift.FleetAttributes

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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_cert(rName, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &conf),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf gamelift.FleetAttributes

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_script(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "script_id", "aws_gamelift_script.test", "id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`fleet/fleet-.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_configuration.0.certificate_type", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "default"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "NoProtection"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", "/local/game/lol"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf gamelift.FleetAttributes

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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName, launchPath, params, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName, &conf),
					acctest.CheckResourceDisappears(acctest.Provider, tfgamelift.ResourceFleet(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfgamelift.ResourceFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFleetExists(n string, res *gamelift.FleetAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No GameLift Fleet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

		fleet, err := tfgamelift.FindFleetByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if aws.StringValue(fleet.FleetId) != rs.Primary.ID {
			return fmt.Errorf("GameLift Fleet not found")
		}

		*res = *fleet

		return nil
	}
}

func testAccCheckFleetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_fleet" {
			continue
		}

		_, err := tfgamelift.FindFleetByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		return nil
	}

	return nil
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
  operating_system = "WINDOWS_2012"

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
