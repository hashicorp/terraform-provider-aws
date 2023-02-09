package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceInstanceStorageConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceStorageConfigRead,
		Schema: map[string]*schema.Schema{
			"association_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(connect.InstanceStorageResourceType_Values(), false),
			},
			"storage_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kinesis_firehose_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"firehose_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"kinesis_stream_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"stream_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"kinesis_video_stream_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"encryption_config": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"key_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"prefix": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"retention_period_hours": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"s3_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"bucket_prefix": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"encryption_config": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"key_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"storage_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceInstanceStorageConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn()

	associationId := d.Get("association_id").(string)
	instanceId := d.Get("instance_id").(string)
	resourceType := d.Get("resource_type").(string)

	input := &connect.DescribeInstanceStorageConfigInput{
		AssociationId: aws.String(associationId),
		InstanceId:    aws.String(instanceId),
		ResourceType:  aws.String(resourceType),
	}

	resp, err := conn.DescribeInstanceStorageConfigWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("getting Connect Instance Storage Config for Connect Instance (%s,%s,%s): %s", associationId, instanceId, resourceType, err)
	}

	if resp == nil || resp.StorageConfig == nil {
		return diag.Errorf("getting Connect Instance Storage Config: empty response")
	}

	storageConfig := resp.StorageConfig

	if err := d.Set("storage_config", flattenStorageConfig(storageConfig)); err != nil {
		return diag.Errorf("setting storage_config: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", instanceId, associationId, resourceType))

	return nil
}
