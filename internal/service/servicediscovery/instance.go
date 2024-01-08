// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_service_discovery_instance")
func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstancePut,
		ReadWithoutTimeout:   resourceInstanceRead,
		UpdateWithoutTimeout: resourceInstancePut,
		DeleteWithoutTimeout: resourceInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceInstanceImport,
		},

		Schema: map[string]*schema.Schema{
			"attributes": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ValidateDiagFunc: validation.AllDiag(
					validation.MapKeyLenBetween(1, 255),
					validation.MapKeyMatch(regexache.MustCompile(`^[0-9A-Za-z!-~]+$`), ""),
					validation.MapValueLenBetween(0, 1024),
					validation.MapValueMatch(regexache.MustCompile(`^([0-9A-Za-z!-~][0-9A-Za-z \t!-~]*){0,1}[0-9A-Za-z!-~]{0,1}$`), ""),
				),
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_/:.@-]+$`), ""),
				),
			},
			"service_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceInstancePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	instanceID := d.Get("instance_id").(string)
	input := &servicediscovery.RegisterInstanceInput{
		Attributes:       flex.ExpandStringMap(d.Get("attributes").(map[string]interface{})),
		CreatorRequestId: aws.String(id.UniqueId()),
		InstanceId:       aws.String(instanceID),
		ServiceId:        aws.String(d.Get("service_id").(string)),
	}

	log.Printf("[DEBUG] Registering Service Discovery Instance: %s", input)
	output, err := conn.RegisterInstanceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("registering Service Discovery Instance (%s): %s", instanceID, err)
	}

	d.SetId(instanceID)

	if output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId)); err != nil {
			return diag.Errorf("waiting for Service Discovery Instance (%s) create: %s", d.Id(), err)
		}
	}

	return resourceInstanceRead(ctx, d, meta)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	instance, err := FindInstanceByServiceIDAndInstanceID(ctx, conn, d.Get("service_id").(string), d.Get("instance_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Service Discovery Instance (%s): %s", d.Id(), err)
	}

	attributes := instance.Attributes
	// https://docs.aws.amazon.com/cloud-map/latest/api/API_RegisterInstance.html#cloudmap-RegisterInstance-request-Attributes.
	// "When the AWS_EC2_INSTANCE_ID attribute is specified, then the AWS_INSTANCE_IPV4 attribute will be filled out with the primary private IPv4 address."
	if _, ok := attributes["AWS_EC2_INSTANCE_ID"]; ok {
		delete(attributes, "AWS_INSTANCE_IPV4")
	}

	d.Set("attributes", aws.StringValueMap(attributes))
	d.Set("instance_id", instance.Id)

	return nil
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn(ctx)

	err := deregisterInstance(ctx, conn, d.Get("service_id").(string), d.Get("instance_id").(string))

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-id>/<instance-id>", d.Id())
	}

	instanceID := parts[1]
	serviceID := parts[0]
	d.Set("instance_id", instanceID)
	d.Set("service_id", serviceID)
	d.SetId(instanceID)

	return []*schema.ResourceData{d}, nil
}
