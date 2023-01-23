package rds

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	clusterEndpointCreateTimeout   = 30 * time.Minute
	clusterEndpointRetryDelay      = 5 * time.Second
	ClusterEndpointRetryMinTimeout = 3 * time.Second
)

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
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validIdentifier,
			},
			"cluster_identifier": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
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
			"excluded_members": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"static_members"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
			},
			"static_members": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"excluded_members"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	clusterId := d.Get("cluster_identifier").(string)
	endpointId := d.Get("cluster_endpoint_identifier").(string)
	endpointType := d.Get("custom_endpoint_type").(string)

	createClusterEndpointInput := &rds.CreateDBClusterEndpointInput{
		DBClusterIdentifier:         aws.String(clusterId),
		DBClusterEndpointIdentifier: aws.String(endpointId),
		EndpointType:                aws.String(endpointType),
		Tags:                        Tags(tags.IgnoreAWS()),
	}

	if v := d.Get("static_members"); v != nil {
		createClusterEndpointInput.StaticMembers = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v := d.Get("excluded_members"); v != nil {
		createClusterEndpointInput.ExcludedMembers = flex.ExpandStringSet(v.(*schema.Set))
	}

	_, err := conn.CreateDBClusterEndpointWithContext(ctx, createClusterEndpointInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Cluster Endpoint: %s", err)
	}

	d.SetId(endpointId)

	err = resourceClusterEndpointWaitForAvailable(ctx, clusterEndpointCreateTimeout, d.Id(), conn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Cluster Endpoint: waiting for completion: %s", err)
	}

	return append(diags, resourceClusterEndpointRead(ctx, d, meta)...)
}

func resourceClusterEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &rds.DescribeDBClusterEndpointsInput{
		DBClusterEndpointIdentifier: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Describing RDS Cluster: %s", input)
	resp, err := conn.DescribeDBClusterEndpointsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing RDS Cluster Endpoints (%s): %s", d.Id(), err)
	}

	if resp == nil {
		return sdkdiag.AppendErrorf(diags, "retrieving RDS Cluster Endpoints: empty response for: %s", input)
	}

	var clusterEp *rds.DBClusterEndpoint
	for _, e := range resp.DBClusterEndpoints {
		if aws.StringValue(e.DBClusterEndpointIdentifier) == d.Id() {
			clusterEp = e
			break
		}
	}

	if clusterEp == nil {
		log.Printf("[WARN] RDS Cluster Endpoint (%s) not found", d.Id())
		d.SetId("")
		return diags
	}

	arn := clusterEp.DBClusterEndpointArn
	d.Set("cluster_endpoint_identifier", clusterEp.DBClusterEndpointIdentifier)
	d.Set("cluster_identifier", clusterEp.DBClusterIdentifier)
	d.Set("arn", arn)
	d.Set("endpoint", clusterEp.Endpoint)
	d.Set("custom_endpoint_type", clusterEp.CustomEndpointType)

	if err := d.Set("excluded_members", flex.FlattenStringList(clusterEp.ExcludedMembers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting excluded_members: %s", err)
	}

	if err := d.Set("static_members", flex.FlattenStringList(clusterEp.StaticMembers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting static_members: %s", err)
	}

	tags, err := ListTags(ctx, conn, *arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for RDS Cluster Endpoint (%s): %s", *arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceClusterEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	input := &rds.ModifyDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(d.Id()),
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Cluster Endpoint (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	if v, ok := d.GetOk("custom_endpoint_type"); ok {
		input.EndpointType = aws.String(v.(string))
	}

	if attr := d.Get("excluded_members").(*schema.Set); attr.Len() > 0 {
		input.ExcludedMembers = flex.ExpandStringSet(attr)
	} else {
		input.ExcludedMembers = make([]*string, 0)
	}

	if attr := d.Get("static_members").(*schema.Set); attr.Len() > 0 {
		input.StaticMembers = flex.ExpandStringSet(attr)
	} else {
		input.StaticMembers = make([]*string, 0)
	}

	_, err := conn.ModifyDBClusterEndpointWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "modifying RDS Cluster Endpoint: %s", err)
	}

	return append(diags, resourceClusterEndpointRead(ctx, d, meta)...)
}

func resourceClusterEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	input := &rds.DeleteDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteDBClusterEndpointWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster Endpoint (%s): %s", d.Id(), err)
	}

	if err := resourceClusterEndpointWaitForDestroy(ctx, d.Timeout(schema.TimeoutDelete), d.Id(), conn); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster Endpoint (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func resourceClusterEndpointWaitForDestroy(ctx context.Context, timeout time.Duration, id string, conn *rds.RDS) error {
	log.Printf("Waiting for RDS Cluster Endpoint %s to be deleted...", id)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"available", "deleting"},
		Target:     []string{"destroyed"},
		Refresh:    DBClusterEndpointStateRefreshFunc(ctx, conn, id),
		Timeout:    timeout,
		Delay:      clusterEndpointRetryDelay,
		MinTimeout: ClusterEndpointRetryMinTimeout,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("Error waiting for RDS Cluster Endpoint (%s) to be deleted: %v", id, err)
	}
	return nil
}

func resourceClusterEndpointWaitForAvailable(ctx context.Context, timeout time.Duration, id string, conn *rds.RDS) error {
	log.Printf("Waiting for RDS Cluster Endpoint %s to become available...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"available"},
		Refresh:    DBClusterEndpointStateRefreshFunc(ctx, conn, id),
		Timeout:    timeout,
		Delay:      clusterEndpointRetryDelay,
		MinTimeout: ClusterEndpointRetryMinTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("Error waiting for RDS Cluster Endpoint (%s) to be ready: %v", id, err)
	}
	return nil
}

func DBClusterEndpointStateRefreshFunc(ctx context.Context, conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &rds.DescribeDBClusterEndpointsOutput{}

		resp, err := conn.DescribeDBClusterEndpointsWithContext(ctx, &rds.DescribeDBClusterEndpointsInput{
			DBClusterEndpointIdentifier: aws.String(id),
		})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
				return emptyResp, "destroyed", nil
			} else if resp != nil && len(resp.DBClusterEndpoints) == 0 {
				return emptyResp, "destroyed", nil
			} else {
				return emptyResp, "", fmt.Errorf("Error on refresh: %+v", err)
			}
		}

		if resp == nil || resp.DBClusterEndpoints == nil || len(resp.DBClusterEndpoints) == 0 {
			return emptyResp, "destroyed", nil
		}

		return resp.DBClusterEndpoints[0], *resp.DBClusterEndpoints[0].Status, nil
	}
}
