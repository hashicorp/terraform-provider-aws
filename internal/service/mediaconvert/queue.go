package mediaconvert

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQueueCreate,
		ReadWithoutTimeout:   resourceQueueRead,
		UpdateWithoutTimeout: resourceQueueUpdate,
		DeleteWithoutTimeout: resourceQueueDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pricing_plan": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  mediaconvert.PricingPlanOnDemand,
				ValidateFunc: validation.StringInSlice([]string{
					mediaconvert.PricingPlanOnDemand,
					mediaconvert.PricingPlanReserved,
				}, false),
			},
			"reservation_plan_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"commitment": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								mediaconvert.CommitmentOneYear,
							}, false),
						},
						"renewal_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								mediaconvert.RenewalTypeAutoRenew,
								mediaconvert.RenewalTypeExpire,
							}, false),
						},
						"reserved_slots": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  mediaconvert.QueueStatusActive,
				ValidateFunc: validation.StringInSlice([]string{
					mediaconvert.QueueStatusActive,
					mediaconvert.QueueStatusPaused,
				}, false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn, err := GetAccountClient(meta.(*conns.AWSClient))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Media Convert Account Client: %s", err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	createOpts := &mediaconvert.CreateQueueInput{
		Name:        aws.String(d.Get("name").(string)),
		Status:      aws.String(d.Get("status").(string)),
		PricingPlan: aws.String(d.Get("pricing_plan").(string)),
		Tags:        Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("description"); ok {
		createOpts.Description = aws.String(v.(string))
	}

	if v, ok := d.Get("reservation_plan_settings").([]interface{}); ok && len(v) > 0 && v[0] != nil {
		createOpts.ReservationPlanSettings = expandReservationPlanSettings(v[0].(map[string]interface{}))
	}

	resp, err := conn.CreateQueueWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Media Convert Queue: %s", err)
	}

	d.SetId(aws.StringValue(resp.Queue.Name))

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn, err := GetAccountClient(meta.(*conns.AWSClient))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Media Convert Account Client: %s", err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	getOpts := &mediaconvert.GetQueueInput{
		Name: aws.String(d.Id()),
	}

	resp, err := conn.GetQueueWithContext(ctx, getOpts)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, mediaconvert.ErrCodeNotFoundException) {
		log.Printf("[WARN] Media Convert Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Media Convert Queue: %s", err)
	}

	d.Set("arn", resp.Queue.Arn)
	d.Set("name", resp.Queue.Name)
	d.Set("description", resp.Queue.Description)
	d.Set("pricing_plan", resp.Queue.PricingPlan)
	d.Set("status", resp.Queue.Status)

	if err := d.Set("reservation_plan_settings", flattenReservationPlan(resp.Queue.ReservationPlan)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Media Convert Queue reservation_plan_settings: %s", err)
	}

	tags, err := ListTags(ctx, conn, aws.StringValue(resp.Queue.Arn))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Media Convert Queue (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn, err := GetAccountClient(meta.(*conns.AWSClient))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Media Convert Account Client: %s", err)
	}

	if d.HasChanges("description", "reservation_plan_settings", "status") {
		updateOpts := &mediaconvert.UpdateQueueInput{
			Name:   aws.String(d.Id()),
			Status: aws.String(d.Get("status").(string)),
		}

		if v, ok := d.GetOk("description"); ok {
			updateOpts.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("reservation_plan_settings"); ok {
			reservationPlanSettings := v.([]interface{})[0].(map[string]interface{})
			updateOpts.ReservationPlanSettings = expandReservationPlanSettings(reservationPlanSettings)
		}

		_, err = conn.UpdateQueueWithContext(ctx, updateOpts)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Media Convert Queue (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceQueueRead(ctx, d, meta)...)
}

func resourceQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn, err := GetAccountClient(meta.(*conns.AWSClient))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Media Convert Account Client: %s", err)
	}

	delOpts := &mediaconvert.DeleteQueueInput{
		Name: aws.String(d.Id()),
	}

	_, err = conn.DeleteQueueWithContext(ctx, delOpts)
	if tfawserr.ErrCodeEquals(err, mediaconvert.ErrCodeNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Media Convert Queue (%s): %s", d.Id(), err)
	}

	return diags
}

func GetAccountClient(awsClient *conns.AWSClient) (*mediaconvert.MediaConvert, error) {
	const mutexKey = `mediaconvertaccountconn`
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	if awsClient.MediaConvertAccountConn != nil {
		return awsClient.MediaConvertAccountConn, nil
	}

	input := &mediaconvert.DescribeEndpointsInput{
		Mode: aws.String(mediaconvert.DescribeEndpointsModeDefault),
	}

	output, err := awsClient.MediaConvertConn().DescribeEndpoints(input)

	if err != nil {
		return nil, fmt.Errorf("error describing MediaConvert Endpoints: %w", err)
	}

	if output == nil || len(output.Endpoints) == 0 || output.Endpoints[0] == nil || output.Endpoints[0].Url == nil {
		return nil, fmt.Errorf("error describing MediaConvert Endpoints: empty response or URL")
	}

	endpointURL := aws.StringValue(output.Endpoints[0].Url)

	sess, err := session.NewSession(&awsClient.MediaConvertConn().Config)

	if err != nil {
		return nil, fmt.Errorf("error creating AWS MediaConvert session: %w", err)
	}

	conn := mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(endpointURL)}))

	awsClient.MediaConvertAccountConn = conn

	return conn, nil
}
