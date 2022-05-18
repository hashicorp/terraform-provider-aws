package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// There is no constant in the SDK for this resource type
	resourceTypeCapacityReservation = "capacity-reservation"
)

func ResourceCapacityReservation() *schema.Resource {
	return &schema.Resource{
		Create: resourceCapacityReservationCreate,
		Read:   resourceCapacityReservationRead,
		Update: resourceCapacityReservationUpdate,
		Delete: resourceCapacityReservationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"end_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"end_date_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.EndDateTypeUnlimited,
				ValidateFunc: validation.StringInSlice(ec2.EndDateType_Values(), false),
			},
			"ephemeral_storage": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"instance_match_criteria": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.InstanceMatchCriteriaOpen,
				ValidateFunc: validation.StringInSlice(ec2.InstanceMatchCriteria_Values(), false),
			},
			"instance_platform": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.CapacityReservationInstancePlatform_Values(), false),
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"outpost_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"tenancy": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.CapacityReservationTenancyDefault,
				ValidateFunc: validation.StringInSlice(ec2.CapacityReservationTenancy_Values(), false),
			},
		},
	}
}

func resourceCapacityReservationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateCapacityReservationInput{
		AvailabilityZone:  aws.String(d.Get("availability_zone").(string)),
		EndDateType:       aws.String(d.Get("end_date_type").(string)),
		InstanceCount:     aws.Int64(int64(d.Get("instance_count").(int))),
		InstancePlatform:  aws.String(d.Get("instance_platform").(string)),
		InstanceType:      aws.String(d.Get("instance_type").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, resourceTypeCapacityReservation),
	}

	if v, ok := d.GetOk("ebs_optimized"); ok {
		input.EbsOptimized = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("end_date"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))

		input.EndDate = aws.Time(v)
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok {
		input.EphemeralStorage = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("instance_match_criteria"); ok {
		input.InstanceMatchCriteria = aws.String(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tenancy"); ok {
		input.Tenancy = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Capacity Reservation: %s", input)
	output, err := conn.CreateCapacityReservation(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Capacity Reservation: %w", err)
	}

	d.SetId(aws.StringValue(output.CapacityReservation.CapacityReservationId))

	if _, err := WaitCapacityReservationActive(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Capacity Reservation (%s) create: %w", d.Id(), err)
	}

	return resourceCapacityReservationRead(d, meta)
}

func resourceCapacityReservationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	reservation, err := FindCapacityReservationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Capacity Reservation %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Capacity Reservation (%s): %w", d.Id(), err)
	}

	d.Set("arn", reservation.CapacityReservationArn)
	d.Set("availability_zone", reservation.AvailabilityZone)
	d.Set("ebs_optimized", reservation.EbsOptimized)
	if reservation.EndDate != nil {
		d.Set("end_date", aws.TimeValue(reservation.EndDate).Format(time.RFC3339))
	} else {
		d.Set("end_date", nil)
	}
	d.Set("end_date_type", reservation.EndDateType)
	d.Set("ephemeral_storage", reservation.EphemeralStorage)
	d.Set("instance_count", reservation.TotalInstanceCount)
	d.Set("instance_match_criteria", reservation.InstanceMatchCriteria)
	d.Set("instance_platform", reservation.InstancePlatform)
	d.Set("instance_type", reservation.InstanceType)
	d.Set("outpost_arn", reservation.OutpostArn)
	d.Set("owner_id", reservation.OwnerId)
	d.Set("tenancy", reservation.Tenancy)

	tags := KeyValueTags(reservation.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCapacityReservationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyCapacityReservationInput{
			CapacityReservationId: aws.String(d.Id()),
			EndDateType:           aws.String(d.Get("end_date_type").(string)),
			InstanceCount:         aws.Int64(int64(d.Get("instance_count").(int))),
		}

		if v, ok := d.GetOk("end_date"); ok {
			v, _ := time.Parse(time.RFC3339, v.(string))

			input.EndDate = aws.Time(v)
		}

		log.Printf("[DEBUG] Updating EC2 Capacity Reservation: %s", input)
		_, err := conn.ModifyCapacityReservation(input)

		if err != nil {
			return fmt.Errorf("error updating EC2 Capacity Reservation (%s): %w", d.Id(), err)
		}

		if _, err := WaitCapacityReservationActive(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for EC2 Capacity Reservation (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceCapacityReservationRead(d, meta)
}

func resourceCapacityReservationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Capacity Reservation: %s", d.Id())
	_, err := conn.CancelCapacityReservation(&ec2.CancelCapacityReservationInput{
		CapacityReservationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCapacityReservationIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Capacity Reservation (%s): %w", d.Id(), err)
	}

	if _, err := WaitCapacityReservationDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Capacity Reservation (%s) delete: %w", d.Id(), err)
	}

	return nil
}
