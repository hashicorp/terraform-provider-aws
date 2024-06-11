// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfchimesdkvoice "github.com/hashicorp/terraform-provider-aws/internal/service/chimesdkvoice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccChimeSDKVoiceGlobalSettings_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccGlobalSettings_basic,
		acctest.CtDisappears: testAccGlobalSettings_disappears,
		"update":             testAccGlobalSettings_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccGlobalSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_global_settings.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "voice_connector.0.cdr_bucket", bucketResourceName, names.AttrID),
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

func testAccGlobalSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_global_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchimesdkvoice.ResourceGlobalSettings(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGlobalSettings_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_global_settings.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID) // run test in us-east-1 only since eventual consistency causes intermittent failures in other regions.
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalSettingsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "voice_connector.0.cdr_bucket", bucketResourceName, names.AttrID),
				),
			},
			// Note: due to eventual consistency, the read after update can occasionally
			// return the previous cdr_bucket name, causing this test to fail. Running
			// in us-east-1 produces the most consistent results.
			{
				Config: testAccGlobalSettingsConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "voice_connector.0.cdr_bucket", bucketResourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccCheckGlobalSettingsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chimesdkvoice_global_settings" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)
			input := &chimesdkvoice.GetGlobalSettingsInput{}

			const retryTimeout = 10 * time.Second
			response := &chimesdkvoice.GetGlobalSettingsOutput{}

			err := tfresource.Retry(ctx, retryTimeout, func() *retry.RetryError {
				var err error
				response, err = conn.GetGlobalSettings(ctx, input)
				if err == nil && response.VoiceConnector.CdrBucket != nil {
					return retry.RetryableError(errors.New("error Chime Voice Connector Global settings still exists"))
				}
				return nil
			})

			if err != nil {
				return fmt.Errorf("error Chime Voice Connector Global settings still exists")
			}
		}
		return nil
	}
}

func testAccGlobalSettingsConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_chimesdkvoice_global_settings" "test" {
  voice_connector {
    cdr_bucket = aws_s3_bucket.test.id
  }
}
`, rName)
}
