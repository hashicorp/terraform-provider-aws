// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_opensearch_vpc_endpoint")
func ResourceVPCEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointCreate,
		ReadWithoutTimeout:   resourceVPCEndpointRead,
		UpdateWithoutTimeout: resourceVPCEndpointPut,
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
			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"vpc_options": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceVPCEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	input := &opensearchservice.CreateVpcEndpointInput{
		DomainArn:  aws.String(d.Get("domain_arn").(string)),
		VpcOptions: expandVPCOptions(d.Get("vpc_options").([]interface{})[0].(map[string]interface{})),
	}

	output, err := conn.CreateVpcEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpenSearch VPC Endpoint: %s", err)
	}

	d.SetId(aws.StringValue(output.VpcEndpoint.VpcEndpointId))

	if err := vpcEndpointWaitUntilActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch VPC Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCEndpointRead(ctx, d, meta)...)
}

func resourceVPCEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	endpointRaw, status, err := vpcEndpointRefreshState(ctx, conn, d.Id())()

	if err != nil {
		return diag.Errorf("reading vpc endpoint: %s", err)
	}

	endpoint := endpointRaw.(*opensearchservice.VpcEndpoint)
	log.Printf("[DEBUG] vpc endpoint response: %#v", endpoint)

	d.Set("connection_status", status)
	d.Set("domain_arn", endpoint.DomainArn)

	if endpoint.VpcOptions == nil {
		return diag.Errorf("reading vpc endpoint vpc options ")
	}

	d.Set("vpc_options", flattenVPCDerivedInfo(endpoint.VpcOptions))
	return nil
}

func resourceVPCEndpointPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	// Update the VPC Endpoint
	input := &opensearchservice.UpdateVpcEndpointInput{
		VpcEndpointId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("vpc_options"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return sdkdiag.AppendErrorf(diags, "At least one field is expected inside vpc_options")
		}

		s := options[0].(map[string]interface{})
		input.VpcOptions = expandVPCOptions(s)
	}

	log.Printf("[DEBUG] Updating vpc endpoint %s", input)

	_, err := conn.UpdateVpcEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating vpc endpoint (%s): %s", d.Id(), err)
	}

	err = vpcEndpointWaitUntilUpdate(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.Errorf("waiting for vpc endpoint to become active: %s", err)
	}

	return append(diags, resourceVPCEndpointRead(ctx, d, meta)...)
}

func resourceVPCEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	req := &opensearchservice.DeleteVpcEndpointInput{
		VpcEndpointId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVpcEndpointWithContext(ctx, req)

	if tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting vpc endpoint (%s): %s", d.Id(), err)
	}

	return nil
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
	_, ok := err.(*vpcEndpointNotFoundError) //nolint:errorlint // Explicitly does *not* match down the error tree
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

func vpcEndpointError(apiObject *opensearchservice.VpcEndpointError) error {
	if apiObject == nil {
		return nil
	}

	errorCode := aws.StringValue(apiObject.ErrorCode)
	innerError := fmt.Errorf("%s: %s", errorCode, aws.StringValue(apiObject.ErrorMessage))
	err := fmt.Errorf("%s: %w", aws.StringValue(apiObject.VpcEndpointId), innerError)

	if errorCode == opensearchservice.VpcEndpointErrorCodeEndpointNotFound {
		err = &vpcEndpointNotFoundError{apiError: err}
	}

	return err
}

func vpcEndpointsError(apiObjects []*opensearchservice.VpcEndpointError) error {
	var errs []error

	for _, apiObject := range apiObjects {
		errs = append(errs, vpcEndpointError(apiObject))
	}

	return errors.Join(errs...)
}

func findVPCEndpoints(ctx context.Context, conn *opensearchservice.OpenSearchService, input *opensearchservice.DescribeVpcEndpointsInput) ([]*opensearchservice.VpcEndpoint, error) {
	output, err := conn.DescribeVpcEndpointsWithContext(ctx, input)

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

func vpcEndpointRefreshState(ctx context.Context, conn *opensearchservice.OpenSearchService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeVpcEndpointsWithContext(ctx, &opensearchservice.DescribeVpcEndpointsInput{
			VpcEndpointIds: []*string{aws.String(id)},
		})
		if err != nil {
			return nil, "", err
		}

		if resp == nil || resp.VpcEndpoints == nil ||
			len(resp.VpcEndpoints) == 0 || resp.VpcEndpoints[0] == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our connection yet. Return an empty state.
			return nil, "", nil
		}
		endpoint := resp.VpcEndpoints[0]
		if endpoint.Status == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our connection yet. Return an empty state.
			return nil, "", nil
		}
		statusCode := aws.StringValue(endpoint.Status)

		return endpoint, statusCode, nil
	}
}

func vpcEndpointWaitUntilActive(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for VPC Endpoint (%s) to become available.", id)
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			opensearchservice.VpcEndpointStatusCreating,
		},
		Target: []string{
			opensearchservice.VpcEndpointStatusActive,
		},
		Refresh: vpcEndpointRefreshState(ctx, conn, id),
		Timeout: timeout,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for VPC Endpoint (%s) to become available: %s", id, err)
	}
	return nil
}

func vpcEndpointWaitUntilUpdate(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for VPC Endpoint (%s) to become available.", id)
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			opensearchservice.VpcEndpointStatusUpdating,
		},
		Target: []string{
			opensearchservice.VpcEndpointStatusActive,
		},
		Refresh: vpcEndpointRefreshState(ctx, conn, id),
		Timeout: timeout,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for VPC Endpoint (%s) to become available: %s", id, err)
	}
	return nil
}
