// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticsearch_vpc_endpoint", name="VPC Endpoint")
func resourceVPCEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointCreate,
		ReadWithoutTimeout:   resourceVPCEndpointRead,
		UpdateWithoutTimeout: resourceVPCEndpointUpdate,
		DeleteWithoutTimeout: resourceVPCEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"domain_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_options": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZones: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceVPCEndpointCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	input := &elasticsearchservice.CreateVpcEndpointInput{
		DomainArn:  aws.String(d.Get("domain_arn").(string)),
		VpcOptions: expandVPCOptions(d.Get("vpc_options").([]any)[0].(map[string]any)),
	}

	output, err := conn.CreateVpcEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elasticsearch VPC Endpoint: %s", err)
	}

	d.SetId(aws.ToString(output.VpcEndpoint.VpcEndpointId))

	if _, err := waitVPCEndpointCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch VPC Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCEndpointRead(ctx, d, meta)...)
}

func resourceVPCEndpointRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	endpoint, err := findVPCEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch VPC Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch VPC Endpoint (%s): %s", d.Id(), err)
	}

	d.Set("domain_arn", endpoint.DomainArn)
	d.Set(names.AttrEndpoint, endpoint.Endpoint)
	if endpoint.VpcOptions != nil {
		if err := d.Set("vpc_options", []any{flattenVPCDerivedInfo(endpoint.VpcOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_options: %s", err)
		}
	} else {
		d.Set("vpc_options", nil)
	}

	return diags
}

func resourceVPCEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	input := &elasticsearchservice.UpdateVpcEndpointInput{
		VpcOptions:    expandVPCOptions(d.Get("vpc_options").([]any)[0].(map[string]any)),
		VpcEndpointId: aws.String(d.Id()),
	}

	_, err := conn.UpdateVpcEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Elasticsearch VPC Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCEndpointUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch VPC Endpoint (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceVPCEndpointRead(ctx, d, meta)...)
}

func resourceVPCEndpointDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	log.Printf("[DEBUG] Deleting Elasticsearch VPC Endpoint: %s", d.Id())
	_, err := conn.DeleteVpcEndpoint(ctx, &elasticsearchservice.DeleteVpcEndpointInput{
		VpcEndpointId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elasticsearch VPC Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch VPC Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

type vpcEndpointNotFoundError struct {
	apiError error
}

func (e *vpcEndpointNotFoundError) Error() string {
	if e.apiError != nil {
		return e.apiError.Error()
	}

	return "VPC endpoint not found"
}

func (e *vpcEndpointNotFoundError) Is(err error) bool {
	_, ok := err.(*vpcEndpointNotFoundError)
	return ok
}

func (e *vpcEndpointNotFoundError) As(target any) bool {
	t, ok := target.(**retry.NotFoundError)
	if !ok {
		return false
	}

	*t = &retry.NotFoundError{
		Message: e.Error(),
	}

	return true
}

func vpcEndpointError(apiObject awstypes.VpcEndpointError) error {
	errorCode := string(apiObject.ErrorCode)
	innerError := fmt.Errorf("%s: %s", errorCode, aws.ToString(apiObject.ErrorMessage))
	err := fmt.Errorf("%s: %w", aws.ToString(apiObject.VpcEndpointId), innerError)

	if errorCode == string(awstypes.VpcEndpointErrorCodeEndpointNotFound) {
		err = &vpcEndpointNotFoundError{apiError: err}
	}

	return err
}

func vpcEndpointsError(apiObjects []awstypes.VpcEndpointError) error {
	var errs []error

	for _, apiObject := range apiObjects {
		errs = append(errs, vpcEndpointError(apiObject))
	}

	return errors.Join(errs...)
}

func findVPCEndpointByID(ctx context.Context, conn *elasticsearchservice.Client, id string) (*awstypes.VpcEndpoint, error) {
	input := &elasticsearchservice.DescribeVpcEndpointsInput{
		VpcEndpointIds: []string{id},
	}

	return findVPCEndpoint(ctx, conn, input)
}

func findVPCEndpoint(ctx context.Context, conn *elasticsearchservice.Client, input *elasticsearchservice.DescribeVpcEndpointsInput) (*awstypes.VpcEndpoint, error) {
	output, err := findVPCEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpoints(ctx context.Context, conn *elasticsearchservice.Client, input *elasticsearchservice.DescribeVpcEndpointsInput) ([]awstypes.VpcEndpoint, error) {
	output, err := conn.DescribeVpcEndpoints(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if errs := output.VpcEndpointErrors; len(errs) > 0 {
		return nil, vpcEndpointsError(errs)
	}

	return output.VpcEndpoints, nil
}

func statusVPCEndpoint(ctx context.Context, conn *elasticsearchservice.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findVPCEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitVPCEndpointCreated(ctx context.Context, conn *elasticsearchservice.Client, id string, timeout time.Duration) (*awstypes.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpcEndpointStatusCreating),
		Target:  enum.Slice(awstypes.VpcEndpointStatusActive),
		Refresh: statusVPCEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointUpdated(ctx context.Context, conn *elasticsearchservice.Client, id string, timeout time.Duration) (*awstypes.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpcEndpointStatusUpdating),
		Target:  enum.Slice(awstypes.VpcEndpointStatusActive),
		Refresh: statusVPCEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointDeleted(ctx context.Context, conn *elasticsearchservice.Client, id string, timeout time.Duration) (*awstypes.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpcEndpointStatusDeleting),
		Target:  []string{},
		Refresh: statusVPCEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcEndpoint); ok {
		return output, err
	}

	return nil, err
}
