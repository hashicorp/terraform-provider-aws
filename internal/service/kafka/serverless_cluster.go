package kafka

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceServerlessCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServerlessClusterCreate,
		ReadWithoutTimeout:   resourceServerlessClusterRead,
		UpdateWithoutTimeout: resourceServerlessClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_authentication": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sasl": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"iam": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceServerlessClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("cluster_name").(string)
	input := &kafka.CreateClusterV2Input{
		ClusterName: aws.String(name),
		Tags:        Tags(tags.IgnoreAWS()),
	}

	output, err := conn.CreateClusterV2WithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating MSK Serverless Cluster (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ClusterArn))

	_, err = waitClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("waiting for MSK Serverless Cluster (%s) create: %s", d.Id(), err)
	}

	return resourceServerlessClusterRead(ctx, d, meta)
}

func resourceServerlessClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	cluster, err := FindServerlessClusterByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK Serverless Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading MSK Serverless Cluster (%s): %s", d.Id(), err)
	}

	d.Set("arn", cluster.ClusterArn)
	d.Set("cluster_name", cluster.ClusterName)

	tags := KeyValueTags(cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceServerlessClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTagsWithContext(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating MSK Serverless Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceServerlessClusterRead(ctx, d, meta)
}
