package connect

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(connectInstanceCreatedTimeout),
			Delete: schema.DefaultTimeout(connectInstanceDeletedTimeout),
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_resolve_best_voices_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true, //verified default result from ListInstanceAttributes()
			},
			"contact_flow_logs_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false, //verified default result from ListInstanceAttributes()
			},
			"contact_lens_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true, //verified default result from ListInstanceAttributes()
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(12, 12),
				AtLeastOneOf: []string{"directory_id", "instance_alias"},
			},
			"early_media_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true, //verified default result from ListInstanceAttributes()
			},
			"identity_management_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(connect.DirectoryType_Values(), false),
			},
			"inbound_calls_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"instance_alias": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{"directory_id", "instance_alias"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^([\da-zA-Z]+)([\da-zA-Z-]+)$`), "must contain only alphanumeric or hyphen characters"),
					validation.StringDoesNotMatch(regexp.MustCompile(`^(d-).+$`), "can not start with d-"),
				),
			},
			"outbound_calls_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Pre-release feature requiring allow-list from AWS. Removing all functionality until feature is GA
			// "use_custom_tts_voices_enabled": {
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Default:  false, //verified default result from ListInstanceAttributes()
			// },
		},
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	input := &connect.CreateInstanceInput{
		ClientToken:            aws.String(resource.UniqueId()),
		IdentityManagementType: aws.String(d.Get("identity_management_type").(string)),
		InboundCallsEnabled:    aws.Bool(d.Get("inbound_calls_enabled").(bool)),
		OutboundCallsEnabled:   aws.Bool(d.Get("outbound_calls_enabled").(bool)),
	}

	if v, ok := d.GetOk("directory_id"); ok {
		input.DirectoryId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("instance_alias"); ok {
		input.InstanceAlias = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Connect Instance %s", input)
	output, err := conn.CreateInstanceWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Instance (%s): %w", d.Id(), err))
	}

	d.SetId(aws.StringValue(output.Id))

	if _, err := waitInstanceCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Connect instance creation (%s): %w", d.Id(), err))
	}

	for att := range InstanceAttributeMapping() {
		rKey := InstanceAttributeMapping()[att]

		if v, ok := d.GetOk(rKey); ok {
			err := resourceInstanceUpdateAttribute(ctx, conn, d.Id(), att, strconv.FormatBool(v.(bool)))
			//Pre-release attribute, user/account/instance now allow-listed
			if err != nil && tfawserr.ErrCodeEquals(err, ErrCodeAccessDeniedException) || tfawserr.ErrMessageContains(err, ErrCodeAccessDeniedException, "not authorized to update") {
				log.Printf("[WARN] error setting Connect instance (%s) attribute (%s): %s", d.Id(), att, err)
			} else if err != nil {
				return diag.FromErr(fmt.Errorf("error setting Connect instance (%s) attribute (%s): %w", d.Id(), att, err))
			}
		}
	}

	return resourceInstanceRead(ctx, d, meta)
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	for att := range InstanceAttributeMapping() {
		rKey := InstanceAttributeMapping()[att]
		if d.HasChange(rKey) {
			_, n := d.GetChange(rKey)
			err := resourceInstanceUpdateAttribute(ctx, conn, d.Id(), att, strconv.FormatBool(n.(bool)))
			//Pre-release attribute, user/account/instance now allow-listed
			if err != nil && tfawserr.ErrCodeEquals(err, ErrCodeAccessDeniedException) || tfawserr.ErrMessageContains(err, ErrCodeAccessDeniedException, "not authorized to update") {
				log.Printf("[WARN] error setting Connect instance (%s) attribute (%s): %s", d.Id(), att, err)
			} else if err != nil {
				return diag.FromErr(fmt.Errorf("error setting Connect instance (%s) attribute (%s): %s", d.Id(), att, err))
			}
		}
	}

	return nil
}
func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	input := connect.DescribeInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Connect Instance %s", d.Id())
	output, err := conn.DescribeInstanceWithContext(ctx, &input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Connect Instance (%s): %s", d.Id(), err))
	}

	instance := output.Instance

	d.SetId(aws.StringValue(instance.Id))
	d.Set("arn", instance.Arn)
	d.Set("created_time", instance.CreatedTime.Format(time.RFC3339))
	d.Set("identity_management_type", instance.IdentityManagementType)
	d.Set("inbound_calls_enabled", instance.InboundCallsEnabled)
	d.Set("instance_alias", instance.InstanceAlias)
	d.Set("outbound_calls_enabled", instance.OutboundCallsEnabled)
	d.Set("service_role", instance.ServiceRole)
	d.Set("status", instance.InstanceStatus)

	for att := range InstanceAttributeMapping() {
		value, err := resourceInstanceReadAttribute(ctx, conn, d.Id(), att)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error reading Connect instance (%s) attribute (%s): %s", d.Id(), att, err))
		}
		d.Set(InstanceAttributeMapping()[att], value)
	}

	return nil
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn

	input := &connect.DeleteInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Connect Instance %s", d.Id())

	_, err := conn.DeleteInstance(input)

	if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Instance (%s): %s", d.Id(), err))
	}

	if _, err := waitInstanceDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Connect Instance deletion (%s): %s", d.Id(), err))
	}
	return nil
}

func resourceInstanceUpdateAttribute(ctx context.Context, conn *connect.Connect, instanceID string, attributeType string, value string) error {
	input := &connect.UpdateInstanceAttributeInput{
		InstanceId:    aws.String(instanceID),
		AttributeType: aws.String(attributeType),
		Value:         aws.String(value),
	}

	_, err := conn.UpdateInstanceAttributeWithContext(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func resourceInstanceReadAttribute(ctx context.Context, conn *connect.Connect, instanceID string, attributeType string) (bool, error) {
	input := &connect.DescribeInstanceAttributeInput{
		InstanceId:    aws.String(instanceID),
		AttributeType: aws.String(attributeType),
	}

	output, err := conn.DescribeInstanceAttributeWithContext(ctx, input)
	if err != nil {
		return false, err
	}
	result, parseerr := strconv.ParseBool(*output.Attribute.Value)
	return result, parseerr
}
