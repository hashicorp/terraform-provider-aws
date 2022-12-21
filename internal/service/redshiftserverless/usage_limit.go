package redshiftserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"amount": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"breach_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      redshiftserverless.UsageLimitBreachActionLog,
				ValidateFunc: validation.StringInSlice(redshiftserverless.UsageLimitBreachAction_Values(), false),
			},
			"period": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      redshiftserverless.UsageLimitPeriodMonthly,
				ValidateFunc: validation.StringInSlice(redshiftserverless.UsageLimitPeriod_Values(), false),
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"usage_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(redshiftserverless.UsageLimitUsageType_Values(), false),
			},
		},
	}
}

func resourceUsageLimitCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	input := redshiftserverless.CreateUsageLimitInput{
		Amount:      aws.Int64(int64(d.Get("amount").(int))),
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
		UsageType:   aws.String(d.Get("usage_type").(string)),
	}

	if v, ok := d.GetOk("period"); ok {
		input.Period = aws.String(v.(string))
	}

	if v, ok := d.GetOk("breach_action"); ok {
		input.BreachAction = aws.String(v.(string))
	}

	out, err := conn.CreateUsageLimit(&input)

	if err != nil {
		return fmt.Errorf("error creating Redshift Serverless Usage Limit : %w", err)
	}

	d.SetId(aws.StringValue(out.UsageLimit.UsageLimitId))

	return resourceUsageLimitRead(d, meta)
}

func resourceUsageLimitRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	out, err := FindUsageLimitByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless UsageLimit (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Serverless Usage Limit (%s): %w", d.Id(), err)
	}

	d.Set("arn", out.UsageLimitArn)
	d.Set("breach_action", out.BreachAction)
	d.Set("period", out.Period)
	d.Set("usage_type", out.UsageType)
	d.Set("resource_arn", out.ResourceArn)
	d.Set("amount", out.Amount)

	return nil
}

func resourceUsageLimitUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	input := &redshiftserverless.UpdateUsageLimitInput{
		UsageLimitId: aws.String(d.Id()),
	}

	if d.HasChange("amount") {
		input.Amount = aws.Int64(int64(d.Get("amount").(int)))
	}

	if d.HasChange("breach_action") {
		input.BreachAction = aws.String(d.Get("breach_action").(string))
	}

	_, err := conn.UpdateUsageLimit(input)
	if err != nil {
		return fmt.Errorf("error updating Redshift Serverless Usage Limit (%s): %w", d.Id(), err)
	}

	return resourceUsageLimitRead(d, meta)
}

func resourceUsageLimitDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	log.Printf("[DEBUG] Deleting Redshift Serverless Usage Limit: %s", d.Id())
	_, err := conn.DeleteUsageLimit(&redshiftserverless.DeleteUsageLimitInput{
		UsageLimitId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Redshift Serverless Usage Limit (%s): %w", d.Id(), err)
	}

	return nil
}
