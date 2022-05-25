package pinpointsmsvoicev2

import (
	"context"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePhoneNumber() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePhoneNumberCreate,
		ReadWithoutTimeout:   resourcePhoneNumberRead,
		UpdateWithoutTimeout: resourcePhoneNumberUpdate,
		DeleteWithoutTimeout: resourcePhoneNumberDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn_two_way_channel": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"iso_country_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"message_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"monthly_leasing_price": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_capabilities": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				ForceNew: true,
			},
			"number_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"opt_out_list_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Default",
			},
			"phone_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"registration_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"self_managed_opt_outs_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"two_way_channel_arn": {
				Type:     schema.TypeString,
				Optional: true,
				RequiredWith: []string{
					"two_way_channel_enabled",
				},
			},
			"two_way_channel_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				RequiredWith: []string{
					"two_way_channel_arn",
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePhoneNumberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Conn

	in := &pinpointsmsvoicev2.RequestPhoneNumberInput{
		IsoCountryCode:     aws.String(d.Get("iso_country_code").(string)),
		MessageType:        aws.String(d.Get("message_type").(string)),
		NumberCapabilities: flex.ExpandStringList(d.Get("number_capabilities").([]interface{})),
		NumberType:         aws.String(d.Get("number_type").(string)),
	}

	if v, ok := d.GetOk("deletion_protection_enabled"); ok {
		in.DeletionProtectionEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("opt_out_list_name"); ok {
		in.OptOutListName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pool_id"); ok {
		in.PoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("registration_id"); ok {
		in.RegistrationId = aws.String(v.(string))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.RequestPhoneNumberWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("creating Amazon Pinpoint SMS and Voice V2 PhoneNumber: %s", err)
	}

	if out == nil || out.PhoneNumber == nil {
		return diag.Errorf("creating Amazon Pinpoint SMS and Voice V2 PhoneNumber: empty output")
	}

	d.SetId(aws.ToString(out.PhoneNumberId))

	if _, err := waitPhoneNumberCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Amazon Pinpoint SMS and Voice V2 PhoneNumber (%s) create: %s", d.Id(), err)
	}

	if checkUpdateAfterCreateNeeded(d, []string{
		"self_managed_opt_outs_enabled",
		"two_way_channel_arn",
		"two_way_channel_enabled",
	}) {
		if err := resourcePhoneNumberUpdate(ctx, d, meta); err != nil {
			return err
		}
	}

	return resourcePhoneNumberRead(ctx, d, meta)
}

func resourcePhoneNumberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Conn

	out, err := findPhoneNumberByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] PinpointSMSVoiceV2 PhoneNumber (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading PinpointSMSVoiceV2 PhoneNumber (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.PhoneNumberArn)
	d.Set("arn_two_way_channel", out.TwoWayChannelArn)
	d.Set("deletion_protection_enabled", out.DeletionProtectionEnabled)
	d.Set("iso_country_code", out.IsoCountryCode)
	d.Set("message_type", out.MessageType)
	d.Set("monthly_leasing_price", out.MonthlyLeasingPrice)
	d.Set("number_type", out.NumberType)
	d.Set("opt_out_list_name", out.OptOutListName)
	d.Set("phone_number", out.PhoneNumber)
	d.Set("pool_id", out.PoolId)
	d.Set("self_managed_opt_outs_enabled", out.SelfManagedOptOutsEnabled)
	d.Set("two_way_channel_arn", out.TwoWayChannelArn)
	d.Set("two_way_channel_enabled", out.TwoWayEnabled)

	if err := d.Set("number_capabilities", out.NumberCapabilities); err != nil {
		return diag.Errorf("setting complex argument: %s", err)
	}

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return diag.Errorf("listing tags for PinpointSMSVoiceV2 PhoneNumber (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourcePhoneNumberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Conn

	update := false

	in := &pinpointsmsvoicev2.UpdatePhoneNumberInput{
		PhoneNumberId: aws.String(d.Id()),
		TwoWayEnabled: nil,
	}

	if d.HasChanges("deletion_protection_enabled") {
		in.DeletionProtectionEnabled = aws.Bool(d.Get("deletion_protection_enabled").(bool))
		update = true
	}

	if d.HasChanges("opt_out_list_name") {
		in.OptOutListName = aws.String(d.Get("opt_out_list_name").(string))
		update = true
	}

	if d.HasChanges("self_managed_opt_outs_enabled") {
		in.SelfManagedOptOutsEnabled = aws.Bool(d.Get("self_managed_opt_outs_enabled").(bool))
		update = true
	}

	if d.HasChanges("two_way_channel_arn") {
		in.TwoWayChannelArn = aws.String(d.Get("two_way_channel_arn").(string))
		update = true
	}

	if d.HasChanges("two_way_channel_enabled") {
		in.TwoWayEnabled = aws.Bool(d.Get("two_way_channel_enabled").(bool))
		update = true
	}

	if !update {
		// TIP: If update doesn't do anything at all, which is rare, you can
		// return nil. Otherwise, return a read call, as below.
		return nil
	}

	log.Printf("[DEBUG] Updating PinpointSMSVoiceV2 PhoneNumber (%s): %#v", d.Id(), in)
	out, err := conn.UpdatePhoneNumberWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("updating PinpointSMSVoiceV2 PhoneNumber (%s): %s", d.Id(), err)
	}

	if _, err := waitPhoneNumberUpdated(ctx, conn, aws.ToString(out.PhoneNumberId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.Errorf("waiting for PinpointSMSVoiceV2 PhoneNumber (%s) update: %s", d.Id(), err)
	}

	return resourcePhoneNumberRead(ctx, d, meta)
}

func resourcePhoneNumberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PinpointSMSVoiceV2Conn

	log.Printf("[INFO] Deleting PinpointSMSVoiceV2 PhoneNumber %s", d.Id())

	_, err := conn.ReleasePhoneNumberWithContext(ctx, &pinpointsmsvoicev2.ReleasePhoneNumberInput{
		PhoneNumberId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpointsmsvoicev2.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting PinpointSMSVoiceV2 PhoneNumber (%s): %s", d.Id(), err)
	}

	if _, err := waitPhoneNumberDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for PinpointSMSVoiceV2 PhoneNumber (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func waitPhoneNumberCreated(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string, timeout time.Duration) (*pinpointsmsvoicev2.PhoneNumberInformation, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{pinpointsmsvoicev2.NumberStatusActive},
		Refresh:                   statusPhoneNumber(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*pinpointsmsvoicev2.PhoneNumberInformation); ok {
		return out, err
	}

	return nil, err
}

func waitPhoneNumberUpdated(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string, timeout time.Duration) (*pinpointsmsvoicev2.PhoneNumberInformation, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{pinpointsmsvoicev2.NumberStatusActive},
		Refresh:                   statusPhoneNumber(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*pinpointsmsvoicev2.PhoneNumberInformation); ok {
		return out, err
	}

	return nil, err
}

func waitPhoneNumberDeleted(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string, timeout time.Duration) (*pinpointsmsvoicev2.PhoneNumberInformation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{},
		Refresh: statusPhoneNumber(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*pinpointsmsvoicev2.PhoneNumberInformation); ok {
		return out, err
	}

	return nil, err
}

func statusPhoneNumber(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findPhoneNumberByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

func findPhoneNumberByID(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string) (*pinpointsmsvoicev2.PhoneNumberInformation, error) {
	in := &pinpointsmsvoicev2.DescribePhoneNumbersInput{
		PhoneNumberIds: aws.StringSlice([]string{id}),
	}

	out, err := conn.DescribePhoneNumbersWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, pinpointsmsvoicev2.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.PhoneNumbers == nil || len(out.PhoneNumbers) != 1 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.PhoneNumbers[0], nil
}

func checkUpdateAfterCreateNeeded(d *schema.ResourceData, schemaKeys []string) bool {
	for _, schemaKey := range schemaKeys {
		if _, ok := d.GetOk(schemaKey); ok {
			return true
		}
	}

	return false
}
