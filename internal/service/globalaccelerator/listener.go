package globalaccelerator

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceListener() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerCreate,
		ReadWithoutTimeout:   resourceListenerRead,
		UpdateWithoutTimeout: resourceListenerUpdate,
		DeleteWithoutTimeout: resourceListenerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"accelerator_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_affinity": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      globalaccelerator.ClientAffinityNone,
				ValidateFunc: validation.StringInSlice(globalaccelerator.ClientAffinity_Values(), false),
			},
			"port_range": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"to_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(globalaccelerator.Protocol_Values(), false),
			},
		},
	}
}

func resourceListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	acceleratorARN := d.Get("accelerator_arn").(string)

	input := &globalaccelerator.CreateListenerInput{
		AcceleratorArn:   aws.String(acceleratorARN),
		ClientAffinity:   aws.String(d.Get("client_affinity").(string)),
		IdempotencyToken: aws.String(resource.UniqueId()),
		PortRanges:       expandPortRanges(d.Get("port_range").(*schema.Set).List()),
		Protocol:         aws.String(d.Get("protocol").(string)),
	}

	resp, err := conn.CreateListenerWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Global Accelerator Listener: %s", err)
	}

	d.SetId(aws.StringValue(resp.Listener.ListenerArn))

	// Creating a listener triggers the accelerator to change status to InPending.
	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()

	listener, err := FindListenerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Global Accelerator Listener (%s): %s", d.Id(), err)
	}

	acceleratorARN, err := ListenerOrEndpointGroupARNToAcceleratorARN(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("accelerator_arn", acceleratorARN)
	d.Set("client_affinity", listener.ClientAffinity)
	if err := d.Set("port_range", flattenPortRanges(listener.PortRanges)); err != nil {
		return diag.Errorf("setting port_range: %s", err)
	}
	d.Set("protocol", listener.Protocol)

	return nil
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	acceleratorARN := d.Get("accelerator_arn").(string)

	input := &globalaccelerator.UpdateListenerInput{
		ClientAffinity: aws.String(d.Get("client_affinity").(string)),
		ListenerArn:    aws.String(d.Id()),
		PortRanges:     expandPortRanges(d.Get("port_range").(*schema.Set).List()),
		Protocol:       aws.String(d.Get("protocol").(string)),
	}

	_, err := conn.UpdateListenerWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating Global Accelerator Listener (%s): %s", d.Id(), err)
	}

	// Updating a listener triggers the accelerator to change status to InPending.
	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn()
	acceleratorARN := d.Get("accelerator_arn").(string)

	log.Printf("[DEBUG] Deleting Global Accelerator Listener: %s", d.Id())
	_, err := conn.DeleteListenerWithContext(ctx, &globalaccelerator.DeleteListenerInput{
		ListenerArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeListenerNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Global Accelerator Listener (%s): %s", d.Id(), err)
	}

	// Deleting a listener triggers the accelerator to change status to InPending.
	if _, err := waitAcceleratorDeployed(ctx, conn, acceleratorARN, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Global Accelerator Accelerator (%s) deployment: %s", acceleratorARN, err)
	}

	return nil
}

func expandPortRange(tfMap map[string]interface{}) *globalaccelerator.PortRange {
	if tfMap == nil {
		return nil
	}

	apiObject := &globalaccelerator.PortRange{}

	if v, ok := tfMap["from_port"].(int); ok && v != 0 {
		apiObject.FromPort = aws.Int64(int64(v))
	}

	if v, ok := tfMap["to_port"].(int); ok && v != 0 {
		apiObject.ToPort = aws.Int64(int64(v))
	}

	return apiObject
}

func expandPortRanges(tfList []interface{}) []*globalaccelerator.PortRange {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*globalaccelerator.PortRange

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPortRange(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenPortRange(apiObject *globalaccelerator.PortRange) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FromPort; v != nil {
		tfMap["from_port"] = aws.Int64Value(v)
	}

	if v := apiObject.ToPort; v != nil {
		tfMap["to_port"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenPortRanges(apiObjects []*globalaccelerator.PortRange) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenPortRange(apiObject))
	}

	return tfList
}
