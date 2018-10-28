package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsCapacityReservation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCapacityReservationCreate,
		Read:   resourceAwsCapacityReservationRead,
		Update: resourceAwsCapacityReservationUpdate,
		Delete: resourceAwsCapacityReservationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
				Type:     schema.TypeString,
				Optional: true,
			},
			"end_date_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ec2.EndDateTypeUnlimited,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.EndDateTypeUnlimited,
					ec2.EndDateTypeLimited,
				}, false),
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.InstanceMatchCriteriaOpen,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.InstanceMatchCriteriaOpen,
					ec2.InstanceMatchCriteriaTargeted,
				}, false),
			},
			"instance_platform": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.CapacityReservationInstancePlatformLinuxUnix,
					ec2.CapacityReservationInstancePlatformRedHatEnterpriseLinux,
					ec2.CapacityReservationInstancePlatformSuselinux,
					ec2.CapacityReservationInstancePlatformWindows,
					ec2.CapacityReservationInstancePlatformWindowswithSqlserver,
					ec2.CapacityReservationInstancePlatformWindowswithSqlserverEnterprise,
					ec2.CapacityReservationInstancePlatformWindowswithSqlserverStandard,
					ec2.CapacityReservationInstancePlatformWindowswithSqlserverWeb,
				}, false),
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"tenancy": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.CapacityReservationTenancyDefault,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.CapacityReservationTenancyDefault,
					ec2.CapacityReservationTenancyDedicated,
				}, false),
			},
		},
	}
}

func resourceAwsCapacityReservationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	opts := &ec2.CreateCapacityReservationInput{
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		InstanceCount:    aws.Int64(int64(d.Get("instance_count").(int))),
		InstancePlatform: aws.String(d.Get("instance_platform").(string)),
		InstanceType:     aws.String(d.Get("instance_type").(string)),
	}

	if v, ok := d.GetOk("ebs_optimized"); ok {
		opts.EbsOptimized = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("end_date"); ok {
		t, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return fmt.Errorf("Error parsing capacity reservation end date: %s", err.Error())
		}
		opts.EndDate = aws.Time(t)
	}

	if v, ok := d.GetOk("end_date_type"); ok {
		opts.EndDateType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok {
		opts.EphemeralStorage = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("instance_match_criteria"); ok {
		opts.InstanceMatchCriteria = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tenancy"); ok {
		opts.Tenancy = aws.String(v.(string))
	}

	restricted := meta.(*AWSClient).IsChinaCloud()
	if !restricted {

		tagsSpec := make([]*ec2.TagSpecification, 0)

		if v, ok := d.GetOk("tags"); ok {
			tags := tagsFromMap(v.(map[string]interface{}))

			spec := &ec2.TagSpecification{
				// There is no constant in the SDK for this resource type
				ResourceType: aws.String("capacity-reservation"),
				Tags:         tags,
			}

			tagsSpec = append(tagsSpec, spec)
		}

		if len(tagsSpec) > 0 {
			opts.TagSpecifications = tagsSpec
		}
	}

	log.Printf("[DEBUG] Capacity reservation: %s", opts)

	out, err := conn.CreateCapacityReservation(opts)
	if err != nil {
		return fmt.Errorf("Error creating capacity reservation: %s", err)
	}
	d.SetId(*out.CapacityReservation.CapacityReservationId)
	return resourceAwsCapacityReservationRead(d, meta)
}

func resourceAwsCapacityReservationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resp, err := conn.DescribeCapacityReservations(&ec2.DescribeCapacityReservationsInput{
		CapacityReservationIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		// TODO: Check if error is raised if capacity reservation has gone
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidInstanceID.NotFound" {
			d.SetId("")
			return nil
		}

		// Some other error, report it
		return err
	}

	// If nothing was found, then return no state
	if len(resp.CapacityReservations) == 0 {
		d.SetId("")
		return nil
	}

	reservation := resp.CapacityReservations[0]

	d.Set("availability_zone", reservation.AvailabilityZone)
	d.Set("ebs_optimized", reservation.EbsOptimized)
	d.Set("end_date", reservation.EndDate)
	d.Set("end_date_type", reservation.EndDateType)
	d.Set("ephemeral_storage", reservation.EphemeralStorage)
	d.Set("instance_count", reservation.TotalInstanceCount)
	d.Set("instance_match_criteria", reservation.InstanceMatchCriteria)
	d.Set("instance_platform", reservation.InstancePlatform)
	d.Set("instance_type", reservation.InstanceType)
	d.Set("tags", tagsToMap(reservation.Tags))
	d.Set("tenancy", reservation.Tenancy)

	return nil
}

func resourceAwsCapacityReservationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	d.Partial(true)
	restricted := meta.(*AWSClient).IsChinaCloud()

	if d.HasChange("tags") {
		if !d.IsNewResource() || restricted {
			if err := setTags(conn, d); err != nil {
				return err
			} else {
				d.SetPartial("tags")
			}
		}
	}

	d.Partial(false)

	opts := &ec2.ModifyCapacityReservationInput{
		CapacityReservationId: aws.String(d.Id()),
	}

	if d.HasChange("end_date") {
		if v, ok := d.GetOk("end_date"); ok {
			t, err := time.Parse(time.RFC3339, v.(string))
			if err != nil {
				return fmt.Errorf("Error parsing capacity reservation end date: %s", err.Error())
			}
			opts.EndDate = aws.Time(t)
		}
	}

	if d.HasChange("end_date_type") {
		opts.EndDateType = aws.String(d.Get("end_date_type").(string))
	}

	if d.HasChange("instance_count") {
		opts.InstanceCount = aws.Int64(int64(d.Get("instance_count").(int)))
	}

	log.Printf("[DEBUG] Capacity reservation: %s", opts)

	_, err := conn.ModifyCapacityReservation(opts)
	if err != nil {
		return fmt.Errorf("Error modifying capacity reservation: %s", err)
	}
	return resourceAwsCapacityReservationRead(d, meta)
}

func resourceAwsCapacityReservationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	opts := &ec2.CancelCapacityReservationInput{
		CapacityReservationId: aws.String(d.Id()),
	}

	_, err := conn.CancelCapacityReservation(opts)
	if err != nil {
		return fmt.Errorf("Error cancelling capacity reservation: %s", err)
	}

	return nil
}
