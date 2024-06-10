// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfchimesdkvoice "github.com/hashicorp/terraform-provider-aws/internal/service/chimesdkvoice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccChimeSDKVoiceVoiceProfileDomain_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"VoiceProfileDomain": {
			acctest.CtBasic:      testAccVoiceProfileDomain_basic,
			acctest.CtDisappears: testAccVoiceProfileDomain_disappears,
			"update":             testAccVoiceProfileDomain_update,
			"tags":               testAccVoiceProfileDomain_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccVoiceProfileDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var voiceprofiledomain awstypes.VoiceProfileDomain
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_voice_profile_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChimeSDKVoiceEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceProfileDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceProfileDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceProfileDomainExists(ctx, resourceName, &voiceprofiledomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "server_side_encryption_configuration.0.kms_key_arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "chime", regexache.MustCompile(`voice-profile-domain/+.`)),
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

func testAccVoiceProfileDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var voiceprofiledomain awstypes.VoiceProfileDomain
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_voice_profile_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChimeSDKVoiceEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceProfileDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceProfileDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceProfileDomainExists(ctx, resourceName, &voiceprofiledomain),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchimesdkvoice.ResourceVoiceProfileDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVoiceProfileDomain_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.VoiceProfileDomain
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_chimesdkvoice_voice_profile_domain.test"
	description := "TF Acceptance test resource"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChimeSDKVoiceEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceProfileDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceProfileDomainConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceProfileDomainExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttrSet(resourceName, "server_side_encryption_configuration.0.kms_key_arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "chime", regexache.MustCompile(`voice-profile-domain/+.`)),
				),
			},
			{
				Config: testAccVoiceProfileDomainConfig_description(rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceProfileDomainExists(ctx, resourceName, &v2),
					testAccCheckVoiceProfileDomainNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrSet(resourceName, "server_side_encryption_configuration.0.kms_key_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "chime", regexache.MustCompile(`voice-profile-domain/+.`)),
				),
			},
		},
	})
}

func testAccVoiceProfileDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var voiceprofiledomain awstypes.VoiceProfileDomain
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_voice_profile_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ChimeSDKVoiceEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVoiceProfileDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVoiceProfileDomainConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceProfileDomainExists(ctx, resourceName, &voiceprofiledomain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVoiceProfileDomainConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceProfileDomainExists(ctx, resourceName, &voiceprofiledomain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVoiceProfileDomainConfig_tags1(rName, acctest.CtKey2, "value3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVoiceProfileDomainExists(ctx, resourceName, &voiceprofiledomain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, "value3"),
				),
			},
		},
	})
}

func testAccCheckVoiceProfileDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chimesdkvoice_voice_profile_domain" {
				continue
			}

			_, err := tfchimesdkvoice.FindVoiceProfileDomainByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if errs.IsA[*awstypes.NotFoundException](err) {
					return nil
				}
				return err
			}

			return create.Error(names.ChimeSDKVoice, create.ErrActionCheckingDestroyed, tfchimesdkvoice.ResNameVoiceProfileDomain, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckVoiceProfileDomainExists(ctx context.Context, name string, voiceprofiledomain *awstypes.VoiceProfileDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ChimeSDKVoice, create.ErrActionCheckingExistence, tfchimesdkvoice.ResNameVoiceProfileDomain, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ChimeSDKVoice, create.ErrActionCheckingExistence, tfchimesdkvoice.ResNameVoiceProfileDomain, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)
		resp, err := tfchimesdkvoice.FindVoiceProfileDomainByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ChimeSDKVoice, create.ErrActionCheckingExistence, tfchimesdkvoice.ResNameVoiceProfileDomain, rs.Primary.ID, err)
		}

		*voiceprofiledomain = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.ListVoiceProfileDomainsInput{}
	_, err := conn.ListVoiceProfileDomains(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckVoiceProfileDomainNotRecreated(before, after *awstypes.VoiceProfileDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if beforeID, afterID := aws.ToString(before.VoiceProfileDomainId), aws.ToString(after.VoiceProfileDomainId); beforeID != afterID {
			return create.Error(names.ChimeSDKVoice, create.ErrActionCheckingNotRecreated, tfchimesdkvoice.ResNameVoiceProfileDomain, beforeID, errors.New("recreated"))
		}

		return nil
	}
}

func testAccVoiceProfileDomainConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "TF Acceptance Test Voice Profile Domain"
  deletion_window_in_days = 7
}

resource "aws_chimesdkvoice_voice_profile_domain" "test" {
  name = %[1]q
  server_side_encryption_configuration {
    kms_key_arn = aws_kms_key.test.arn
  }
}
`, rName)
}

func testAccVoiceProfileDomainConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "TF Acceptance Test Voice Profile Domain"
  deletion_window_in_days = 7
}

resource "aws_chimesdkvoice_voice_profile_domain" "test" {
  name = %[1]q
  server_side_encryption_configuration {
    kms_key_arn = aws_kms_key.test.arn
  }
  description = %[2]q
}
`, rName, description)
}

func testAccVoiceProfileDomainConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "TF Acceptance Test Voice Profile Domain"
  deletion_window_in_days = 7
}

resource "aws_chimesdkvoice_voice_profile_domain" "test" {
  name = %[1]q
  server_side_encryption_configuration {
    kms_key_arn = aws_kms_key.test.arn
  }
  description = "TF Acceptance Test Resource"
  tags = {
    %[2]s = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVoiceProfileDomainConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "TF Acceptance Test Voice Profile Domain"
  deletion_window_in_days = 7
}

resource "aws_chimesdkvoice_voice_profile_domain" "test" {
  name = %[1]q
  server_side_encryption_configuration {
    kms_key_arn = aws_kms_key.test.arn
  }
  description = "TF Acceptance Test Resource"
  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
