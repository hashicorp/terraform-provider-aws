// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_cluster_endpoint", name="Cluster Endpoint")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceClusterEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterEndpointCreate,
		ReadWithoutTimeout:   resourceClusterEndpointRead,
		UpdateWithoutTimeout: resourceClusterEndpointUpdate,
		DeleteWithoutTimeout: resourceClusterEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_endpoint_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			names.AttrClusterIdentifier: {
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
			names.AttrEndpoint: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	endpointID := d.Get("cluster_endpoint_identifier").(string)
	input := &rds.CreateDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(endpointID),
		DBClusterIdentifier:         aws.String(d.Get(names.AttrClusterIdentifier).(string)),
		EndpointType:                aws.String(d.Get("custom_endpoint_type").(string)),
		Tags:                        getTagsInV2(ctx),
	}

	if v, ok := d.GetOk("excluded_members"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedMembers = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("static_members"); ok && v.(*schema.Set).Len() > 0 {
		input.StaticMembers = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	_, err := conn.CreateDBClusterEndpoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Cluster Endpoint (%s): %s", endpointID, err)
	}

	d.SetId(endpointID)

	const (
		timeout = 30 * time.Minute
	)
	if _, err := waitClusterEndpointCreated(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Cluster Endpoint (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterEndpointRead(ctx, d, meta)...)
}

func resourceClusterEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	clusterEp, err := findDBClusterEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Cluster Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Endpoint (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, clusterEp.DBClusterEndpointArn)
	d.Set("cluster_endpoint_identifier", clusterEp.DBClusterEndpointIdentifier)
	d.Set(names.AttrClusterIdentifier, clusterEp.DBClusterIdentifier)
	d.Set("custom_endpoint_type", clusterEp.CustomEndpointType)
	d.Set(names.AttrEndpoint, clusterEp.Endpoint)
	d.Set("excluded_members", clusterEp.ExcludedMembers)
	d.Set("static_members", clusterEp.StaticMembers)

	return diags
}

func resourceClusterEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &rds.ModifyDBClusterEndpointInput{
			DBClusterEndpointIdentifier: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("custom_endpoint_type"); ok {
			input.EndpointType = aws.String(v.(string))
		}

		if v, ok := d.GetOk("excluded_members"); ok && v.(*schema.Set).Len() > 0 {
			input.ExcludedMembers = flex.ExpandStringValueSet(v.(*schema.Set))
		} else {
			input.ExcludedMembers = []string{}
		}

		if v, ok := d.GetOk("static_members"); ok && v.(*schema.Set).Len() > 0 {
			input.StaticMembers = flex.ExpandStringValueSet(v.(*schema.Set))
		} else {
			input.StaticMembers = []string{}
		}

		_, err := conn.ModifyDBClusterEndpoint(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying RDS Cluster Endpoint (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterEndpointRead(ctx, d, meta)...)
}

func resourceClusterEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting RDS Cluster Endpoint: %s", d.Id())
	_, err := conn.DeleteDBClusterEndpoint(ctx, &rds.DeleteDBClusterEndpointInput{
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

func findDBClusterEndpointByID(ctx context.Context, conn *rds.Client, id string) (*types.DBClusterEndpoint, error) {
	input := &rds.DescribeDBClusterEndpointsInput{
		DBClusterEndpointIdentifier: aws.String(id),
	}
	output, err := findDBClusterEndpoint(ctx, conn, input, tfslices.PredicateTrue[*types.DBClusterEndpoint]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBClusterEndpointIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBClusterEndpoint(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClusterEndpointsInput, filter tfslices.Predicate[*types.DBClusterEndpoint]) (*types.DBClusterEndpoint, error) {
	output, err := findDBClusterEndpoints(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClusterEndpoints(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClusterEndpointsInput, filter tfslices.Predicate[*types.DBClusterEndpoint]) ([]types.DBClusterEndpoint, error) {
	var output []types.DBClusterEndpoint

	pages := rds.NewDescribeDBClusterEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusterEndpoints {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusClusterEndpoint(ctx context.Context, conn *rds.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBClusterEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitClusterEndpointCreated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.DBClusterEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterEndpointStatusCreating},
		Target:     []string{clusterEndpointStatusAvailable},
		Refresh:    statusClusterEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBClusterEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitClusterEndpointDeleted(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.DBClusterEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterEndpointStatusAvailable, clusterEndpointStatusDeleting},
		Target:     []string{},
		Refresh:    statusClusterEndpoint(ctx, conn, id),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBClusterEndpoint); ok {
		return output, err
	}

	return nil, err
}
