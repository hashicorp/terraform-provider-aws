package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediaconvert"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsMediaConvertQueue() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaConvertQueueCreate,
		Read:   resourceAwsMediaConvertQueueRead,
		Update: resourceAwsMediaConvertQueueUpdate,
		Delete: resourceAwsMediaConvertQueueDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsMediaConvertQueueCreate(d *schema.ResourceData, meta interface{}) error {
	originalConn := meta.(*AWSClient).mediaconvertconn

	endpointURL, err := getAwsMediaConvertEndpoint(originalConn)
	if err != nil {
		return fmt.Errorf("Error getting Media Convert Endpoint: %s", err)
	}

	sess, err := session.NewSession(&originalConn.Config)
	if err != nil {
		return fmt.Errorf("Error creating AWS session: %s", err)
	}
	conn := mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(endpointURL)}))

	createOpts := &mediaconvert.CreateQueueInput{
		Name:        aws.String(d.Get("name").(string)),
		Status:      aws.String(d.Get("status").(string)),
		PricingPlan: aws.String(d.Get("pricing_plan").(string)),
		Tags:        tagsFromMapGeneric(d.Get("tags").(map[string]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		createOpts.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("reservation_plan_settings"); ok {
		reservationPlanSettings := v.([]interface{})[0].(map[string]interface{})
		createOpts.ReservationPlanSettings = expandMediaConvertReservationPlanSettings(reservationPlanSettings)
	}

	resp, err := conn.CreateQueue(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Media Convert Queue: %s", err)
	}

	d.SetId(aws.StringValue(resp.Queue.Name))

	return resourceAwsMediaConvertQueueRead(d, meta)
}

func resourceAwsMediaConvertQueueRead(d *schema.ResourceData, meta interface{}) error {
	originalConn := meta.(*AWSClient).mediaconvertconn

	endpointURL, err := getAwsMediaConvertEndpoint(originalConn)
	if err != nil {
		return fmt.Errorf("Error getting Media Convert Endpoint: %s", err)
	}

	sess, err := session.NewSession(&originalConn.Config)
	if err != nil {
		return fmt.Errorf("Error creating AWS session: %s", err)
	}
	conn := mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(endpointURL)}))

	getOpts := &mediaconvert.GetQueueInput{
		Name: aws.String(d.Id()),
	}

	resp, err := conn.GetQueue(getOpts)
	if isAWSErr(err, mediaconvert.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Media Convert Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting Media Convert Queue: %s", err)
	}

	d.Set("arn", resp.Queue.Arn)
	d.Set("name", resp.Queue.Name)
	d.Set("description", resp.Queue.Description)
	d.Set("pricing_plan", resp.Queue.PricingPlan)
	d.Set("status", resp.Queue.Status)

	if err := d.Set("reservation_plan_settings", flattenMediaConvertReservationPlan(resp.Queue.ReservationPlan)); err != nil {
		return fmt.Errorf("Error setting Media Convert Queue reservation_plan_settings: %s", err)
	}

	if err := saveTagsMediaConvert(conn, d, aws.StringValue(resp.Queue.Arn)); err != nil {
		return fmt.Errorf("Error setting Media Convert Queue tags: %s", err)
	}

	return nil
}

func resourceAwsMediaConvertQueueUpdate(d *schema.ResourceData, meta interface{}) error {
	originalConn := meta.(*AWSClient).mediaconvertconn

	endpointURL, err := getAwsMediaConvertEndpoint(originalConn)
	if err != nil {
		return fmt.Errorf("Error getting Media Convert Endpoint: %s", err)
	}

	sess, err := session.NewSession(&originalConn.Config)
	if err != nil {
		return fmt.Errorf("Error creating AWS session: %s", err)
	}
	conn := mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(endpointURL)}))

	updateOpts := &mediaconvert.UpdateQueueInput{
		Name:   aws.String(d.Id()),
		Status: aws.String(d.Get("status").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		updateOpts.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("reservation_plan_settings"); ok {
		reservationPlanSettings := v.([]interface{})[0].(map[string]interface{})
		updateOpts.ReservationPlanSettings = expandMediaConvertReservationPlanSettings(reservationPlanSettings)
	}

	_, err = conn.UpdateQueue(updateOpts)
	if isAWSErr(err, mediaconvert.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Media Convert Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error updating Media Convert Queue: %s", err)
	}

	if err := setTagsMediaConvert(conn, d, d.Get("arn").(string)); err != nil {
		return fmt.Errorf("error updating Media Convert Queue (%s) tags: %s", d.Id(), err)
	}

	return resourceAwsMediaConvertQueueRead(d, meta)
}

func resourceAwsMediaConvertQueueDelete(d *schema.ResourceData, meta interface{}) error {
	originalConn := meta.(*AWSClient).mediaconvertconn

	endpointURL, err := getAwsMediaConvertEndpoint(originalConn)
	if err != nil {
		return fmt.Errorf("Error getting Media Convert Endpoint: %s", err)
	}

	sess, err := session.NewSession(&originalConn.Config)
	if err != nil {
		return fmt.Errorf("Error creating AWS session: %s", err)
	}
	conn := mediaconvert.New(sess.Copy(&aws.Config{Endpoint: aws.String(endpointURL)}))

	delOpts := &mediaconvert.DeleteQueueInput{
		Name: aws.String(d.Id()),
	}

	_, err = conn.DeleteQueue(delOpts)
	if isAWSErr(err, mediaconvert.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting Media Convert Queue: %s", err)
	}

	return nil
}

func getAwsMediaConvertEndpoint(conn *mediaconvert.MediaConvert) (string, error) {
	descOpts := &mediaconvert.DescribeEndpointsInput{
		Mode: aws.String(mediaconvert.DescribeEndpointsModeDefault),
	}

	resp, err := conn.DescribeEndpoints(descOpts)
	if err != nil {
		return "", err
	}

	endpoint := resp.Endpoints[0]

	return aws.StringValue(endpoint.Url), nil
}
