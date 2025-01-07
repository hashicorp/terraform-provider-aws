// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

type testAccGame struct {
	Location   *awstypes.S3Location
	LaunchPath string
}

func (gg *testAccGame) Parameters(portNumber int) string {
	return fmt.Sprintf("+sv_port %d +gamelift_start_server", portNumber)
}

// Location found from CloudTrail event after finishing tutorial
// e.g. https://us-west-2.console.aws.amazon.com/gamelift/home?region=us-west-2#/r/fleets/sample
func testAccSampleGame(region string) (*testAccGame, error) {
	version := "v1.2.0.0"
	accId, err := testAccAccountIdByRegion(region)
	if err != nil {
		return nil, err
	}
	bucket := fmt.Sprintf("gamelift-sample-builds-prod-%s", region)
	key := fmt.Sprintf("%s/server/sample_build_%s", version, version)
	roleArn := fmt.Sprintf("arn:%s:iam::%s:role/sample-build-upload-role-%s", acctest.Partition(), accId, region)
	launchPath := `C:\game\Bin64.Release.Dedicated\MultiplayerProjectLauncher_Server.exe`

	gg := &testAccGame{
		Location: &awstypes.S3Location{
			Bucket:  aws.String(bucket),
			Key:     aws.String(key),
			RoleArn: aws.String(roleArn),
		},
		LaunchPath: launchPath,
	}

	return gg, nil
}

// Account ID found from CloudTrail event (role ARN) after finishing tutorial in given region
func testAccAccountIdByRegion(region string) (string, error) {
	m := map[string]string{
		endpoints.ApNortheast1RegionID: "120069834884",
		endpoints.ApNortheast2RegionID: "805673136642",
		endpoints.ApSouth1RegionID:     "134975661615",
		endpoints.ApSoutheast1RegionID: "077577004113",
		endpoints.ApSoutheast2RegionID: "112188327105",
		endpoints.CaCentral1RegionID:   "800535022691",
		endpoints.EuCentral1RegionID:   "797584052317",
		endpoints.EuWest1RegionID:      "319803218673",
		endpoints.EuWest2RegionID:      "937342764187",
		endpoints.SaEast1RegionID:      "028872612690",
		endpoints.UsEast1RegionID:      "783764748367",
		endpoints.UsEast2RegionID:      "415729564621",
		endpoints.UsWest1RegionID:      "715879310420",
		endpoints.UsWest2RegionID:      "741061592171",
	}

	if accId, ok := m[region]; ok {
		return accId, nil
	}

	return "", &retry.NotFoundError{Message: fmt.Sprintf("GameLift Account ID not found for region %q", region)}
}
