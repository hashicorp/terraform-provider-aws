package redshift

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceUsageLimit() *schema.Resource {
	return &schema.Resource{
		Create: resourceUsageLimitCreate,
		Read:   resourceUsageLimitRead,
		Update: resourceUsageLimitUpdate,
		Delete: resourceUsageLimitDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"amount": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"breach_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      redshift.UsageLimitBreachActionLog,
				ValidateFunc: validation.StringInSlice(redshift.UsageLimitBreachAction_Values(), false),
			},
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"feature_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(redshift.UsageLimitFeatureType_Values(), false),
			},
			"limit_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(redshift.UsageLimitLimitType_Values(), false),
			},
			"period": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      redshift.UsageLimitPeriodMonthly,
				ValidateFunc: validation.StringInSlice(redshift.UsageLimitPeriod_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceUsageLimitCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	clusterId := d.Get("cluster_identifier").(string)

	input := redshift.CreateUsageLimitInput{
		Amount:            aws.Int64(int64(d.Get("amount").(int))),
		ClusterIdentifier: aws.String(clusterId),
		FeatureType:       aws.String(d.Get("feature_type").(string)),
		LimitType:         aws.String(d.Get("limit_type").(string)),
	}

	if v, ok := d.GetOk("breach_action"); ok {
		input.BreachAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk("period"); ok {
		input.Period = aws.String(v.(string))
	}

	input.Tags = Tags(tags.IgnoreAWS())

	out, err := conn.CreateUsageLimit(&input)

	if err != nil {
		return fmt.Errorf("error creating Redshift Usage Limit (%s): %s", clusterId, err)
	}

	d.SetId(aws.StringValue(out.UsageLimitId))

	return resourceUsageLimitRead(d, meta)
}

func resourceUsageLimitRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	out, err := FindUsageLimitByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Usage Limit (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Usage Limit (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("usagelimit:%s", d.Id()),
	}.String()

	d.Set("arn", arn)
	d.Set("amount", out.Amount)
	d.Set("period", out.Period)
	d.Set("limit_type", out.LimitType)
	d.Set("feature_type", out.FeatureType)
	d.Set("breach_action", out.BreachAction)
	d.Set("cluster_identifier", out.ClusterIdentifier)

	tags := KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceUsageLimitUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &redshift.ModifyUsageLimitInput{
			UsageLimitId: aws.String(d.Id()),
		}

		if d.HasChange("amount") {
			input.Amount = aws.Int64(int64(d.Get("amount").(int)))
		}

		if d.HasChange("breach_action") {
			input.BreachAction = aws.String(d.Get("breach_action").(string))
		}

		_, err := conn.ModifyUsageLimit(input)
		if err != nil {
			return fmt.Errorf("error updating Redshift Usage Limit (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Usage Limit (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceUsageLimitRead(d, meta)
}

func resourceUsageLimitDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	deleteInput := redshift.DeleteUsageLimitInput{
		UsageLimitId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting snapshot copy grant: %s", d.Id())
	_, err := conn.DeleteUsageLimit(&deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeUsageLimitNotFoundFault) {
			return nil
		}
		return err
	}

	return err
}
