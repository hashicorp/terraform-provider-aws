// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_instance_public_ports")
func ResourceInstancePublicPorts() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstancePublicPortsCreate,
		ReadWithoutTimeout:   resourceInstancePublicPortsRead,
		DeleteWithoutTimeout: resourceInstancePublicPortsDelete,

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
						"cidrs": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							// Default:  []string{"0.0.0.0/0"},
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidCIDRNetworkAddress,
							},
						},
						"cidr_list_aliases": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"from_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"ipv6_cidrs": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							// Default:  []string{"::/0"},
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidCIDRNetworkAddress,
							},
						},
						names.AttrProtocol: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(flattenNetworkProtocolValues(types.NetworkProtocol("").Values()), false),
						},
						"to_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
					},
				},
			},
		},
	}
}

func resourceInstancePublicPortsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	var portInfos []types.PortInfo
	if v, ok := d.GetOk("port_info"); ok && v.(*schema.Set).Len() > 0 {
		portInfos = expandPortInfos(v.(*schema.Set).List())
	}

	input := &lightsail.PutInstancePublicPortsInput{
		InstanceName: aws.String(d.Get("instance_name").(string)),
		PortInfos:    portInfos,
	}

	_, err := conn.PutInstancePublicPorts(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "unable to create public ports for instance %s: %s", d.Get("instance_name").(string), err)
	}

	var buffer bytes.Buffer
	for _, portInfo := range portInfos {
		buffer.WriteString(fmt.Sprintf("%s-%d-%d\n", string(portInfo.Protocol), int64(portInfo.FromPort), int64(portInfo.ToPort)))
	}

	d.SetId(fmt.Sprintf("%s-%d", d.Get("instance_name").(string), create.StringHashcode(buffer.String())))

	return append(diags, resourceInstancePublicPortsRead(ctx, d, meta)...)
}

func resourceInstancePublicPortsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	input := &lightsail.GetInstancePortStatesInput{
		InstanceName: aws.String(d.Get("instance_name").(string)),
	}

	output, err := conn.GetInstancePortStates(ctx, input)

	if !d.IsNewResource() && IsANotFoundError(err) {
		log.Printf("[WARN] Lightsail instance public ports (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lightsail instance public ports (%s): %s", d.Id(), err)
	}

	if output == nil || len(output.PortStates) == 0 || output.PortStates == nil {
		log.Printf("[WARN] Lightsail instance public ports (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err := d.Set("port_info", flattenInstancePortStates(output.PortStates)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting port_info: %s", err)
	}

	return diags
}

func resourceInstancePublicPortsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	var errs []error

	var portInfos []types.PortInfo
	if v, ok := d.GetOk("port_info"); ok && v.(*schema.Set).Len() > 0 {
		portInfos = expandPortInfos(v.(*schema.Set).List())
	}

	for _, portInfo := range portInfos {
		_, portError := conn.CloseInstancePublicPorts(ctx, &lightsail.CloseInstancePublicPortsInput{
			InstanceName: aws.String(d.Get("instance_name").(string)),
			PortInfo:     &portInfo,
		})

		if portError != nil {
			errs = append(errs, portError)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return sdkdiag.AppendErrorf(diags, "unable to close public ports for instance %s: %s", d.Get("instance_name").(string), err)
	}

	return diags
}

func expandPortInfo(tfMap map[string]interface{}) types.PortInfo {
	// if tfMap == nil {
	// 	return nil
	// }

	apiObject := types.PortInfo{
		FromPort: int32(tfMap["from_port"].(int)),
		ToPort:   int32(tfMap["to_port"].(int)),
		Protocol: types.NetworkProtocol(tfMap[names.AttrProtocol].(string)),
	}

	if v, ok := tfMap["cidrs"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Cidrs = aws.ToStringSlice(flex.ExpandStringSet(v))
	}

	if v, ok := tfMap["cidr_list_aliases"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CidrListAliases = aws.ToStringSlice(flex.ExpandStringSet(v))
	}

	if v, ok := tfMap["ipv6_cidrs"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Ipv6Cidrs = aws.ToStringSlice(flex.ExpandStringSet(v))
	}

	return apiObject
}

func expandPortInfos(tfList []interface{}) []types.PortInfo {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.PortInfo

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPortInfo(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenInstancePortState(apiObject types.InstancePortState) map[string]interface{} {
	// if apiObject == (types.InstancePortState{}) {
	// 	return nil
	// }

	tfMap := map[string]interface{}{}

	tfMap["from_port"] = int(apiObject.FromPort)
	tfMap["to_port"] = int(apiObject.ToPort)
	tfMap[names.AttrProtocol] = string(apiObject.Protocol)

	if v := apiObject.Cidrs; v != nil {
		tfMap["cidrs"] = v
	}

	if v := apiObject.CidrListAliases; v != nil {
		tfMap["cidr_list_aliases"] = v
	}

	if v := apiObject.Ipv6Cidrs; v != nil {
		tfMap["ipv6_cidrs"] = v
	}

	return tfMap
}

func flattenInstancePortStates(apiObjects []types.InstancePortState) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		// if apiObject == nil {
		// 	continue
		// }

		tfList = append(tfList, flattenInstancePortState(apiObject))
	}

	return tfList
}

func flattenNetworkProtocolValues(t []types.NetworkProtocol) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}
