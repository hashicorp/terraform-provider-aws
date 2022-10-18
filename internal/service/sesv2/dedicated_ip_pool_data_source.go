package sesv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceDedicatedIPPool() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDedicatedIPPoolRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dedicated_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"warmup_percentage": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"warmup_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"pool_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameDedicatedIPPool = "Dedicated IP Pool Data Source"
)

func dataSourceDedicatedIPPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn
	poolName := d.Get("pool_name").(string)

	out, err := FindDedicatedIPPoolByID(ctx, conn, poolName)
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, DSNameDedicatedIPPool, poolName, err)
	}
	d.SetId(poolName)

	poolNameARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("dedicated-ip-pool/%s", d.Id()),
	}.String()
	d.Set("arn", poolNameARN)

	d.Set("dedicated_ips", flattenDedicatedIPs(out.DedicatedIps))

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, DSNameDedicatedIPPool, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.SESV2, create.ErrActionSetting, DSNameDedicatedIPPool, d.Id(), err)
	}

	return nil
}

func flattenDedicatedIPs(apiObjects []types.DedicatedIp) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var dedicatedIps []interface{}
	for _, apiObject := range apiObjects {
		ip := map[string]interface{}{
			"ip":                aws.ToString(apiObject.Ip),
			"warmup_percentage": apiObject.WarmupPercentage,
			"warmup_status":     string(apiObject.WarmupStatus),
		}

		dedicatedIps = append(dedicatedIps, ip)
	}

	return dedicatedIps
}
