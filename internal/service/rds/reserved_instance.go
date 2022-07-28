package rds

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResReservedInstance = "Reserved Instance"
)

func ResourceReservedInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReservedInstanceCreate,
		ReadContext:   resourceReservedInstanceRead,
		DeleteContext: resourceReservedInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"currency_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"duration": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"fixed_price": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"instance_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"lease_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"offering_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"offering_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recurring_charges": {
				Type:     schema.TypeList,
				Computed: true,
			},
			"start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"usage_price": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReservedInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &rds.PurchaseReservedDBInstancesOfferingInput{
		DBInstanceCount:               aws.Int64(d.Get("instance_cout").(int64)),
		ReservedDBInstancesOfferingId: aws.String(d.Get("offering_id").(string)),
		ReservedDBInstanceId:          aws.String(d.Get("instance_id").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.PurchaseReservedDBInstancesOfferingWithContext(ctx, input)

	if err != nil {
		return names.DiagError(names.RDS, names.ErrActionCreating, ResReservedInstance, d.Id(), err)
	}

	d.SetId(aws.ToString(resp.ReservedDBInstance.ReservedDBInstanceId))

	return resourceReservedInstanceRead(ctx, d, meta)
}

func resourceReservedInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	reservation, err := FindReservedDBInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		names.LogNotFoundRemoveState(names.RDS, names.ErrActionReading, ResReservedInstance, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.DiagError(names.RDS, names.ErrActionReading, ResReservedInstance, d.Id(), err)
	}

	d.Set("arn", reservation.ReservedDBInstanceArn)
	d.Set("currency_code", reservation.CurrencyCode)
	d.Set("duration", reservation.Duration)
	d.Set("fixed_price", reservation.FixedPrice)
	d.Set("instance_class", reservation.DBInstanceClass)
	d.Set("instance_count", reservation.DBInstanceCount)
	d.Set("instance_id", reservation.ReservedDBInstanceId)
	d.Set("lease_id", reservation.LeaseId)
	d.Set("multi_az", reservation.MultiAZ)
	d.Set("offering_id", reservation.ReservedDBInstancesOfferingId)
	d.Set("offering_type", reservation.OfferingType)
	d.Set("product_description", reservation.ProductDescription)
	d.Set("recurring_charges", reservation.RecurringCharges)
	d.Set("start_time", reservation.StartTime)
	d.Set("state", reservation.State)
	d.Set("usage_price", reservation.UsagePrice)

	tags, err := ListTags(conn, aws.ToString(reservation.ReservedDBInstanceArn))
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResTags, d.Id(), err)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return names.DiagError(names.CE, names.ErrActionUpdating, ResTags, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return names.DiagError(names.CE, names.ErrActionUpdating, ResTags, d.Id(), err)
	}

	return nil
}

func resourceReservedInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	// Reservations cannot be deleted. Removing from state.
	log.Printf("[DEBUG] %s %s cannot be deleted. Removing from state.: %s", names.RDS, ResReservedInstance, d.Id())

	return nil
}
