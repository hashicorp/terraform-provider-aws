// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccQueue_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	originalDescription := "Created"
	updatedDescription := "Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName, rName2, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_basic(rName, rName2, updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
		},
	})
}

func testAccQueue_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName, rName2, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourceQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccQueue_updateHoursOfOperationId(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_hoursOfOperation(rName, rName2, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_hoursOfOperation(rName, rName2, "second"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_hoursOfOperation(rName, rName2, "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
		},
	})
}

func testAccQueue_updateMaxContacts(t *testing.T) {
	t.Skip("A bug in the service API has been reported")

	ctx := acctest.Context(t)

	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	originalMaxContacts := acctest.Ct1
	updatedMaxContacts := acctest.Ct2

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_maxContacts(rName, rName2, originalMaxContacts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "max_contacts", originalMaxContacts),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_maxContacts(rName, rName2, updatedMaxContacts),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "max_contacts", updatedMaxContacts),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
		},
	})
}

func testAccQueue_updateOutboundCallerConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	originalOutboundCallerIdName := "original"
	updatedOutboundCallerIdName := "updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_outboundCaller(rName, rName2, originalOutboundCallerIdName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "outbound_caller_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "outbound_caller_config.0.outbound_caller_id_name", originalOutboundCallerIdName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_outboundCaller(rName, rName2, updatedOutboundCallerIdName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "outbound_caller_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "outbound_caller_config.0.outbound_caller_id_name", updatedOutboundCallerIdName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
		},
	})
}

func testAccQueue_updateStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	originalStatus := connect.QueueStatusEnabled
	updatedStatus := connect.QueueStatusDisabled

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_status(rName, rName2, originalStatus),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, originalStatus),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_status(rName, rName2, updatedStatus),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, updatedStatus),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
		},
	})
}

func testAccQueue_updateQuickConnectIds(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	description := "test queue integrations with quick connects"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// start with no quick connects associated with the queue
				Config: testAccQueueConfig_basic(rName, rName4, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName4),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// associate one quick connect to the queue
				Config: testAccQueueConfig_quickConnect1(rName, rName2, rName3, rName4, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName4),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "quick_connect_ids.0", "aws_connect_quick_connect.test1", "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// associate two quick connects to the queue
				Config: testAccQueueConfig_quickConnect2(rName, rName2, rName3, rName4, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName4),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// remove one quick connect
				Config: testAccQueueConfig_quickConnect1(rName, rName2, rName3, rName4, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "hours_of_operation_id", "data.aws_connect_hours_of_operation.test", "hours_of_operation_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName4),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "quick_connect_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "quick_connect_ids.0", "aws_connect_quick_connect.test1", "quick_connect_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, connect.QueueStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
				),
			},
		},
	})
}

func testAccQueue_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeQueueOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_connect_queue.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccQueueConfig_basic(rName, rName2, names.AttrTags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Queue"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccQueueConfig_tags(rName, rName2, names.AttrTags),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Queue"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccQueueConfig_tagsUpdated(rName, rName2, names.AttrTags),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckQueueExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Queue"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func testAccCheckQueueExists(ctx context.Context, resourceName string, function *connect.DescribeQueueOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Queue not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Queue ID not set")
		}
		instanceID, queueID, err := tfconnect.QueueParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

		params := &connect.DescribeQueueInput{
			QueueId:    aws.String(queueID),
			InstanceId: aws.String(instanceID),
		}

		getFunction, err := conn.DescribeQueueWithContext(ctx, params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckQueueDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_queue" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

			instanceID, queueID, err := tfconnect.QueueParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			params := &connect.DescribeQueueInput{
				QueueId:    aws.String(queueID),
				InstanceId: aws.String(instanceID),
			}

			_, err = conn.DescribeQueueWithContext(ctx, params)

			if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccQueueConfig_base(rName string) string {
	// Use the aws_connect_hours_of_operation data source with the default "Basic Hours" that comes with connect instances.
	// Because if a resource is used, Terraform will not be able to delete it since queues do not have support for the delete api
	// yet but still references hours_of_operation_id. However, using the data source will result in the failure of the
	// disppears test (removed till delete api is available) for the connect instance (We test disappears on the Connect instance
	// instead of the queue since the queue does not support delete). The error is:
	// Step 1/1 error: Error running post-apply plan: exit status 1
	// Error: error finding Connect Hours of Operation Summary by name (Basic Hours): ResourceNotFoundException: Instance not found
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

data "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = "Basic Hours"
}
`, rName)
}

func testAccQueueConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = %[2]q
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, label))
}

func testAccQueueConfig_hoursOfOperation(rName, rName2, selectHoursOfOperationId string) string {
	return acctest.ConfigCompose(
		testAccQueueConfig_base(rName),
		fmt.Sprintf(`
locals {
  select_hours_of_operation_id = %[2]q
}

resource "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = "Example aws_connect_hours_of_operation to test updates"
  time_zone   = "EST"

  config {
    day = "MONDAY"

    end_time {
      hours   = 23
      minutes = 8
    }

    start_time {
      hours   = 8
      minutes = 0
    }
  }
}

resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = "Test update hours_of_operation_id"
  hours_of_operation_id = local.select_hours_of_operation_id == "first" ? data.aws_connect_hours_of_operation.test.hours_of_operation_id : aws_connect_hours_of_operation.test.hours_of_operation_id

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, selectHoursOfOperationId))
}

//lint:ignore U1000 Ignore unused function temporarily
func testAccQueueConfig_maxContacts(rName, rName2, maxContacts string) string {
	return acctest.ConfigCompose(
		testAccQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = "Test update max contacts"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id
  max_contacts          = %[2]q

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, maxContacts))
}

func testAccQueueConfig_outboundCaller(rName, rName2, OutboundCallerIdName string) string {
	return acctest.ConfigCompose(
		testAccQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = "Test update outbound caller config"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  outbound_caller_config {
    outbound_caller_id_name = %[2]q
  }

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, OutboundCallerIdName))
}

func testAccQueueConfig_status(rName, rName2, status string) string {
	return acctest.ConfigCompose(
		testAccQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = "Test update status"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id
  status                = %[2]q

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName2, status))
}

func testAccQueueQuickConnectConfig_base(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_connect_quick_connect" "test1" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = "Test Quick Connect 1"

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
      phone_number = "+12345678912"
    }
  }

  tags = {
    "Name" = "Test Quick Connect 1"
  }
}

resource "aws_connect_quick_connect" "test2" {
  instance_id = aws_connect_instance.test.id
  name        = %[2]q
  description = "Test Quick Connect 2"

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
      phone_number = "+12345678913"
    }
  }

  tags = {
    "Name" = "Test Quick Connect 2"
  }
}
`, rName, rName2)
}

func testAccQueueConfig_quickConnect1(rName, rName2, rName3, rName4, label string) string {
	return acctest.ConfigCompose(
		testAccQueueConfig_base(rName),
		testAccQueueQuickConnectConfig_base(rName2, rName3),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = %[2]q
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  quick_connect_ids = [
    aws_connect_quick_connect.test1.quick_connect_id,
  ]

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName4, label))
}

func testAccQueueConfig_quickConnect2(rName, rName2, rName3, rName4, label string) string {
	return acctest.ConfigCompose(
		testAccQueueConfig_base(rName),
		testAccQueueQuickConnectConfig_base(rName2, rName3),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = %[2]q
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  quick_connect_ids = [
    aws_connect_quick_connect.test1.quick_connect_id,
    aws_connect_quick_connect.test2.quick_connect_id,
  ]

  tags = {
    "Name" = "Test Queue",
  }
}
`, rName4, label))
}

func testAccQueueConfig_tags(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = %[2]q
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  tags = {
    "Name" = "Test Queue",
    "Key2" = "Value2a"
  }
}
`, rName2, label))
}

func testAccQueueConfig_tagsUpdated(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[1]q
  description           = %[2]q
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  tags = {
    "Name" = "Test Queue",
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName2, label))
}
