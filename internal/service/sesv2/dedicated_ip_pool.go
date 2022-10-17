package sesv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceDedicatedIPPool() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDedicatedIPPoolCreate,
		ReadWithoutTimeout:   resourceDedicatedIPPoolRead,
		UpdateWithoutTimeout: resourceDedicatedIPPoolUpdate,
		DeleteWithoutTimeout: resourceDedicatedIPPoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"pool_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameDedicatedIPPool = "Dedicated IP Pool"
)

func resourceDedicatedIPPoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	in := &sesv2.CreateDedicatedIpPoolInput{
		PoolName: aws.String(d.Get("pool_name").(string)),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateDedicatedIpPool(ctx, in)
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionCreating, ResNameDedicatedIPPool, d.Get("pool_name").(string), err)
	}
	if out == nil {
		return create.DiagError(names.SESV2, create.ErrActionCreating, ResNameDedicatedIPPool, d.Get("pool_name").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("pool_name").(string))
	return resourceDedicatedIPPoolRead(ctx, d, meta)
}

func resourceDedicatedIPPoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	out, err := FindDedicatedIPPoolByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 DedicatedIPPool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, ResNameDedicatedIPPool, d.Id(), err)
	}
	d.Set("pool_name", d.Id())

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
		return create.DiagError(names.SESV2, create.ErrActionReading, ResNameDedicatedIPPool, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.SESV2, create.ErrActionSetting, ResNameDedicatedIPPool, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.SESV2, create.ErrActionSetting, ResNameDedicatedIPPool, d.Id(), err)
	}

	return nil
}

func resourceDedicatedIPPoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	if d.HasChanges("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return create.DiagError(names.SESV2, create.ErrActionUpdating, ResNameDedicatedIPPool, d.Id(), err)
		}
	}

	return resourceDedicatedIPPoolRead(ctx, d, meta)
}

func resourceDedicatedIPPoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	log.Printf("[INFO] Deleting SESV2 DedicatedIPPool %s", d.Id())
	_, err := conn.DeleteDedicatedIpPool(ctx, &sesv2.DeleteDedicatedIpPoolInput{
		PoolName: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil
		}
		return create.DiagError(names.SESV2, create.ErrActionDeleting, ResNameDedicatedIPPool, d.Id(), err)
	}

	return nil
}

func FindDedicatedIPPoolByID(ctx context.Context, conn *sesv2.Client, id string) (*sesv2.GetDedicatedIpsOutput, error) {
	in := &sesv2.GetDedicatedIpsInput{
		PoolName: aws.String(id),
	}
	out, err := conn.GetDedicatedIps(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
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
