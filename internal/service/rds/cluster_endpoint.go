// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_cluster_endpoint", name="Cluster Endpoint")
// @Tags(identifierAttribute="arn")
func ResourceClusterEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterEndpointCreate,
		ReadWithoutTimeout:   resourceClusterEndpointRead,
		UpdateWithoutTimeout: resourceClusterEndpointUpdate,
		DeleteWithoutTimeout: resourceClusterEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_endpoint_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"cluster_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"custom_endpoint_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"READER",
					"ANY",
				}, false),
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"excluded_members": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"static_members"},
			},
			"static_members": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"excluded_members"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const (
		clusterEndpointCreateTimeout = 30 * time.Minute
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	endpointID := d.Get("cluster_endpoint_identifier").(string)
	input := &rds.CreateDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(endpointID),
		DBClusterIdentifier:         aws.String(d.Get("cluster_identifier").(string)),
		EndpointType:                aws.String(d.Get("custom_endpoint_type").(string)),
		Tags:                        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("excluded_members"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedMembers = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("static_members"); ok && v.(*schema.Set).Len() > 0 {
		input.StaticMembers = flex.ExpandStringSet(v.(*schema.Set))
	}

	_, err := conn.CreateDBClusterEndpointWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Cluster Endpoint (%s): %s", endpointID, err)
	}

	d.SetId(endpointID)

	if _, err := waitClusterEndpointCreated(ctx, conn, d.Id(), clusterEndpointCreateTimeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterEndpointRead(ctx, d, meta)...)
}

func resourceClusterEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	clusterEp, err := FindDBClusterEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Cluster Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Endpoint (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(clusterEp.DBClusterEndpointArn)
	d.Set("arn", arn)
	d.Set("cluster_endpoint_identifier", clusterEp.DBClusterEndpointIdentifier)
	d.Set("cluster_identifier", clusterEp.DBClusterIdentifier)
	d.Set("custom_endpoint_type", clusterEp.CustomEndpointType)
	d.Set("endpoint", clusterEp.Endpoint)
	d.Set("excluded_members", aws.StringValueSlice(clusterEp.ExcludedMembers))
	d.Set("static_members", aws.StringValueSlice(clusterEp.StaticMembers))

	return diags
}

func resourceClusterEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &rds.ModifyDBClusterEndpointInput{
			DBClusterEndpointIdentifier: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("custom_endpoint_type"); ok {
			input.EndpointType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("excluded_members"); ok && v.(*schema.Set).Len() > 0 {
			input.ExcludedMembers = flex.ExpandStringSet(v.(*schema.Set))
		} else {
			input.ExcludedMembers = aws.StringSlice([]string{})
		}

		if v, ok := d.GetOk("static_members"); ok && v.(*schema.Set).Len() > 0 {
			input.StaticMembers = flex.ExpandStringSet(v.(*schema.Set))
		} else {
			input.StaticMembers = aws.StringSlice([]string{})
		}

		_, err := conn.ModifyDBClusterEndpointWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying RDS Cluster Endpoint (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterEndpointRead(ctx, d, meta)...)
}

func resourceClusterEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	log.Printf("[DEBUG] Deleting RDS Cluster Endpoint: %s", d.Id())
	_, err := conn.DeleteDBClusterEndpointWithContext(ctx, &rds.DeleteDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster Endpoint (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Endpoint (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindDBClusterEndpointByID(ctx context.Context, conn *rds.RDS, id string) (*rds.DBClusterEndpoint, error) {
	input := &rds.DescribeDBClusterEndpointsInput{
		DBClusterEndpointIdentifier: aws.String(id),
	}

	output, err := conn.DescribeDBClusterEndpointsWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DBClusterEndpoints) == 0 || output.DBClusterEndpoints[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.DBClusterEndpoints); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	dbClusterEndpoint := output.DBClusterEndpoints[0]

	// Eventual consistency check.
	if aws.StringValue(dbClusterEndpoint.DBClusterEndpointIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return dbClusterEndpoint, nil
}

func statusClusterEndpoint(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBClusterEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitClusterEndpointCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBClusterEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"available"},
		Refresh:    statusClusterEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBClusterEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitClusterEndpointDeleted(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBClusterEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"available", "deleting"},
		Target:     []string{},
		Refresh:    statusClusterEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBClusterEndpoint); ok {
		return output, err
	}

	return nil, err
}
