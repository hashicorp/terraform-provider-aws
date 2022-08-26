package dynamodb

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceTable() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTableRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attribute": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
					return create.StringHashcode(buf.String())
				},
			},
			"global_secondary_index": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"write_capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"read_capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"hash_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"range_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"projection_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"non_key_attributes": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"hash_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_secondary_index": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"range_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"projection_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"non_key_attributes": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
					return create.StringHashcode(buf.String())
				},
			},
			"range_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"read_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"stream_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stream_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"stream_label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stream_view_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"ttl": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"write_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"replica": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"server_side_encryption": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"kms_key_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"billing_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"point_in_time_recovery": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"table_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	})

	if err != nil {
		return fmt.Errorf("error reading Dynamodb Table (%s): %w", name, err)
	}

	if result == nil || result.Table == nil {
		return fmt.Errorf("error reading Dynamodb Table (%s): not found", name)
	}

	table := result.Table

	d.SetId(aws.StringValue(table.TableName))

	d.Set("arn", table.TableArn)
	d.Set("name", table.TableName)

	if table.BillingModeSummary != nil {
		d.Set("billing_mode", table.BillingModeSummary.BillingMode)
	} else {
		d.Set("billing_mode", dynamodb.BillingModeProvisioned)
	}

	if table.ProvisionedThroughput != nil {
		d.Set("write_capacity", table.ProvisionedThroughput.WriteCapacityUnits)
		d.Set("read_capacity", table.ProvisionedThroughput.ReadCapacityUnits)
	}

	if err := d.Set("attribute", flattenTableAttributeDefinitions(table.AttributeDefinitions)); err != nil {
		return fmt.Errorf("error setting attribute: %w", err)
	}

	for _, attribute := range table.KeySchema {
		if aws.StringValue(attribute.KeyType) == dynamodb.KeyTypeHash {
			d.Set("hash_key", attribute.AttributeName)
		}

		if aws.StringValue(attribute.KeyType) == dynamodb.KeyTypeRange {
			d.Set("range_key", attribute.AttributeName)
		}
	}

	if err := d.Set("local_secondary_index", flattenTableLocalSecondaryIndex(table.LocalSecondaryIndexes)); err != nil {
		return fmt.Errorf("error setting local_secondary_index: %w", err)
	}

	if err := d.Set("global_secondary_index", flattenTableGlobalSecondaryIndex(table.GlobalSecondaryIndexes)); err != nil {
		return fmt.Errorf("error setting global_secondary_index: %w", err)
	}

	if table.StreamSpecification != nil {
		d.Set("stream_view_type", table.StreamSpecification.StreamViewType)
		d.Set("stream_enabled", table.StreamSpecification.StreamEnabled)
	} else {
		d.Set("stream_view_type", "")
		d.Set("stream_enabled", false)
	}

	d.Set("stream_arn", table.LatestStreamArn)
	d.Set("stream_label", table.LatestStreamLabel)

	if err := d.Set("server_side_encryption", flattenTableServerSideEncryption(table.SSEDescription)); err != nil {
		return fmt.Errorf("error setting server_side_encryption: %w", err)
	}

	if err := d.Set("replica", flattenReplicaDescriptions(table.Replicas)); err != nil {
		return fmt.Errorf("error setting replica: %w", err)
	}

	if table.TableClassSummary != nil {
		d.Set("table_class", table.TableClassSummary.TableClass)
	} else {
		d.Set("table_class", nil)
	}

	pitrOut, err := conn.DescribeContinuousBackups(&dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(d.Id()),
	})

	if err != nil && !tfawserr.ErrCodeEquals(err, "UnknownOperationException") {
		return fmt.Errorf("error describing DynamoDB Table (%s) Continuous Backups: %w", d.Id(), err)
	}

	if err := d.Set("point_in_time_recovery", flattenPITR(pitrOut)); err != nil {
		return fmt.Errorf("error setting point_in_time_recovery: %w", err)
	}

	ttlOut, err := conn.DescribeTimeToLive(&dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error describing DynamoDB Table (%s) Time to Live: %w", d.Id(), err)
	}

	if err := d.Set("ttl", flattenTTL(ttlOut)); err != nil {
		return fmt.Errorf("error setting ttl: %w", err)
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil && !tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
		return fmt.Errorf("error listing tags for DynamoDB Table (%s): %w", d.Get("arn").(string), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
