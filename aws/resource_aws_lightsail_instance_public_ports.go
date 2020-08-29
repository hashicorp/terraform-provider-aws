package aws

import (
	"bytes"
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
)

func resourceAwsLightsailInstancePublicPorts() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLightsailInstancePublicPortsCreate,
		Read:   resourceAwsLightsailInstancePublicPortsRead,
		Delete: resourceAwsLightsailInstancePublicPortsDelete,

		Schema: map[string]*schema.Schema{
			"instance_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port_info": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(lightsail.NetworkProtocol_Values(), false),
						},
						"to_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"cidrs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateCIDRNetworkAddress,
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsLightsailInstancePublicPortsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	var portInfos []*lightsail.PortInfo
	if v, ok := d.GetOk("port_info"); ok && v.(*schema.Set).Len() > 0 {
		portInfos = expandLightsailPortInfos(v.(*schema.Set).List())
	}

	input := &lightsail.PutInstancePublicPortsInput{
		InstanceName: aws.String(d.Get("instance_name").(string)),
		PortInfos:    portInfos,
	}

	_, err := conn.PutInstancePublicPorts(input)

	if err != nil {
		return fmt.Errorf("unable to create public ports for instance %s: %w", d.Get("instance_name").(string), err)
	}

	var buffer bytes.Buffer
	for _, portInfo := range portInfos {
		buffer.WriteString(fmt.Sprintf("%s-%d-%d\n", aws.StringValue(portInfo.Protocol), aws.Int64Value(portInfo.FromPort), aws.Int64Value(portInfo.ToPort)))
	}

	d.SetId(fmt.Sprintf("%s-%d", d.Get("instance_name").(string), hashcode.String(buffer.String())))

	return resourceAwsLightsailInstancePublicPortsRead(d, meta)
}

func resourceAwsLightsailInstancePublicPortsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	input := &lightsail.GetInstancePortStatesInput{
		InstanceName: aws.String(d.Get("instance_name").(string)),
	}

	output, err := conn.GetInstancePortStates(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "NotFoundException") {
		log.Printf("[WARN] Lightsail instance public ports (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Lightsail instance public ports (%s): %w", d.Id(), err)
	}

	if err := d.Set("port_info", flattenLightsailInstancePortStates(output.PortStates)); err != nil {
		return fmt.Errorf("error setting port_info: %w", err)
	}

	return nil
}

func resourceAwsLightsailInstancePublicPortsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	var err *multierror.Error

	var portInfos []*lightsail.PortInfo
	if v, ok := d.GetOk("port_info"); ok && v.(*schema.Set).Len() > 0 {
		portInfos = expandLightsailPortInfos(v.(*schema.Set).List())
	}

	for _, portInfo := range portInfos {
		_, portError := conn.CloseInstancePublicPorts(&lightsail.CloseInstancePublicPortsInput{
			InstanceName: aws.String(d.Get("instance_name").(string)),
			PortInfo:     portInfo,
		})

		if portError != nil {
			err = multierror.Append(err, portError)
		}
	}

	if err != nil {
		return fmt.Errorf("unable to close public ports for instance %s: %w", d.Get("instance_name").(string), err)
	}

	return nil
}

func expandLightsailPortInfo(tfMap map[string]interface{}) *lightsail.PortInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &lightsail.PortInfo{
		FromPort: aws.Int64((int64)(tfMap["from_port"].(int))),
		ToPort:   aws.Int64((int64)(tfMap["to_port"].(int))),
		Protocol: aws.String(tfMap["protocol"].(string)),
	}
	if cidrs, ok := tfMap["cidrs"]; ok {
		for _, v := range cidrs.([]interface{}) {
			apiObject.Cidrs = append(apiObject.Cidrs, aws.String(v.(string)))
		}

		sort.Slice(apiObject.Cidrs, func(i, j int) bool {
			return *apiObject.Cidrs[i] > *apiObject.Cidrs[j]
		})
	}

	return apiObject
}

func expandLightsailPortInfos(tfList []interface{}) []*lightsail.PortInfo {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*lightsail.PortInfo

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLightsailPortInfo(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenLightsailInstancePortState(apiObject *lightsail.InstancePortState) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["from_port"] = aws.Int64Value(apiObject.FromPort)
	tfMap["to_port"] = aws.Int64Value(apiObject.ToPort)
	tfMap["protocol"] = aws.StringValue(apiObject.Protocol)

	cidrs := apiObject.Cidrs
	sort.Slice(cidrs, func(i, j int) bool {
		return *cidrs[i] > *cidrs[j]
	})
	tfMap["cidrs"] = aws.StringValueSlice(cidrs)

	return tfMap
}

func flattenLightsailInstancePortStates(apiObjects []*lightsail.InstancePortState) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLightsailInstancePortState(apiObject))
	}

	return tfList
}
