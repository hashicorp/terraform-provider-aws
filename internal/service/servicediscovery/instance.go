// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_service_discovery_instance", name="Instance")
func resourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstancePut,
		ReadWithoutTimeout:   resourceInstanceRead,
		UpdateWithoutTimeout: resourceInstancePut,
		DeleteWithoutTimeout: resourceInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceInstanceImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAttributes: {
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
			names.AttrInstanceID: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &servicediscovery.RegisterInstanceInput{
		Attributes:       flex.ExpandStringValueMap(d.Get(names.AttrAttributes).(map[string]interface{})),
		CreatorRequestId: aws.String(id.UniqueId()),
		InstanceId:       aws.String(instanceID),
		ServiceId:        aws.String(d.Get("service_id").(string)),
	}

	output, err := conn.RegisterInstance(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering Service Discovery Instance (%s): %s", instanceID, err)
	}

	d.SetId(instanceID)

	if output != nil && output.OperationId != nil {
		if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Service Discovery Instance (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	instance, err := findInstanceByTwoPartKey(ctx, conn, d.Get("service_id").(string), d.Get(names.AttrInstanceID).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Discovery Instance (%s): %s", d.Id(), err)
	}

	attributes := instance.Attributes
	// https://docs.aws.amazon.com/cloud-map/latest/api/API_RegisterInstance.html#cloudmap-RegisterInstance-request-Attributes.
	// "When the AWS_EC2_INSTANCE_ID attribute is specified, then the AWS_INSTANCE_IPV4 attribute will be filled out with the primary private IPv4 address."
	if _, ok := attributes["AWS_EC2_INSTANCE_ID"]; ok {
		delete(attributes, "AWS_INSTANCE_IPV4")
	}

	d.Set(names.AttrAttributes, attributes)
	d.Set(names.AttrInstanceID, instance.Id)

	return diags
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	err := deregisterInstance(ctx, conn, d.Get("service_id").(string), d.Get(names.AttrInstanceID).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-id>/<instance-id>", d.Id())
	}

	instanceID := parts[1]
	serviceID := parts[0]
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("service_id", serviceID)
	d.SetId(instanceID)

	return []*schema.ResourceData{d}, nil
}

func deregisterInstance(ctx context.Context, conn *servicediscovery.Client, serviceID, instanceID string) error {
	input := &servicediscovery.DeregisterInstanceInput{
		InstanceId: aws.String(instanceID),
		ServiceId:  aws.String(serviceID),
	}

	log.Printf("[INFO] Deregistering Service Discovery Service (%s) Instance: %s", serviceID, instanceID)
	output, err := conn.DeregisterInstance(ctx, input)

	if err != nil {
		return fmt.Errorf("deregistering Service Discovery Service (%s) Instance (%s): %w", serviceID, instanceID, err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId)); err != nil {
			return fmt.Errorf("waiting for Service Discovery Service (%s) Instance (%s) delete: %w", serviceID, instanceID, err)
		}
	}

	return nil
}

func findInstanceByTwoPartKey(ctx context.Context, conn *servicediscovery.Client, serviceID, instanceID string) (*awstypes.Instance, error) {
	input := &servicediscovery.GetInstanceInput{
		InstanceId: aws.String(instanceID),
		ServiceId:  aws.String(serviceID),
	}

	output, err := conn.GetInstance(ctx, input)

	if errs.IsA[*awstypes.InstanceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Instance == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Instance, nil
}
