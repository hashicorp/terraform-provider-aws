// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTEventConfigurations_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccEventConfigurations_basic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccEventConfigurations_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iot_event_configurations.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccEventConfigurationsConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_configurations.%", "11"),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.THING", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.THING_GROUP", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.THING_TYPE", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.THING_GROUP_MEMBERSHIP", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.THING_GROUP_HIERARCHY", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.THING_TYPE_ASSOCIATION", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.JOB", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.JOB_EXECUTION", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.POLICY", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.CERTIFICATE", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_configurations.CA_CERTIFICATE", acctest.CtTrue),
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

const testAccEventConfigurationsConfig_basic = `
resource "aws_iot_event_configurations" "test" {
  event_configurations = {
    "THING"                  = true,
    "THING_GROUP"            = false,
    "THING_TYPE"             = false,
    "THING_GROUP_MEMBERSHIP" = false,
    "THING_GROUP_HIERARCHY"  = false,
    "THING_TYPE_ASSOCIATION" = false,
    "JOB"                    = false,
    "JOB_EXECUTION"          = false,
    "POLICY"                 = false,
    "CERTIFICATE"            = true,
    "CA_CERTIFICATE"         = true,
  }
}
`
