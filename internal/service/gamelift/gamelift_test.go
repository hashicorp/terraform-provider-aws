// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
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
		names.APNortheast1RegionID: "120069834884",
		names.APNortheast2RegionID: "805673136642",
		names.APSouth1RegionID:     "134975661615",
		names.APSoutheast1RegionID: "077577004113",
		names.APSoutheast2RegionID: "112188327105",
		names.CACentral1RegionID:   "800535022691",
		names.EUCentral1RegionID:   "797584052317",
		names.EUWest1RegionID:      "319803218673",
		names.EUWest2RegionID:      "937342764187",
		names.SAEast1RegionID:      "028872612690",
		names.USEast1RegionID:      "783764748367",
		names.USEast2RegionID:      "415729564621",
		names.USWest1RegionID:      "715879310420",
		names.USWest2RegionID:      "741061592171",
	}

	if accId, ok := m[region]; ok {
		return accId, nil
	}

	return "", &retry.NotFoundError{Message: fmt.Sprintf("GameLift Account ID not found for region %q", region)}
}
