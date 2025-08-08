// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3outposts

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3outposts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3outposts/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3outposts_endpoint", name="Endpoint")
func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointCreate,
		ReadWithoutTimeout:   resourceEndpointRead,
		DeleteWithoutTimeout: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceEndpointImportState,
		},

		Schema: map[string]*schema.Schema{
			"access_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.EndpointAccessType](),
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

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3OutpostsClient(ctx)

	input := &s3outposts.CreateEndpointInput{
		OutpostId:       aws.String(d.Get("outpost_id").(string)),
		SecurityGroupId: aws.String(d.Get("security_group_id").(string)),
		SubnetId:        aws.String(d.Get(names.AttrSubnetID).(string)),
	}

	if v, ok := d.GetOk("access_type"); ok {
		input.AccessType = awstypes.EndpointAccessType(v.(string))
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		input.CustomerOwnedIpv4Pool = aws.String(v.(string))
	}

	output, err := conn.CreateEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Outposts Endpoint: %s", err)
	}

	d.SetId(aws.ToString(output.EndpointArn))

	if _, err := waitEndpointStatusCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Outposts Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3OutpostsClient(ctx)

	endpoint, err := findEndpointByARN(ctx, conn, d.Id())

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
		d.Set(names.AttrCreationTime, aws.ToTime(endpoint.CreationTime).Format(time.RFC3339))
	}
	d.Set("customer_owned_ipv4_pool", endpoint.CustomerOwnedIpv4Pool)
	if err := d.Set("network_interfaces", flattenNetworkInterfaces(endpoint.NetworkInterfaces)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_interfaces: %s", err)
	}
	d.Set("outpost_id", endpoint.OutpostsId)

	return diags
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3OutpostsClient(ctx)

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
	_, err = conn.DeleteEndpoint(ctx, &s3outposts.DeleteEndpointInput{
		EndpointId: aws.String(arnResourceParts[3]),
		OutpostId:  aws.String(arnResourceParts[1]),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Outposts Endpoint (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceEndpointImportState(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
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

func findEndpointByARN(ctx context.Context, conn *s3outposts.Client, arn string) (*awstypes.Endpoint, error) {
	input := &s3outposts.ListEndpointsInput{}

	output, err := findEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws.ToString(v.EndpointArn) == arn && v.Status != awstypes.EndpointStatusDeleting && v.Status != awstypes.EndpointStatusDeleteFailed {
			return &v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func findEndpoints(ctx context.Context, conn *s3outposts.Client, input *s3outposts.ListEndpointsInput) ([]awstypes.Endpoint, error) {
	var output []awstypes.Endpoint

	pages := s3outposts.NewListEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Endpoints...)
	}

	return output, nil
}

func statusEndpoint(ctx context.Context, conn *s3outposts.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findEndpointByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitEndpointStatusCreated(ctx context.Context, conn *s3outposts.Client, arn string) (*awstypes.Endpoint, error) {
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EndpointStatusPending),
		Target:  enum.Slice(awstypes.EndpointStatusAvailable),
		Refresh: statusEndpoint(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Endpoint); ok {
		if failedReason := output.FailedReason; failedReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(failedReason.ErrorCode), aws.ToString(failedReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func flattenNetworkInterfaces(apiObjects []awstypes.NetworkInterface) []any {
	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenNetworkInterface(apiObject))
	}

	return tfList
}

func flattenNetworkInterface(apiObject awstypes.NetworkInterface) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap[names.AttrNetworkInterfaceID] = aws.ToString(v)
	}

	return tfMap
}
