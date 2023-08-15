// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
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
				Optional: true,
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
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Optional: true,
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

	// Create the VPC Endpoint
	input := &opensearchservice.CreateVpcEndpointInput{
		DomainArn: aws.String(d.Get("domain_arn").(string)),
	}

	if v, ok := d.GetOk("vpc_options"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return sdkdiag.AppendErrorf(diags, "At least one field is expected inside vpc_options")
		}

		s := options[0].(map[string]interface{})
		input.VpcOptions = expandVPCOptions(s)
	}

	log.Printf("[DEBUG] Create VPC Endpoint options: %#v", input)

	resp, err := conn.CreateVpcEndpointWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("creating vpc endpoint : %s", err)
	}

	// Get the ID and store it
	d.SetId(aws.StringValue(resp.VpcEndpoint.VpcEndpointId))
	log.Printf("[INFO] open search vpc endpoint ID: %s", d.Id())

	err = vpcEndpointWaitUntilActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.Errorf("waiting for vpc endpoint to become active: %s", err)
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
