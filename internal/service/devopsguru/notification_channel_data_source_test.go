// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccNotificationChannelDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_devopsguru_notification_channel.test"
	notificationChannelResourceName := "aws_devopsguru_notification_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationChannelDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, notificationChannelResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, "sns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "sns.0.topic_name", notificationChannelResourceName, "sns.0.topic_name"),
				),
			},
		},
	})
}

func testAccNotificationChannelDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_devopsguru_notification_channel" "test" {
  sns {
    topic_arn = aws_sns_topic.test.arn
  }
}

data "aws_devopsguru_notification_channel" "test" {
  id = aws_devopsguru_notification_channel.test.id
}
`, rName)
}
