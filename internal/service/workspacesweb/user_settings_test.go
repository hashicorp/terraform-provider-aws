// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workspacesweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkspacesweb "github.com/hashicorp/terraform-provider-aws/internal/service/workspacesweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkSpacesWebUserSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "copy_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "download_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "paste_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "print_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "upload_allowed", "Enabled"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "user_settings_arn", "workspaces-web", regexache.MustCompile(`userSettings/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkspacesweb.ResourceUserSettings, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsConfig_complete(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "copy_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "download_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "paste_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "print_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "upload_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "deep_link_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "disconnect_timeout_in_minutes", "30"),
					resource.TestCheckResourceAttr(resourceName, "idle_disconnect_timeout_in_minutes", "15"),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.toolbar_type", "Docked"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.visual_mode", "Dark"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.0", "Webcam"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.1", "Microphone"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.domain", "example1.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.path", "/path1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.name", "ExampleAllow1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.1.domain", "example2.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.1.path", "/path2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.1.name", "ExampleAllow2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.domain", "blocked1.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.path", "/path3"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.name", "ExampleBlock1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.1.domain", "blocked2.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.1.path", "/path4"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.1.name", "ExampleBlock2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_update(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsConfig_updateBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "copy_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "download_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "paste_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "print_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "upload_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "deep_link_allowed", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "disconnect_timeout_in_minutes", "30"),
					resource.TestCheckResourceAttr(resourceName, "idle_disconnect_timeout_in_minutes", "15"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.toolbar_type", "Docked"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.visual_mode", "Dark"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.0", "Webcam"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.domain", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.path", "/path1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.name", "ExampleAllow"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.domain", "blocked.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.path", "/path2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.name", "ExampleBlock"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
			{
				Config: testAccUserSettingsConfig_updateAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "copy_allowed", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "download_allowed", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "paste_allowed", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "print_allowed", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "upload_allowed", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "deep_link_allowed", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "disconnect_timeout_in_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "idle_disconnect_timeout_in_minutes", "30"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.toolbar_type", "Floating"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.visual_mode", "Light"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.0", "Webcam"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.1", "Microphone"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.domain", "example1.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.path", "/path1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.name", "ExampleAllow1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.1.domain", "example2.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.1.path", "/path2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.1.name", "ExampleAllow2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.domain", "blocked1.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.path", "/path3"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.name", "ExampleBlock1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.1.domain", "blocked2.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.1.path", "/path4"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.1.name", "ExampleBlock2"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_customerManagedKey(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsConfig_customerManagedKey(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_customerManagedKeyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsConfig_customerManagedKeyBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
			{
				Config: testAccUserSettingsConfig_customerManagedKeyAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttrPair(resourceName, "customer_managed_key", kmsKeyResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_additionalEncryptionContextUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsConfig_additionalEncryptionContextBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Development"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Project", "Test"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
			{
				Config: testAccUserSettingsConfig_additionalEncryptionContextAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "additional_encryption_context.Project", "Live"),
				),
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_toolbarConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsConfig_toolbarConfigurationBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.toolbar_type", "Docked"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.visual_mode", "Dark"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.0", "Webcam"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
			{
				Config: testAccUserSettingsConfig_toolbarConfigurationAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.toolbar_type", "Floating"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.visual_mode", "Light"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.0", "Webcam"),
					resource.TestCheckResourceAttr(resourceName, "toolbar_configuration.0.hidden_toolbar_items.1", "Microphone"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_cookieSynchronizationConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSettingsConfig_cookieSynchronizationBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.domain", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.path", "/path1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.name", "ExampleAllow"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
			{
				Config: testAccUserSettingsConfig_cookieSynchronizationAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.domain", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.path", "/path1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.allowlist.0.name", "ExampleAllow"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.domain", "blocked.com"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.path", "/path2"),
					resource.TestCheckResourceAttr(resourceName, "cookie_synchronization_configuration.0.blocklist.0.name", "ExampleBlock"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "user_settings_arn"),
				ImportStateVerifyIdentifierAttribute: "user_settings_arn",
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_upgradeFromV5(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		CheckDestroy: testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.99.0",
					},
				},
				Config: testAccUserSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccUserSettingsConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
		},
	})
}

func TestAccWorkSpacesWebUserSettings_upgradeFromV5AndUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var userSettings awstypes.UserSettings
	resourceName := "aws_workspacesweb_user_settings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkSpacesWebEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.WorkSpacesWebServiceID),
		CheckDestroy: testAccCheckUserSettingsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.99.0",
					},
				},
				Config: testAccUserSettingsConfig_toolbarConfigurationBefore(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("toolbar_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"hidden_toolbar_items": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("Webcam"),
							}),
							"max_display_resolution": knownvalue.Null(),
							"toolbar_type":           knownvalue.StringExact("Docked"),
							"visual_mode":            knownvalue.StringExact("Dark"),
						}),
					})),
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New(names.AttrRegion)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccUserSettingsConfig_toolbarConfigurationAfter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserSettingsExists(ctx, t, resourceName, &userSettings),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("toolbar_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"hidden_toolbar_items": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("Webcam"),
								knownvalue.StringExact("Microphone"),
							}),
							"max_display_resolution": knownvalue.Null(),
							"toolbar_type":           knownvalue.StringExact("Floating"),
							"visual_mode":            knownvalue.StringExact("Light"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
		},
	})
}

func testAccCheckUserSettingsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspacesweb_user_settings" {
				continue
			}

			_, err := tfworkspacesweb.FindUserSettingsByARN(ctx, conn, rs.Primary.Attributes["user_settings_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WorkSpaces Web User Settings %s still exists", rs.Primary.Attributes["user_settings_arn"])
		}

		return nil
	}
}

func testAccCheckUserSettingsExists(ctx context.Context, t *testing.T, n string, v *awstypes.UserSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkSpacesWebClient(ctx)

		output, err := tfworkspacesweb.FindUserSettingsByARN(ctx, conn, rs.Primary.Attributes["user_settings_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccUserSettingsConfig_basic() string {
	return `
resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"
}
`
}

func testAccUserSettingsConfig_complete() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed                       = "Enabled"
  download_allowed                   = "Enabled"
  paste_allowed                      = "Enabled"
  print_allowed                      = "Enabled"
  upload_allowed                     = "Enabled"
  deep_link_allowed                  = "Enabled"
  disconnect_timeout_in_minutes      = 30
  idle_disconnect_timeout_in_minutes = 15
  customer_managed_key               = aws_kms_key.test.arn

  additional_encryption_context = {
    Environment = "Production"
  }

  toolbar_configuration {
    toolbar_type         = "Docked"
    visual_mode          = "Dark"
    hidden_toolbar_items = ["Webcam", "Microphone"]
  }

  cookie_synchronization_configuration {
    allowlist {
      domain = "example1.com"
      path   = "/path1"
      name   = "ExampleAllow1"
    }
    allowlist {
      domain = "example2.com"
      path   = "/path2"
      name   = "ExampleAllow2"
    }
    blocklist {
      domain = "blocked1.com"
      path   = "/path3"
      name   = "ExampleBlock1"
    }
    blocklist {
      domain = "blocked2.com"
      path   = "/path4"
      name   = "ExampleBlock2"
    }
  }
}
`
}

func testAccUserSettingsConfig_updateBefore() string {
	return `
resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed                       = "Enabled"
  download_allowed                   = "Enabled"
  paste_allowed                      = "Enabled"
  print_allowed                      = "Enabled"
  upload_allowed                     = "Enabled"
  deep_link_allowed                  = "Enabled"
  disconnect_timeout_in_minutes      = 30
  idle_disconnect_timeout_in_minutes = 15

  additional_encryption_context = {
    Environment = "Development"
  }

  toolbar_configuration {
    toolbar_type         = "Docked"
    visual_mode          = "Dark"
    hidden_toolbar_items = ["Webcam"]
  }


  cookie_synchronization_configuration {
    allowlist {
      domain = "example.com"
      path   = "/path1"
      name   = "ExampleAllow"
    }
    blocklist {
      domain = "blocked.com"
      path   = "/path2"
      name   = "ExampleBlock"
    }
  }

}
`
}

func testAccUserSettingsConfig_updateAfter() string {
	return `
resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed                       = "Disabled"
  download_allowed                   = "Disabled"
  paste_allowed                      = "Disabled"
  print_allowed                      = "Disabled"
  upload_allowed                     = "Disabled"
  deep_link_allowed                  = "Disabled"
  disconnect_timeout_in_minutes      = 60
  idle_disconnect_timeout_in_minutes = 30

  additional_encryption_context = {
    Environment = "Production"
  }

  toolbar_configuration {
    toolbar_type         = "Floating"
    visual_mode          = "Light"
    hidden_toolbar_items = ["Webcam", "Microphone"]
  }

  cookie_synchronization_configuration {
    allowlist {
      domain = "example1.com"
      path   = "/path1"
      name   = "ExampleAllow1"
    }
    allowlist {
      domain = "example2.com"
      path   = "/path2"
      name   = "ExampleAllow2"
    }
    blocklist {
      domain = "blocked1.com"
      path   = "/path3"
      name   = "ExampleBlock1"
    }
    blocklist {
      domain = "blocked2.com"
      path   = "/path4"
      name   = "ExampleBlock2"
    }
  }
}
`
}

func testAccUserSettingsConfig_customerManagedKey() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed         = "Enabled"
  download_allowed     = "Enabled"
  paste_allowed        = "Enabled"
  print_allowed        = "Enabled"
  upload_allowed       = "Enabled"
  customer_managed_key = aws_kms_key.test.arn
}
`
}

func testAccUserSettingsConfig_toolbarConfigurationBefore() string {
	return `
resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"

  toolbar_configuration {
    toolbar_type         = "Docked"
    visual_mode          = "Dark"
    hidden_toolbar_items = ["Webcam"]
  }
}
`
}

func testAccUserSettingsConfig_toolbarConfigurationAfter() string {
	return `
resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"

  toolbar_configuration {
    toolbar_type         = "Floating"
    visual_mode          = "Light"
    hidden_toolbar_items = ["Webcam", "Microphone"]
  }
}
`
}

func testAccUserSettingsConfig_cookieSynchronizationBefore() string {
	return `
resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"

  cookie_synchronization_configuration {
    allowlist {
      domain = "example.com"
      path   = "/path1"
      name   = "ExampleAllow"
    }
  }
}
`
}

func testAccUserSettingsConfig_cookieSynchronizationAfter() string {
	return `
resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed     = "Enabled"
  download_allowed = "Enabled"
  paste_allowed    = "Enabled"
  print_allowed    = "Enabled"
  upload_allowed   = "Enabled"

  cookie_synchronization_configuration {
    allowlist {
      domain = "example.com"
      path   = "/path1"
      name   = "ExampleAllow"
    }
    blocklist {
      domain = "blocked.com"
      path   = "/path2"
      name   = "ExampleBlock"
    }
  }
}
`
}

func testAccUserSettingsConfig_customerManagedKeyBefore() string {
	return `
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed         = "Enabled"
  download_allowed     = "Enabled"
  paste_allowed        = "Enabled"
  print_allowed        = "Enabled"
  upload_allowed       = "Enabled"
  customer_managed_key = aws_kms_key.test1.arn
}
`
}

func testAccUserSettingsConfig_customerManagedKeyAfter() string {
	return `
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed         = "Enabled"
  download_allowed     = "Enabled"
  paste_allowed        = "Enabled"
  print_allowed        = "Enabled"
  upload_allowed       = "Enabled"
  customer_managed_key = aws_kms_key.test2.arn
}
`
}

func testAccUserSettingsConfig_additionalEncryptionContextBefore() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed         = "Enabled"
  download_allowed     = "Enabled"
  paste_allowed        = "Enabled"
  print_allowed        = "Enabled"
  upload_allowed       = "Enabled"
  customer_managed_key = aws_kms_key.test.arn
  additional_encryption_context = {
    Environment = "Development"
    Project     = "Test"
  }
}
`
}

func testAccUserSettingsConfig_additionalEncryptionContextAfter() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow WorkSpacesWeb to use the key"
        Effect = "Allow"
        Principal = {
          Service = "workspaces-web.amazonaws.com"
        }
        Action = [
          "kms:DescribeKey",
          "kms:GenerateDataKey",
          "kms:GenerateDataKeyWithoutPlaintext",
          "kms:Decrypt",
          "kms:ReEncryptTo",
          "kms:ReEncryptFrom"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_workspacesweb_user_settings" "test" {
  copy_allowed         = "Enabled"
  download_allowed     = "Enabled"
  paste_allowed        = "Enabled"
  print_allowed        = "Enabled"
  upload_allowed       = "Enabled"
  customer_managed_key = aws_kms_key.test.arn
  additional_encryption_context = {
    Environment = "Production"
    Project     = "Live"
  }
}
`
}
