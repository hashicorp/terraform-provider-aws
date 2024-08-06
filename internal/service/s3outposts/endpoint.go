// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3outposts

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3outposts_endpoint")
func ResourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointCreate,
		ReadWithoutTimeout:   resourceEndpointRead,
		DeleteWithoutTimeout: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceEndpointImportState,
		},

		Schema: map[string]*schema.Schema{
			"access_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(s3outposts.EndpointAccessType_Values(), false),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCIDRBlock: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_owned_ipv4_pool": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"network_interfaces": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"outpost_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"security_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			names.AttrSubnetID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3OutpostsConn(ctx)

	input := &s3outposts.CreateEndpointInput{
		OutpostId:       aws.String(d.Get("outpost_id").(string)),
		SecurityGroupId: aws.String(d.Get("security_group_id").(string)),
		SubnetId:        aws.String(d.Get(names.AttrSubnetID).(string)),
	}

	if v, ok := d.GetOk("access_type"); ok {
		input.AccessType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		input.CustomerOwnedIpv4Pool = aws.String(v.(string))
	}

	output, err := conn.CreateEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Outposts Endpoint: %s", err)
	}

	d.SetId(aws.StringValue(output.EndpointArn))

	if _, err := waitEndpointStatusCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Outposts Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3OutpostsConn(ctx)

	endpoint, err := FindEndpointByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Outposts Endpoint %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Outposts Endpoint (%s): %s", d.Id(), err)
	}

	d.Set("access_type", endpoint.AccessType)
	d.Set(names.AttrARN, endpoint.EndpointArn)
	d.Set(names.AttrCIDRBlock, endpoint.CidrBlock)
	if endpoint.CreationTime != nil {
		d.Set(names.AttrCreationTime, aws.TimeValue(endpoint.CreationTime).Format(time.RFC3339))
	}
	d.Set("customer_owned_ipv4_pool", endpoint.CustomerOwnedIpv4Pool)
	if err := d.Set("network_interfaces", flattenNetworkInterfaces(endpoint.NetworkInterfaces)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_interfaces: %s", err)
	}
	d.Set("outpost_id", endpoint.OutpostsId)

	return diags
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3OutpostsConn(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// ARN resource format: outpost/<outpost-id>/endpoint/<endpoint-id>
	arnResourceParts := strings.Split(parsedArn.Resource, "/")

	if parsedArn.AccountID == "" || len(arnResourceParts) != 4 {
		return sdkdiag.AppendErrorf(diags, "parsing S3 Outposts Endpoint ARN (%s): unknown format", d.Id())
	}

	log.Printf("[DEBUG] Deleting S3 Outposts Endpoint: %s", d.Id())
	_, err = conn.DeleteEndpointWithContext(ctx, &s3outposts.DeleteEndpointInput{
		EndpointId: aws.String(arnResourceParts[3]),
		OutpostId:  aws.String(arnResourceParts[1]),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Outposts Endpoint (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceEndpointImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%s), expected ENDPOINT-ARN,SECURITY-GROUP-ID,SUBNET-ID", d.Id())
	}

	endpointArn := idParts[0]
	securityGroupId := idParts[1]
	subnetId := idParts[2]

	d.SetId(endpointArn)
	d.Set("security_group_id", securityGroupId)
	d.Set(names.AttrSubnetID, subnetId)

	return []*schema.ResourceData{d}, nil
}

func FindEndpointByARN(ctx context.Context, conn *s3outposts.S3Outposts, arn string) (*s3outposts.Endpoint, error) {
	input := &s3outposts.ListEndpointsInput{}

	output, err := findEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws.StringValue(v.EndpointArn) == arn && aws.StringValue(v.Status) != s3outposts.EndpointStatusDeleting && aws.StringValue(v.Status) != s3outposts.EndpointStatusDeleteFailed {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findEndpoints(ctx context.Context, conn *s3outposts.S3Outposts, input *s3outposts.ListEndpointsInput) ([]*s3outposts.Endpoint, error) {
	var output []*s3outposts.Endpoint

	err := conn.ListEndpointsPagesWithContext(ctx, input, func(page *s3outposts.ListEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Endpoints {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusEndpoint(ctx context.Context, conn *s3outposts.S3Outposts, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEndpointByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitEndpointStatusCreated(ctx context.Context, conn *s3outposts.S3Outposts, arn string) (*s3outposts.Endpoint, error) {
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{s3outposts.EndpointStatusPending},
		Target:  []string{s3outposts.EndpointStatusAvailable},
		Refresh: statusEndpoint(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3outposts.Endpoint); ok {
		if failedReason := output.FailedReason; failedReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(failedReason.ErrorCode), aws.StringValue(failedReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func flattenNetworkInterfaces(apiObjects []*s3outposts.NetworkInterface) []interface{} {
	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkInterface(apiObject))
	}

	return tfList
}

func flattenNetworkInterface(apiObject *s3outposts.NetworkInterface) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap[names.AttrNetworkInterfaceID] = aws.StringValue(v)
	}

	return tfMap
}
