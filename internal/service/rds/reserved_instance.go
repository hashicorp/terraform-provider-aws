package rds

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameReservedInstance = "Reserved Instance"
)

func ResourceReservedInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReservedInstanceCreate,
		ReadWithoutTimeout:   resourceReservedInstanceRead,
		UpdateWithoutTimeout: resourceReservedInstanceUpdate,
		DeleteWithoutTimeout: resourceReservedInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
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
			"db_instance_class": {
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
			"instance_count": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  1,
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
				ForceNew: true,
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recurring_charge_amount": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"recurring_charge_frequency": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"reservation_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &rds.PurchaseReservedDBInstancesOfferingInput{
		ReservedDBInstancesOfferingId: aws.String(d.Get("offering_id").(string)),
	}

	if v, ok := d.Get("instance_count").(int); ok && v > 0 {
		input.DBInstanceCount = aws.Int64(int64(d.Get("instance_count").(int)))
	}

	if v, ok := d.Get("reservation_id").(string); ok && v != "" {
		input.ReservedDBInstanceId = aws.String(v)
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.PurchaseReservedDBInstancesOfferingWithContext(ctx, input)
	if err != nil {
		return create.DiagError(names.RDS, create.ErrActionCreating, ResNameReservedInstance, fmt.Sprintf("offering_id: %s, reservation_id: %s", d.Get("offering_id").(string), d.Get("reservation_id").(string)), err)
	}

	d.SetId(aws.ToString(resp.ReservedDBInstance.ReservedDBInstanceId))

	if err := waitReservedInstanceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.RDS, create.ErrActionWaitingForCreation, ResNameReservedInstance, d.Id(), err)
	}

	return resourceReservedInstanceRead(ctx, d, meta)
}

func resourceReservedInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	reservation, err := FindReservedDBInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.RDS, create.ErrActionReading, ResNameReservedInstance, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.RDS, create.ErrActionReading, ResNameReservedInstance, d.Id(), err)
	}

	d.Set("arn", reservation.ReservedDBInstanceArn)
	d.Set("currency_code", reservation.CurrencyCode)
	d.Set("db_instance_class", reservation.DBInstanceClass)
	d.Set("duration", reservation.Duration)
	d.Set("fixed_price", reservation.FixedPrice)
	d.Set("instance_count", reservation.DBInstanceCount)
	d.Set("lease_id", reservation.LeaseId)
	d.Set("multi_az", reservation.MultiAZ)
	d.Set("offering_id", reservation.ReservedDBInstancesOfferingId)
	d.Set("offering_type", reservation.OfferingType)
	d.Set("product_description", reservation.ProductDescription)
	d.Set("recurring_charges", flattenRecurringCharges(reservation.RecurringCharges))
	d.Set("reservation_id", reservation.ReservedDBInstanceId)
	d.Set("start_time", (reservation.StartTime).Format(time.RFC3339))
	d.Set("state", reservation.State)
	d.Set("usage_price", reservation.UsagePrice)

	tags, err := ListTags(ctx, conn, aws.ToString(reservation.ReservedDBInstanceArn))
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err != nil {
		return create.DiagError(names.CE, create.ErrActionReading, ResNameTags, d.Id(), err)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.CE, create.ErrActionUpdating, ResNameTags, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.CE, create.ErrActionUpdating, ResNameTags, d.Id(), err)
	}

	return nil
}

func resourceReservedInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn()

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return create.DiagError(names.RDS, create.ErrActionUpdating, ResNameTags, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return create.DiagError(names.RDS, create.ErrActionUpdating, ResNameTags, d.Id(), err)
		}
	}

	return resourceReservedInstanceRead(ctx, d, meta)
}

func resourceReservedInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Reservations cannot be deleted. Removing from state.
	log.Printf("[DEBUG] %s %s cannot be deleted. Removing from state.: %s", names.RDS, ResNameReservedInstance, d.Id())

	return nil
}

func flattenRecurringCharges(recurringCharges []*rds.RecurringCharge) []interface{} {
	if len(recurringCharges) == 0 {
		return []interface{}{}
	}

	var rawRecurringCharges []interface{}
	for _, recurringCharge := range recurringCharges {
		rawRecurringCharge := map[string]interface{}{
			"recurring_charge_amount":    recurringCharge.RecurringChargeAmount,
			"recurring_charge_frequency": aws.ToString(recurringCharge.RecurringChargeFrequency),
		}

		rawRecurringCharges = append(rawRecurringCharges, rawRecurringCharge)
	}

	return rawRecurringCharges
}
