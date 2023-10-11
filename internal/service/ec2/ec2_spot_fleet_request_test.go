// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2SpotFleetRequest_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "Default"),
					resource.TestCheckResourceAttr(resourceName, "valid_until", validUntil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_context(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	contextId := "test_context"

	// Receiving this error is confirmation that the Context ID was included in the spot fleet request.
	errRegexp, err := regexp.Compile(fmt.Sprintf(`Error: creating EC2 Spot Fleet Request: UnauthorizedOperation: The account "\d+" is not allowed to access Context "%s".`, contextId))
	if err != nil {
		t.Fatalf("error compiling expected error regexp: %s", err)
	}

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccSpotFleetRequestConfig_context(rName, publicKey, validUntil, contextId),
				ExpectError: errRegexp,
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_targetCapacityUnitType(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"
	targetCapacityUnitType := "vcpu"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_targetCapacityUnitType(rName, publicKey, validUntil, targetCapacityUnitType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "Default"),
					resource.TestCheckResourceAttr(resourceName, "valid_until", validUntil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity_unit_type", targetCapacityUnitType),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSpotFleetRequest(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_tags1(rName, publicKey, validUntil, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccSpotFleetRequestConfig_tags2(rName, publicKey, validUntil, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSpotFleetRequestConfig_tags1(rName, publicKey, validUntil, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_associatePublicIPAddress(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_associatePublicIPAddress(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"associate_public_ip_address": "true",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_launchTemplate(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchTemplate(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_LaunchTemplate_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchTemplateMultiple(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "2"),
				),
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_launchTemplateWithInstanceTypeOverrides(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchTemplateInstanceTypeOverrides(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_template_config.*", map[string]string{
						"overrides.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_template_config.*.overrides.*", map[string]string{
						"instance_requirements.#": "0",
						"instance_type":           "t1.micro",
						"weighted_capacity":       "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_template_config.*.overrides.*", map[string]string{
						"instance_requirements.#": "0",
						"instance_type":           "m3.medium",
						"priority":                "1",
						"spot_price":              "0.26",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_launchTemplateWithInstanceRequirementsOverrides(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchTemplateInstanceRequirementsOverrides(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_template_config.*", map[string]string{
						"overrides.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_template_config.*.overrides.*", map[string]string{
						"instance_requirements.#":                        "1",
						"instance_requirements.0.instance_generations.#": "1",
						"instance_requirements.0.memory_mib.#":           "1",
						"instance_requirements.0.memory_mib.0.max":       "50000",
						"instance_requirements.0.memory_mib.0.min":       "500",
						"instance_requirements.0.vcpu_count.#":           "1",
						"instance_requirements.0.vcpu_count.0.max":       "8",
						"instance_requirements.0.vcpu_count.0.min":       "1",
						"instance_type": "",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_launchTemplateToLaunchSpec(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchTemplate(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					testAccCheckSpotFleetRequestRecreatedConfig(t, &before, &after),
				),
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_launchSpecToLaunchTemplate(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
				),
			},
			{
				Config: testAccSpotFleetRequestConfig_launchTemplate(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					testAccCheckSpotFleetRequestRecreatedConfig(t, &before, &after),
				),
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_onDemandTargetCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_onDemandTargetCapacity(rName, publicKey, validUntil, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "on_demand_target_capacity", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccSpotFleetRequestConfig_onDemandTargetCapacity(rName, publicKey, validUntil, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "on_demand_target_capacity", "1"),
				),
			},
			{
				Config: testAccSpotFleetRequestConfig_onDemandTargetCapacity(rName, publicKey, validUntil, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "on_demand_target_capacity", "0"),
				),
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_onDemandMaxTotalPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_onDemandMaxTotalPrice(rName, publicKey, validUntil, "0.05"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "on_demand_max_total_price", "0.05"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_onDemandAllocationStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_onDemandAllocationStrategy(rName, publicKey, validUntil, "prioritized"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "on_demand_allocation_strategy", "prioritized"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_instanceInterruptionBehavior(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "stop"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_fleetType(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_type(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "fleet_type", "request"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_iamInstanceProfileARN(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_iamInstanceProfileARN(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					testAccCheckSpotFleetRequest_IAMInstanceProfileARN(&sfr),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_changePriceForcesNewRequest(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccSpotFleetRequestConfig_changeBidPrice(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					testAccCheckSpotFleetRequestRecreatedConfig(t, &before, &after),
				),
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_updateTargetCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccSpotFleetRequestConfig_targetCapacity(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "3"),
				),
			},
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "2"),
				),
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_updateExcessCapacityTerminationPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "Default"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccSpotFleetRequestConfig_excessCapacityTermination(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "NoTermination"),
				),
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_lowestPriceAzOrSubnetInRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_lowestPriceAzInGivenList(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"
	availabilityZonesDataSource := "data.aws_availability_zones.available"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_azs(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.availability_zone", availabilityZonesDataSource, "names.0"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.availability_zone", availabilityZonesDataSource, "names.1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_lowestPriceSubnetInGivenList(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_subnet(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_multipleInstanceTypesInSameAz(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"
	instanceTypeDataSource := "data.aws_ec2_instance_type_offering.available"
	availabilityZonesDataSource := "data.aws_availability_zones.available"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_multipleInstanceTypesinSameAZ(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.availability_zone", availabilityZonesDataSource, "names.0"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.instance_type", instanceTypeDataSource, "instance_type"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"instance_type": "m3.large",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_multipleInstanceTypesInSameSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_multipleInstanceTypesinSameSubnet(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_overridingSpotPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"
	instanceTypeDataSourceName := "data.aws_ec2_instance_type_offering.available"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_overridingPrice(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"spot_price":    "0.05",
						"instance_type": "m3.large",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.instance_type", instanceTypeDataSourceName, "instance_type"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_withoutSpotPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_noPrice(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_diversifiedAllocation(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_diversifiedAllocation(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "allocation_strategy", "diversified"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_multipleInstancePools(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_multipleInstancePools(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "allocation_strategy", "lowestPrice"),
					resource.TestCheckResourceAttr(resourceName, "instance_pools_to_use_count", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_withWeightedCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	fulfillSleep := func() resource.TestCheckFunc {
		// sleep so that EC2 can fuflill the request. We do this to guard against a
		// regression and possible leak where we'll destroy the request and the
		// associated IAM role before anything is actually provisioned and running,
		// thus leaking when those newly started instances are attempted to be
		// destroyed
		// See https://github.com/hashicorp/terraform/pull/8938
		return func(s *terraform.State) error {
			log.Print("[DEBUG] Test: Sleep to allow EC2 to actually begin fulfilling TestAccEC2SpotFleetRequest_withWeightedCapacity request")
			time.Sleep(1 * time.Minute)
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_weightedCapacity(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					fulfillSleep(),
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"weighted_capacity": "3",
						"instance_type":     "r3.large",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"weighted_capacity": "6",
						"instance_type":     "m3.large",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_withEBSDisk(t *testing.T) {
	ctx := acctest.Context(t)
	var config ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_ebs(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &config),
					testAccCheckSpotFleetRequest_EBSAttributes(&config),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_LaunchSpecificationEBSBlockDevice_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var config ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchSpecificationEBSBlockDeviceKMSKeyID(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &config),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_LaunchSpecificationRootBlockDevice_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var config ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchSpecificationRootBlockDeviceKMSKeyID(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &config),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_LaunchSpecification_ebsBlockDeviceGP3(t *testing.T) {
	ctx := acctest.Context(t)
	var config ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchSpecificationEBSBlockDeviceGP3(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &config),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*.ebs_block_device.*", map[string]string{
						"device_name": "/dev/xvdcz",
						"iops":        "4000",
						"throughput":  "500",
						"volume_size": "15",
						"volume_type": "gp3",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_LaunchSpecification_rootBlockDeviceGP3(t *testing.T) {
	ctx := acctest.Context(t)
	var config ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchSpecificationRootBlockDeviceGP3(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &config),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*.root_block_device.*", map[string]string{
						"iops":        "4000",
						"throughput":  "500",
						"volume_size": "15",
						"volume_type": "gp3",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var config ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_tags(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &config),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"tags.%":      "3",
						"tags.First":  "TfAccTest",
						"tags.Second": "Terraform",
						"tags.Name":   rName,
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_placementTenancyAndGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_tenancyGroup(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					testAccCheckSpotFleetRequest_PlacementAttributes(&sfr, rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_withELBs(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_elbs(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_withTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_targetGroups(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_Zero_capacity(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_zeroCapacity(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "2"),
				),
			},
			{
				Config: testAccSpotFleetRequestConfig_zeroCapacity(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "0"),
				),
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_capacityRebalance(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_capacityRebalance(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_maintenance_strategies.0.capacity_rebalance.0.replacement_strategy", "launch"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_instanceStoreAMI(t *testing.T) {
	ctx := acctest.Context(t)
	var config ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_launchSpecificationInstanceStoreAMI(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.0.ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.0.ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.0.root_block_device.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotFleetRequest_noTerminateInstancesWithExpiration(t *testing.T) {
	ctx := acctest.Context(t)
	var sfr ec2.SpotFleetRequestConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSpotFleetRequest(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotFleetRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotFleetRequestConfig_noTerminateInstancesExpiration(rName, publicKey, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotFleetRequestExists(ctx, resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances_on_delete", "true"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instances_with_expiration", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"terminate_instances_on_delete", "wait_for_fulfillment"},
			},
		},
	})
}

func testAccCheckSpotFleetRequestRecreatedConfig(t *testing.T,
	before, after *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.SpotFleetRequestId == after.SpotFleetRequestId {
			t.Fatalf("Expected change of Spot Fleet Request IDs, but both were %v", before.SpotFleetRequestId)
		}
		return nil
	}
}

func testAccCheckSpotFleetRequestExists(ctx context.Context, n string, v *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No EC2 Spot Fleet Request ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		output, err := tfec2.FindSpotFleetRequestByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSpotFleetRequestDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_spot_fleet_request" {
				continue
			}

			_, err := tfec2.FindSpotFleetRequestByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Spot Fleet Request %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSpotFleetRequest_EBSAttributes(sfr *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(sfr.SpotFleetRequestConfig.LaunchSpecifications) == 0 {
			return errors.New("Missing launch specification")
		}

		spec := *sfr.SpotFleetRequestConfig.LaunchSpecifications[0]

		ebs := spec.BlockDeviceMappings
		if len(ebs) < 2 {
			return fmt.Errorf("Expected %d block device mappings, got %d", 2, len(ebs))
		}

		if *ebs[0].DeviceName != "/dev/xvda" {
			return fmt.Errorf("Expected device 0's name to be %s, got %s", "/dev/xvda", *ebs[0].DeviceName)
		}
		if *ebs[1].DeviceName != "/dev/xvdcz" {
			return fmt.Errorf("Expected device 1's name to be %s, got %s", "/dev/xvdcz", *ebs[1].DeviceName)
		}

		return nil
	}
}

func testAccCheckSpotFleetRequest_PlacementAttributes(
	sfr *ec2.SpotFleetRequestConfig, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(sfr.SpotFleetRequestConfig.LaunchSpecifications) == 0 {
			return errors.New("Missing launch specification")
		}

		spec := *sfr.SpotFleetRequestConfig.LaunchSpecifications[0]

		placement := spec.Placement
		if placement == nil {
			return fmt.Errorf("Expected placement to be set, got nil")
		}
		if *placement.Tenancy != ec2.TenancyDedicated {
			return fmt.Errorf("Expected placement tenancy to be %q, got %q", "dedicated", *placement.Tenancy)
		}

		if aws.StringValue(placement.GroupName) != rName {
			return fmt.Errorf("Expected placement group to be %q, got %q", rName, aws.StringValue(placement.GroupName))
		}

		return nil
	}
}

func testAccPreCheckSpotFleetRequest(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

	_, err := tfec2.FindSpotFleetRequests(ctx, conn, &ec2.DescribeSpotFleetRequestsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckSpotFleetRequest_IAMInstanceProfileARN(sfr *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(sfr.SpotFleetRequestConfig.LaunchSpecifications) == 0 {
			return errors.New("Missing launch specification")
		}

		spec := *sfr.SpotFleetRequestConfig.LaunchSpecifications[0]

		profile := spec.IamInstanceProfile
		if profile == nil {
			return fmt.Errorf("Expected IamInstanceProfile to be set, got nil")
		}
		//Validate the string whether it is ARN
		re := regexache.MustCompile(fmt.Sprintf(`arn:%s:iam::\d{12}:instance-profile/?[0-9A-Za-z@-_+=,.].*`, acctest.Partition())) // regex seems suspicious, @-_ is a range
		if !re.MatchString(*profile.Arn) {
			return fmt.Errorf("Expected IamInstanceProfile input as ARN, got %s", *profile.Arn)
		}

		return nil
	}
}

func testAccSpotFleetRequestConfig_base(rName, publicKey string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = "%[2]s"

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

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
          "spotfleet.${data.aws_partition.current.dns_suffix}",
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  path        = "/"
  description = "Spot Fleet Request ACCTest Policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeImages",
        "ec2:DescribeSubnets",
        "ec2:RequestSpotInstances",
        "ec2:TerminateInstances",
        "ec2:DescribeInstanceStatus",
        "iam:PassRole"
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
`, rName, publicKey))
}

func testAccSpotFleetRequestConfig_basic(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_context(rName, publicKey, validUntil, contextId string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  context                             = %[3]q
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil, contextId))
}

func testAccSpotFleetRequestConfig_tags1(rName, publicKey, validUntil, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name

    tags = {
      Name = %[1]q
    }
  }

  tags = {
    %[3]q = %[4]q
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil, tagKey1, tagValue1))
}

func testAccSpotFleetRequestConfig_tags2(rName, publicKey, validUntil, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name

    tags = {
      Name = %[1]q
    }
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccSpotFleetRequestConfig_associatePublicIPAddress(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
    ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name                    = aws_key_pair.test.key_name
    associate_public_ip_address = true

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_targetCapacity(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 3
  valid_until                         = %[2]q
  fleet_type                          = "request"
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_launchTemplate(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test.name
      version = aws_launch_template.test.latest_version
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_launchTemplateMultiple(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = ["t1.micro"]
  }

  preferred_instance_types = ["t1.micro"]
}

resource "aws_launch_template" "test1" {
  name          = "%[1]s-1"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_launch_template" "test2" {
  name          = "%[1]s-2"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.test.instance_type
  key_name      = aws_key_pair.test.key_name

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test1.name
      version = aws_launch_template.test1.latest_version
    }
  }

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test2.name
      version = aws_launch_template.test2.latest_version
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_launchTemplateInstanceTypeOverrides(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test.name
      version = aws_launch_template.test.latest_version
    }

    overrides {
      instance_type     = "t1.micro"
      weighted_capacity = "2"
    }

    overrides {
      instance_type = "m3.medium"
      priority      = 1
      spot_price    = "0.26"
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_launchTemplateInstanceRequirementsOverrides(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test.name
      version = aws_launch_template.test.latest_version
    }

    overrides {
      availability_zone = data.aws_availability_zones.available.names[2]

      instance_requirements {
        vcpu_count {
          min = 1
          max = 8
        }

        memory_mib {
          min = 500
          max = 50000
        }

        instance_generations = ["current"]
      }
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_excessCapacityTermination(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  excess_capacity_termination_policy  = "NoTermination"
  valid_until                         = %[2]q
  fleet_type                          = "request"
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_type(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  fleet_type                          = "request"
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_iamInstanceProfileARN(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_iam_role" "test-role1" {
  name = "tf-test-role1-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.${data.aws_partition.current.dns_suffix}",
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test-role-policy1" {
  name = "tf-test-role-policy1-%[1]s"
  role = aws_iam_role.test-role1.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_instance_profile" "test-iam-instance-profile1" {
  name = "tf-test-profile1-%[1]s"
  role = aws_iam_role.test-role1.name
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.25"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type            = data.aws_ec2_instance_type_offering.available.instance_type
    ami                      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name                 = aws_key_pair.test.key_name
    iam_instance_profile_arn = aws_iam_instance_profile.test-iam-instance-profile1.arn

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_changeBidPrice(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_azs(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[1]

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_subnet(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.20.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 4
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test1.id

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test2.id

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_elbs(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.20.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_elb" "test" {
  name     = %[1]q
  subnets  = [aws_subnet.test1.id, aws_subnet.test2.id]
  internal = true

  listener {
    instance_port     = 80
    instance_protocol = "HTTP"
    lb_port           = 80
    lb_protocol       = "HTTP"
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.5"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true
  load_balancers                      = [aws_elb.test.name]

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test1.id

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_targetGroups(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.20.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_alb" "test" {
  name     = %[1]q
  internal = true
  subnets  = [aws_subnet.test1.id, aws_subnet.test2.id]
}

resource "aws_alb_listener" "test" {
  load_balancer_arn = aws_alb.test.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_alb_target_group.test.arn
    type             = "forward"
  }
}

resource "aws_alb_target_group" "test" {
  name     = aws_alb.test.name
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.5"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true
  target_group_arns                   = [aws_alb_target_group.test.arn]

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test1.id

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_multipleInstanceTypesinSameAZ(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_multipleInstanceTypesinSameSubnet(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 4
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test.id
  }

  launch_specification {
    instance_type = "r3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test.id

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_overridingPrice(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
    spot_price        = "0.05"

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_noPrice(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_multipleInstancePools(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.7"
  target_capacity                     = 30
  valid_until                         = %[2]q
  instance_pools_to_use_count         = 2
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type     = "r3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_diversifiedAllocation(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.7"
  target_capacity                     = 30
  valid_until                         = %[2]q
  allocation_strategy                 = "diversified"
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type     = "r3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_weightedCapacity(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.7"
  target_capacity                     = 10
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
    weighted_capacity = "6"

    tags = {
      Name = %[1]q
    }
  }

  launch_specification {
    instance_type     = "r3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
    weighted_capacity = "3"

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_ebs(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

    ebs_block_device {
      device_name = "/dev/xvda"
      volume_type = "gp2"
      volume_size = "8"
    }

    ebs_block_device {
      device_name = "/dev/xvdcz"
      volume_type = "gp2"
      volume_size = "100"
    }

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_launchSpecificationEBSBlockDeviceKMSKeyID(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  tags = {
    Name = %[1]q
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  terminate_instances_with_expiration = true
  valid_until                         = %[2]q
  wait_for_fulfillment                = true

  launch_specification {
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    instance_type = "t2.micro"

    ebs_block_device {
      device_name = "/dev/xvda"
      volume_type = "gp2"
      volume_size = 8
    }

    ebs_block_device {
      device_name = "/dev/xvdcz"
      encrypted   = true
      kms_key_id  = aws_kms_key.test.arn
      volume_type = "gp2"
      volume_size = 10
    }

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_launchSpecificationRootBlockDeviceKMSKeyID(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  tags = {
    Name = %[1]q
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  terminate_instances_with_expiration = true
  valid_until                         = %[2]q
  wait_for_fulfillment                = true

  launch_specification {
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    instance_type = "t2.micro"

    root_block_device {
      encrypted   = true
      kms_key_id  = aws_kms_key.test.arn
      volume_type = "gp2"
      volume_size = 10
    }

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_launchSpecificationEBSBlockDeviceGP3(rName, publicKey string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    instance_type = "t2.micro"

    ebs_block_device {
      device_name = "/dev/xvda"
      volume_type = "gp2"
      volume_size = 8
    }

    ebs_block_device {
      device_name = "/dev/xvdcz"
      iops        = 4000
      throughput  = 500
      volume_size = 15
      volume_type = "gp3"
    }

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName))
}

func testAccSpotFleetRequestConfig_launchSpecificationRootBlockDeviceGP3(rName, publicKey string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    instance_type = "t2.micro"

    root_block_device {
      iops        = 4000
      throughput  = 500
      volume_size = 15
      volume_type = "gp3"
    }

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName))
}

func testAccSpotFleetRequestConfig_launchSpecificationInstanceStoreAMI(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(
		testAccAMIDataSourceConfig_latestAmazonLinuxHVMInstanceStore(),
		testAccSpotFleetRequestConfig_base(rName, publicKey),
		fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    ami           = data.aws_ami.amzn-ami-minimal-hvm-instance-store.id
    instance_type = "c3.large"

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_tags(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

    tags = {
      First  = "TfAccTest"
      Second = "Terraform"
      Name   = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_tenancyGroup(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    placement_tenancy = "dedicated"
    placement_group   = aws_placement_group.test.name

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_zeroCapacity(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 0
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_capacityRebalance(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  spot_maintenance_strategies {
    capacity_rebalance {
      replacement_strategy = "launch"
    }
  }

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_onDemandTargetCapacity(rName, publicKey, validUntil string, targetCapacity int) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.005"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true
  on_demand_target_capacity           = %[3]d

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test.name
      version = aws_launch_template.test.latest_version
    }
  }

  depends_on = ["aws_iam_policy_attachment.test"]
}
`, rName, validUntil, targetCapacity))
}

func testAccSpotFleetRequestConfig_onDemandMaxTotalPrice(rName, publicKey, validUntil, price string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.005"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true
  on_demand_max_total_price           = %[3]q

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test.name
      version = aws_launch_template.test.latest_version
    }
  }

  depends_on = ["aws_iam_policy_attachment.test"]
}
`, rName, validUntil, price))
}

func testAccSpotFleetRequestConfig_onDemandAllocationStrategy(rName, publicKey, validUntil, strategy string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.005"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true
  on_demand_allocation_strategy       = %[3]q

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test.name
      version = aws_launch_template.test.latest_version
    }
  }

  depends_on = ["aws_iam_policy_attachment.test"]
}
`, rName, validUntil, strategy))
}

func testAccSpotFleetRequestConfig_noTerminateInstancesExpiration(rName, publicKey, validUntil string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                  = aws_iam_role.test.arn
  spot_price                      = "0.05"
  target_capacity                 = 2
  valid_until                     = %[2]q
  terminate_instances_on_delete   = true
  instance_interruption_behaviour = "stop"
  wait_for_fulfillment            = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name

    tags = {
      Name = %[1]q
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil))
}

func testAccSpotFleetRequestConfig_targetCapacityUnitType(rName, publicKey, validUntil, targetCapacityUnitType string) string {
	return acctest.ConfigCompose(testAccSpotFleetRequestConfig_base(rName, publicKey), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = %[1]q
    }
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  target_capacity_unit_type           = %[3]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test.name
      version = aws_launch_template.test.latest_version
    }

    overrides {
      availability_zone = data.aws_availability_zones.available.names[2]

      instance_requirements {
        vcpu_count {
          min = 1
          max = 8
        }

        memory_mib {
          min = 500
          max = 50000
        }

        instance_generations = ["current"]
      }
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil, targetCapacityUnitType))
}
