// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketReplicationConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	iamRoleResourceName := "aws_iam_role.test"
	dstBucketResourceName := "aws_s3_bucket.destination"
	resourceName := "aws_s3_bucket_replication_configuration.test"
	dataSourceName := "data.aws_s3_bucket_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckBucketReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationDataSourceConfig_basic(rName, string(types.StorageClassStandard)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrRole, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtRulePound, "3"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "rule.*", map[string]string{
						names.AttrID:                                        "rule1",
						names.AttrPriority:                                  "1",
						"filter.#":                                          "1",
						"filter.0.prefix":                                   "foo",
						names.AttrStatus:                                    string(types.ReplicationRuleStatusEnabled),
						"delete_marker_replication.#":                       "1",
						"delete_marker_replication.0.status":                string(types.DeleteMarkerReplicationStatusEnabled),
						"destination.#":                                     "1",
						"destination.0.storage_class":                       string(types.StorageClassStandard),
						"destination.0.replication_time.#":                  "1",
						"destination.0.replication_time.0.status":           string(types.ReplicationTimeStatusEnabled),
						"destination.0.replication_time.0.time.#":           "1",
						"destination.0.replication_time.0.time.0.minutes":   "15",
						"destination.0.metrics.#":                           "1",
						"destination.0.metrics.0.status":                    string(types.MetricsStatusEnabled),
						"destination.0.metrics.0.event_threshold.#":         "1",
						"destination.0.metrics.0.event_threshold.0.minutes": "15",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "rule.*", map[string]string{
						names.AttrID:                                        "rule2",
						names.AttrPriority:                                  "2",
						"filter.#":                                          "1",
						"filter.0.and.#":                                    "1",
						"filter.0.and.0.prefix":                             "bar",
						"filter.0.and.0.tag.#":                              "1",
						"filter.0.and.0.tag.0.key":                          "key1",
						"filter.0.and.0.tag.0.value":                        "value1",
						names.AttrStatus:                                    string(types.ReplicationRuleStatusEnabled),
						"delete_marker_replication.#":                       "1",
						"delete_marker_replication.0.status":                string(types.DeleteMarkerReplicationStatusDisabled),
						"destination.#":                                     "1",
						"destination.0.storage_class":                       string(types.StorageClassStandard),
						"destination.0.replication_time.#":                  "1",
						"destination.0.replication_time.0.status":           string(types.ReplicationTimeStatusDisabled),
						"destination.0.replication_time.0.time.#":           "1",
						"destination.0.replication_time.0.time.0.minutes":   "15",
						"destination.0.metrics.#":                           "1",
						"destination.0.metrics.0.status":                    string(types.MetricsStatusDisabled),
						"destination.0.metrics.0.event_threshold.#":         "1",
						"destination.0.metrics.0.event_threshold.0.minutes": "15",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "rule.*", map[string]string{
						names.AttrID:                                        "rule3",
						names.AttrPriority:                                  "3",
						"filter.#":                                          "1",
						"filter.0.tag.#":                                    "1",
						"filter.0.tag.0.key":                                "Key",
						"filter.0.tag.0.value":                              "Value",
						names.AttrStatus:                                    string(types.ReplicationRuleStatusEnabled),
						"delete_marker_replication.#":                       "1",
						"delete_marker_replication.0.status":                string(types.DeleteMarkerReplicationStatusDisabled),
						"destination.#":                                     "1",
						"destination.0.storage_class":                       string(types.StorageClassStandard),
						"destination.0.replication_time.#":                  "1",
						"destination.0.replication_time.0.status":           string(types.ReplicationTimeStatusDisabled),
						"destination.0.replication_time.0.time.#":           "1",
						"destination.0.replication_time.0.time.0.minutes":   "15",
						"destination.0.metrics.#":                           "1",
						"destination.0.metrics.0.status":                    string(types.MetricsStatusDisabled),
						"destination.0.metrics.0.event_threshold.#":         "1",
						"destination.0.metrics.0.event_threshold.0.minutes": "15",
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "rule.*.destination.0.bucket", dstBucketResourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccBucketReplicationConfigurationDataSourceConfig_basic(rName, storageClass string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfigurationConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket_replication_configuration" "test" {
  depends_on = [
    aws_s3_bucket_versioning.source,
    aws_s3_bucket_versioning.destination
  ]

  bucket = aws_s3_bucket.source.id
  role   = aws_iam_role.test.arn

  rule {
    id = "rule1"
    priority = 1
    filter {
      prefix = "foo"
    }
    status = "Enabled"
    delete_marker_replication {
      status = "Enabled"
    }
    destination {
      bucket = aws_s3_bucket.destination.arn
      storage_class = %[1]q
      replication_time {
        status = "Enabled"
        time {
          minutes = 15
        }
      }
      metrics {
        status = "Enabled"
        event_threshold {
          minutes = 15
        }
      }
    }
  }

  rule {
    id = "rule2"
    priority = 2
    filter {
      and {
        prefix = "bar"
        tags = {
          key1 = "value1"
        }
      }
    }
    status = "Enabled"
    delete_marker_replication {
      status = "Disabled"
    }
    destination {
      bucket = aws_s3_bucket.destination.arn
      storage_class = %[1]q
      replication_time {
        status = "Disabled"
        time {
          minutes = 15
        }
      }
      metrics {
        status = "Disabled"
        event_threshold {
          minutes = 15
        }
      }
    }
  }
  rule {
    id = "rule3"
    priority = 3
    filter {
      tag {
        key   = "Key"
        value = "Value"
      }
    }
    status = "Enabled"
    delete_marker_replication {
      status = "Disabled"
    }
    destination {
      bucket = aws_s3_bucket.destination.arn
      storage_class = %[1]q
      replication_time {
        status = "Disabled"
        time {
          minutes = 15
        }
      }
      metrics {
        status = "Disabled"
        event_threshold {
          minutes = 15
        }
      }
    }
  }
}

data "aws_s3_bucket_replication_configuration" "test" {
  bucket = aws_s3_bucket.source.id

  depends_on = [
    aws_s3_bucket_replication_configuration.test,
  ]
}

`, storageClass))
}
