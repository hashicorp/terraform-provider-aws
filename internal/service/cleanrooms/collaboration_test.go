// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cleanrooms_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcleanrooms "github.com/hashicorp/terraform-provider-aws/internal/service/cleanrooms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCleanRoomsCollaboration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var collaboration cleanrooms.GetCollaborationOutput
	resourceName := "aws_cleanrooms_collaboration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollaborationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollaborationConfig_basic(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationExists(ctx, resourceName, &collaboration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, TEST_NAME),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, TEST_DESCRIPTION),
					resource.TestCheckResourceAttr(resourceName, "query_log_status", TEST_QUERY_LOG_STATUS),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "data_encryption_metadata.*", map[string]string{
						"allow_clear_text": acctest.CtTrue,
						"allow_duplicates": acctest.CtTrue,
						"allow_joins_on_columns_with_different_names": acctest.CtTrue,
						"preserve_nulls": acctest.CtFalse,
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cleanrooms", regexache.MustCompile(`collaboration:*`)),
					testCheckCreatorMember(ctx, resourceName),
					testAccCollaborationTags(ctx, resourceName, map[string]string{
						"Project": TEST_TAG,
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccCleanRoomsCollaboration_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var collaboration cleanrooms.GetCollaborationOutput
	resourceName := "aws_cleanrooms_collaboration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollaborationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollaborationConfig_basic(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationExists(ctx, resourceName, &collaboration),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcleanrooms.ResourceCollaboration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCleanRoomsCollaboration_mutableProperties(t *testing.T) {
	ctx := acctest.Context(t)

	var collaboration cleanrooms.GetCollaborationOutput
	updatedName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cleanrooms_collaboration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollaborationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollaborationConfig_basic(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationExists(ctx, resourceName, &collaboration),
				),
			},
			{
				Config: testAccCollaborationConfig_basic(updatedName, "updated Description", "Not Terraform"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationIsTheSame(resourceName, &collaboration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, updatedName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated Description"),
					testAccCollaborationTags(ctx, resourceName, map[string]string{
						"Project": "Not Terraform",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccCleanRoomsCollaboration_updateCreatorDisplayName(t *testing.T) {
	ctx := acctest.Context(t)

	var collaboration cleanrooms.GetCollaborationOutput
	resourceName := "aws_cleanrooms_collaboration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollaborationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollaborationConfig_basic(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationExists(ctx, resourceName, &collaboration),
				),
			},
			{
				Config: testAccCollaborationConfig_creatorDisplayName(TEST_NAME, TEST_DESCRIPTION, TEST_TAG, "updatedName"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationRecreated(resourceName, &collaboration),
					resource.TestCheckResourceAttr(resourceName, "creator_display_name", "updatedName"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}
func TestAccCleanRoomsCollaboration_updateQueryLogStatus(t *testing.T) {
	ctx := acctest.Context(t)

	var collaboration cleanrooms.GetCollaborationOutput
	resourceName := "aws_cleanrooms_collaboration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollaborationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollaborationConfig_basic(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationExists(ctx, resourceName, &collaboration),
				),
			},
			{
				Config: testAccCollaborationConfig_queryLogStatus(TEST_NAME, TEST_DESCRIPTION, TEST_TAG, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationRecreated(resourceName, &collaboration),
					resource.TestCheckResourceAttr(resourceName, "query_log_status", "ENABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}
func TestAccCleanRoomsCollaboration_dataEncryptionSettings(t *testing.T) {
	ctx := acctest.Context(t)

	var collaboration cleanrooms.GetCollaborationOutput
	resourceName := "aws_cleanrooms_collaboration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollaborationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollaborationConfig_basic(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationExists(ctx, resourceName, &collaboration),
				),
			},
			{
				Config: testAccCollaborationConfig_updatedDataEncryptionSettings(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationRecreated(resourceName, &collaboration),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "data_encryption_metadata.*", map[string]string{
						"allow_clear_text": acctest.CtTrue,
						"allow_duplicates": acctest.CtTrue,
						"allow_joins_on_columns_with_different_names": acctest.CtTrue,
						"preserve_nulls": acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccCollaborationConfig_noDataEncryptionSettings(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationRecreated(resourceName, &collaboration),
					resource.TestCheckResourceAttr(resourceName, "data_encryption_metadata.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccCleanRoomsCollaboration_updateMemberAbilities(t *testing.T) {
	ctx := acctest.Context(t)

	var collaboration cleanrooms.GetCollaborationOutput
	resourceName := "aws_cleanrooms_collaboration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CleanRoomsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollaborationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollaborationConfig_additionalMember(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationExists(ctx, resourceName, &collaboration),
					resource.TestCheckResourceAttr(resourceName, "member.0.account_id", "123456789012"),
					resource.TestCheckResourceAttr(resourceName, "member.0.display_name", "OtherMember"),
					resource.TestCheckResourceAttr(resourceName, "member.0.status", "INVITED"),
					resource.TestCheckResourceAttr(resourceName, "member.0.member_abilities.#", acctest.Ct0),
				),
			},
			{
				Config: testAccCollaborationConfig_swapMemberAbilities(TEST_NAME, TEST_DESCRIPTION, TEST_TAG),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollaborationRecreated(resourceName, &collaboration),
					resource.TestCheckResourceAttr(resourceName, "creator_member_abilities.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "member.0.member_abilities.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func testAccCheckCollaborationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cleanrooms_collaboration" {
				continue
			}

			_, err := conn.GetCollaboration(ctx, &cleanrooms.GetCollaborationInput{
				CollaborationIdentifier: aws.String(rs.Primary.ID),
			})
			if err != nil {
				// We throw access denied exceptions for Not Found Collaboration since they are cross account resources
				var nfe *types.AccessDeniedException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.CleanRooms, create.ErrActionCheckingDestroyed, tfcleanrooms.ResNameCollaboration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCollaborationExists(ctx context.Context, name string, collaboration *cleanrooms.GetCollaborationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameCollaboration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameCollaboration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)
		resp, err := conn.GetCollaboration(ctx, &cleanrooms.GetCollaborationInput{
			CollaborationIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameCollaboration, rs.Primary.ID, err)
		}

		*collaboration = *resp

		return nil
	}
}

func testAccCheckCollaborationIsTheSame(name string, collaboration *cleanrooms.GetCollaborationOutput) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		return checkCollaborationIsSame(name, collaboration, state)
	}
}

func testAccCheckCollaborationRecreated(name string, collaboration *cleanrooms.GetCollaborationOutput) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		err := checkCollaborationIsSame(name, collaboration, state)
		if err == nil {
			return fmt.Errorf("Collaboration was expected to be recreated but was updated")
		}
		return nil
	}
}

func checkCollaborationIsSame(name string, collaboration *cleanrooms.GetCollaborationOutput, s *terraform.State) error {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameCollaboration, name, errors.New("not found"))
	}

	if rs.Primary.ID == "" {
		return create.Error(names.CleanRooms, create.ErrActionCheckingExistence, tfcleanrooms.ResNameCollaboration, name, errors.New("not set"))
	}
	if rs.Primary.ID != *collaboration.Collaboration.Id {
		return fmt.Errorf("New collaboration: %s created instead of updating: %s", rs.Primary.ID, *collaboration.Collaboration.Id)
	}
	return nil
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)

	input := &cleanrooms.ListCollaborationsInput{}
	_, err := conn.ListCollaborations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testCheckCreatorMember(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)
		collaboration, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Collaboration: %s not found in resources", name)
		}
		membersOut, err := conn.ListMembers(ctx, &cleanrooms.ListMembersInput{
			CollaborationIdentifier: &collaboration.Primary.ID,
		})
		if err != nil {
			return err
		}
		if len(membersOut.MemberSummaries) != 1 {
			return fmt.Errorf("Expected 1 member but found %d", len(membersOut.MemberSummaries))
		}
		member := membersOut.MemberSummaries[0]
		if *member.AccountId != acctest.AccountID() {
			return fmt.Errorf("Member account id %s does not match expected value", acctest.AccountID())
		}
		if member.Status != types.MemberStatusInvited {
			return fmt.Errorf("Member status: %s does not match expected value", member.Status)
		}
		if *member.DisplayName != "creator" {
			return fmt.Errorf("member ")
		}
		expectedAbilities := []types.MemberAbility{types.MemberAbilityCanQuery, types.MemberAbilityCanReceiveResults}
		if !reflect.DeepEqual(member.Abilities, expectedAbilities) {
			return fmt.Errorf("Member abilities: %s do not match expected values: %s", member.Abilities, expectedAbilities)
		}

		return nil
	}
}

func testAccCollaborationTags(ctx context.Context, name string, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CleanRoomsClient(ctx)
		collaboration, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Collaboration: %s not found in resources", name)
		}
		tagsOut, err := conn.ListTagsForResource(ctx, &cleanrooms.ListTagsForResourceInput{
			ResourceArn: aws.String(collaboration.Primary.Attributes[names.AttrARN]),
		})
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(expectedTags, tagsOut.Tags) {
			return fmt.Errorf("Actual tags do not match expected")
		}
		return nil
	}
}

const TEST_NAME = names.AttrName
const TEST_DESCRIPTION = names.AttrDescription
const TEST_TAG = "Terraform"
const TEST_MEMBER_ABILITIES = "[\"CAN_QUERY\", \"CAN_RECEIVE_RESULTS\"]"
const TEST_CREATOR_DISPLAY_NAME = "creator"
const TEST_QUERY_LOG_STATUS = "DISABLED"
const TEST_DATA_ENCRYPTION_SETTINGS = `
  data_encryption_metadata {
    allow_clear_text = true
    allow_duplicates = true
    allow_joins_on_columns_with_different_names = true
    preserve_nulls = false
  }
`
const TEST_ADDITIONAL_MEMBER = `
member {
    account_id = 123456789012
    display_name = "OtherMember"
    member_abilities = [] 
  }
`

func testAccCollaborationConfig_basic(rName string, description string, tagValue string) string {
	return testAccCollaboration_configurable(rName, description, tagValue, TEST_MEMBER_ABILITIES,
		TEST_CREATOR_DISPLAY_NAME, TEST_QUERY_LOG_STATUS, TEST_DATA_ENCRYPTION_SETTINGS, "")
}

func testAccCollaborationConfig_additionalMember(rName string, description string, tagValue string) string {
	return testAccCollaboration_configurable(rName, description, tagValue, TEST_MEMBER_ABILITIES,
		TEST_CREATOR_DISPLAY_NAME, TEST_QUERY_LOG_STATUS, TEST_DATA_ENCRYPTION_SETTINGS, TEST_ADDITIONAL_MEMBER)
}

func testAccCollaborationConfig_swapMemberAbilities(rName string, description string, tagValue string) string {
	additionalMember := `
		member {
			account_id = 123456789012
			display_name = "OtherMember"
			member_abilities = ["CAN_QUERY", "CAN_RECEIVE_RESULTS"] 
		}
	`

	return testAccCollaboration_configurable(rName, description, tagValue, "[]",
		TEST_CREATOR_DISPLAY_NAME, TEST_QUERY_LOG_STATUS, TEST_DATA_ENCRYPTION_SETTINGS, additionalMember)
}

func testAccCollaborationConfig_creatorDisplayName(name string, description string, tagValue string, creatorDisplayName string) string {
	return testAccCollaboration_configurable(name, description, tagValue, TEST_MEMBER_ABILITIES,
		creatorDisplayName, TEST_QUERY_LOG_STATUS, TEST_DATA_ENCRYPTION_SETTINGS, "")
}

func testAccCollaborationConfig_queryLogStatus(rName string, description string, tagValue string, queryLogStatus string) string {
	return testAccCollaboration_configurable(rName, description, tagValue, TEST_MEMBER_ABILITIES,
		TEST_CREATOR_DISPLAY_NAME, queryLogStatus, TEST_DATA_ENCRYPTION_SETTINGS, "")
}

func testAccCollaborationConfig_updatedDataEncryptionSettings(name string, description string, tagValue string) string {
	encryptionSettings := `
	data_encryption_metadata {
		allow_clear_text = true
		allow_duplicates = true
		allow_joins_on_columns_with_different_names = true
		preserve_nulls = true
	}
	`
	return testAccCollaboration_configurable(name, description, tagValue, TEST_MEMBER_ABILITIES,
		TEST_CREATOR_DISPLAY_NAME, TEST_QUERY_LOG_STATUS, encryptionSettings, "")
}

func testAccCollaborationConfig_noDataEncryptionSettings(name string, description string, tagValue string) string {
	return testAccCollaboration_configurable(name, description, tagValue, TEST_MEMBER_ABILITIES,
		TEST_CREATOR_DISPLAY_NAME, TEST_QUERY_LOG_STATUS, "", "")
}

func testAccCollaboration_configurable(name string, description string, tagValue string,
	creatorMemberAbilities string, creatorDisplayName string, queryLogStatus string,
	dataEncryptionMetadata string, additionalMember string) string {
	return fmt.Sprintf(`
resource "aws_cleanrooms_collaboration" "test" {
  name                     = %[1]q
  creator_member_abilities = %[4]s
  creator_display_name     = %[5]q
  description              = %[2]q
  query_log_status         = %[6]q

		%[7]s

		%[8]s

  tags = {
    Project = %[3]q
  }
}


	`, name, description, tagValue, creatorMemberAbilities, creatorDisplayName, queryLogStatus,
		dataEncryptionMetadata, additionalMember)
}
