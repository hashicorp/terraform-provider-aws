package networkmanager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDevice() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeviceCreate,
		ReadWithoutTimeout:   resourceDeviceRead,
		UpdateWithoutTimeout: resourceDeviceUpdate,
		DeleteWithoutTimeout: resourceDeviceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parsedARN, err := arn.Parse(d.Id())

				if err != nil {
					return nil, fmt.Errorf("error parsing ARN (%s): %w", d.Id(), err)
				}

				// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_networkmanager.html#networkmanager-resources-for-iam-policies.
				resourceParts := strings.Split(parsedARN.Resource, "/")

				if actual, expected := len(resourceParts), 3; actual < expected {
					return nil, fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, d.Id(), actual)
				}

				d.SetId(resourceParts[2])
				d.Set("global_network_id", resourceParts[1])

				return []*schema.ResourceData{d}, nil
			},
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_location": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_arn": {
							Type:          schema.TypeString,
							Optional:      true,
							ValidateFunc:  verify.ValidARN,
							ConflictsWith: []string{"aws_location.0.zone"},
						},
						"zone": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"aws_location.0.subnet_arn"},
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"latitude": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"longitude": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
					},
				},
			},
			"model": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"serial_number": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"site_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"vendor": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
		},
	}
}

func resourceDeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	globalNetworkID := d.Get("global_network_id").(string)

	input := &networkmanager.CreateDeviceInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("aws_location"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AWSLocation = expandAWSLocation(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("location"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Location = expandLocation(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("model"); ok {
		input.Model = aws.String(v.(string))
	}

	if v, ok := d.GetOk("serial_number"); ok {
		input.SerialNumber = aws.String(v.(string))
	}

	if v, ok := d.GetOk("site_id"); ok {
		input.SiteId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		input.Type = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vendor"); ok {
		input.Vendor = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Network Manager Device: %s", input)
	output, err := conn.CreateDeviceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Network Manager Device: %s", err)
	}

	d.SetId(aws.StringValue(output.Device.DeviceId))

	if _, err := waitDeviceCreated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Network Manager Device (%s) create: %s", d.Id(), err)
	}

	return resourceDeviceRead(ctx, d, meta)
}

func resourceDeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	device, err := FindDeviceByTwoPartKey(ctx, conn, globalNetworkID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Device %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Network Manager Device (%s): %s", d.Id(), err)
	}

	d.Set("arn", device.DeviceArn)
	if device.AWSLocation != nil {
		if err := d.Set("aws_location", []interface{}{flattenAWSLocation(device.AWSLocation)}); err != nil {
			return diag.Errorf("error setting aws_location: %s", err)
		}
	} else {
		d.Set("aws_location", nil)
	}
	d.Set("description", device.Description)
	d.Set("global_network_id", device.GlobalNetworkId)
	if device.Location != nil {
		if err := d.Set("location", []interface{}{flattenLocation(device.Location)}); err != nil {
			return diag.Errorf("error setting location: %s", err)
		}
	} else {
		d.Set("location", nil)
	}
	d.Set("model", device.Model)
	d.Set("serial_number", device.SerialNumber)
	d.Set("site_id", device.SiteId)
	d.Set("type", device.Type)
	d.Set("vendor", device.Vendor)

	tags := KeyValueTags(device.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceDeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	if d.HasChangesExcept("tags", "tags_all") {
		globalNetworkID := d.Get("global_network_id").(string)
		input := &networkmanager.UpdateDeviceInput{
			Description:     aws.String(d.Get("description").(string)),
			DeviceId:        aws.String(d.Id()),
			GlobalNetworkId: aws.String(globalNetworkID),
			Model:           aws.String(d.Get("model").(string)),
			SerialNumber:    aws.String(d.Get("serial_number").(string)),
			SiteId:          aws.String(d.Get("site_id").(string)),
			Type:            aws.String(d.Get("type").(string)),
			Vendor:          aws.String(d.Get("vendor").(string)),
		}

		if v, ok := d.GetOk("aws_location"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.AWSLocation = expandAWSLocation(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("location"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Location = expandLocation(v.([]interface{})[0].(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating Network Manager Device: %s", input)
		_, err := conn.UpdateDeviceWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error updating Network Manager Device (%s): %s", d.Id(), err)
		}

		if _, err := waitDeviceUpdated(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("error waiting for Network Manager Device (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating Network Manager Device (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceDeviceRead(ctx, d, meta)
}

func resourceDeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	globalNetworkID := d.Get("global_network_id").(string)

	log.Printf("[DEBUG] Deleting Network Manager Device: %s", d.Id())
	_, err := conn.DeleteDeviceWithContext(ctx, &networkmanager.DeleteDeviceInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		DeviceId:        aws.String(d.Id()),
	})

	if globalNetworkIDNotFoundError(err) || tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting Network Manager Device (%s): %s", d.Id(), err)
	}

	if _, err := waitDeviceDeleted(ctx, conn, globalNetworkID, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("error waiting for Network Manager Device (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindDevice(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetDevicesInput) (*networkmanager.Device, error) {
	output, err := FindDevices(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindDevices(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetDevicesInput) ([]*networkmanager.Device, error) {
	var output []*networkmanager.Device

	err := conn.GetDevicesPagesWithContext(ctx, input, func(page *networkmanager.GetDevicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Devices {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if globalNetworkIDNotFoundError(err) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindDeviceByTwoPartKey(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, deviceID string) (*networkmanager.Device, error) {
	input := &networkmanager.GetDevicesInput{
		DeviceIds:       aws.StringSlice([]string{deviceID}),
		GlobalNetworkId: aws.String(globalNetworkID),
	}

	output, err := FindDevice(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalNetworkId) != globalNetworkID || aws.StringValue(output.DeviceId) != deviceID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusDeviceState(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, deviceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDeviceByTwoPartKey(ctx, conn, globalNetworkID, deviceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitDeviceCreated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, deviceID string, timeout time.Duration) (*networkmanager.Device, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.DeviceStatePending},
		Target:  []string{networkmanager.DeviceStateAvailable},
		Timeout: timeout,
		Refresh: statusDeviceState(ctx, conn, globalNetworkID, deviceID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Device); ok {
		return output, err
	}

	return nil, err
}

func waitDeviceDeleted(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, deviceID string, timeout time.Duration) (*networkmanager.Device, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.DeviceStateDeleting},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusDeviceState(ctx, conn, globalNetworkID, deviceID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Device); ok {
		return output, err
	}

	return nil, err
}

func waitDeviceUpdated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, deviceID string, timeout time.Duration) (*networkmanager.Device, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.DeviceStateUpdating},
		Target:  []string{networkmanager.DeviceStateAvailable},
		Timeout: timeout,
		Refresh: statusDeviceState(ctx, conn, globalNetworkID, deviceID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.Device); ok {
		return output, err
	}

	return nil, err
}

func expandAWSLocation(tfMap map[string]interface{}) *networkmanager.AWSLocation { // nosemgrep:aws-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &networkmanager.AWSLocation{}

	if v, ok := tfMap["subnet_arn"].(string); ok {
		apiObject.SubnetArn = aws.String(v)
	}

	if v, ok := tfMap["zone"].(string); ok {
		apiObject.Zone = aws.String(v)
	}

	return apiObject
}

func flattenAWSLocation(apiObject *networkmanager.AWSLocation) map[string]interface{} { // nosemgrep:aws-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SubnetArn; v != nil {
		tfMap["subnet_arn"] = aws.StringValue(v)
	}

	if v := apiObject.Zone; v != nil {
		tfMap["zone"] = aws.StringValue(v)
	}

	return tfMap
}
