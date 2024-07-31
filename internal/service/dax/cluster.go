// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dax

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dax"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dax/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dax_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Delete: schema.DefaultTimeout(45 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_endpoint_encryption_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ClusterEndpointEncryptionType](),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// API returns "NONE" by default.
					if old == string(awstypes.ClusterEndpointEncryptionTypeNone) && new == "" {
						return true
					}

					return old == new
				},
			},
			names.AttrClusterName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringMatch(regexache.MustCompile(`^[a-z]`), "must begin with a lowercase letter"),
					validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			names.AttrIAMRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"node_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"replication_factor": {
				Type:     schema.TypeInt,
				Required: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"notification_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrParameterGroupName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"server_side_encryption": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
					},
				},
			},
			"subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"configuration_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	clusterName := d.Get(names.AttrClusterName).(string)
	iamRoleArn := d.Get(names.AttrIAMRoleARN).(string)
	nodeType := d.Get("node_type").(string)
	numNodes := int32(d.Get("replication_factor").(int))
	subnetGroupName := d.Get("subnet_group_name").(string)
	securityIdSet := d.Get(names.AttrSecurityGroupIDs).(*schema.Set)
	securityIds := flex.ExpandStringSet(securityIdSet)
	input := &dax.CreateClusterInput{
		ClusterName:       aws.String(clusterName),
		IamRoleArn:        aws.String(iamRoleArn),
		NodeType:          aws.String(nodeType),
		ReplicationFactor: numNodes,
		SecurityGroupIds:  aws.ToStringSlice(securityIds),
		SubnetGroupName:   aws.String(subnetGroupName),
		Tags:              getTagsIn(ctx),
	}

	// optionals can be defaulted by AWS
	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cluster_endpoint_encryption_type"); ok {
		input.ClusterEndpointEncryptionType = awstypes.ClusterEndpointEncryptionType(v.(string))
	}

	if v, ok := d.GetOk(names.AttrParameterGroupName); ok {
		input.ParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		input.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_topic_arn"); ok {
		input.NotificationTopicArn = aws.String(v.(string))
	}

	preferredAZs := d.Get(names.AttrAvailabilityZones).(*schema.Set)
	if preferredAZs.Len() > 0 {
		input.AvailabilityZones = flex.ExpandStringValueSet(preferredAZs)
	}

	if v, ok := d.GetOk("server_side_encryption"); ok && len(v.([]interface{})) > 0 {
		options := v.([]interface{})
		s := options[0].(map[string]interface{})
		input.SSESpecification = expandEncryptAtRestOptions(s)
	}

	// IAM roles take some time to propagate
	var resp *dax.CreateClusterOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		resp, err = conn.CreateCluster(ctx, input)
		if errs.IsA[*awstypes.InvalidParameterValueException](err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.CreateCluster(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DAX cluster: %s", err)
	}

	// Assign the cluster id as the resource ID
	// DAX always retains the id in lower case, so we have to
	// mimic that or else we won't be able to refresh a resource whose
	// name contained uppercase characters.
	d.SetId(strings.ToLower(*resp.Cluster.ClusterName))

	pending := []string{"creating", "modifying"}
	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     []string{"available"},
		Refresh:    clusterStateRefreshFunc(ctx, conn, d.Id(), "available", pending),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for state to become available: %v", d.Id())
	_, sterr := stateConf.WaitForStateContext(ctx)
	if sterr != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DAX cluster (%s) to be created: %s", d.Id(), sterr)
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	req := &dax.DescribeClustersInput{
		ClusterNames: []string{d.Id()},
	}

	res, err := conn.DescribeClusters(ctx, req)

	if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		log.Printf("[WARN] DAX cluster (%s) not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DAX cluster (%s): %s", d.Id(), err)
	}

	if len(res.Clusters) == 0 {
		log.Printf("[WARN] DAX cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	c := res.Clusters[0]
	d.Set(names.AttrARN, c.ClusterArn)
	d.Set(names.AttrClusterName, c.ClusterName)
	d.Set("cluster_endpoint_encryption_type", c.ClusterEndpointEncryptionType)
	d.Set(names.AttrDescription, c.Description)
	d.Set(names.AttrIAMRoleARN, c.IamRoleArn)
	d.Set("node_type", c.NodeType)
	d.Set("replication_factor", c.TotalNodes)

	if c.ClusterDiscoveryEndpoint != nil {
		d.Set(names.AttrPort, c.ClusterDiscoveryEndpoint.Port)
		d.Set("configuration_endpoint", fmt.Sprintf("%s:%d", aws.ToString(c.ClusterDiscoveryEndpoint.Address), c.ClusterDiscoveryEndpoint.Port))
		d.Set("cluster_address", c.ClusterDiscoveryEndpoint.Address)
	}

	d.Set("subnet_group_name", c.SubnetGroup)
	d.Set(names.AttrSecurityGroupIDs, flattenSecurityGroupIDs(c.SecurityGroups))

	if c.ParameterGroup != nil {
		d.Set(names.AttrParameterGroupName, c.ParameterGroup.ParameterGroupName)
	}

	d.Set("maintenance_window", c.PreferredMaintenanceWindow)

	if c.NotificationConfiguration != nil {
		if aws.ToString(c.NotificationConfiguration.TopicStatus) == "active" {
			d.Set("notification_topic_arn", c.NotificationConfiguration.TopicArn)
		}
	}

	if err := setClusterNodeData(d, c); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DAX cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("server_side_encryption", flattenEncryptAtRestOptions(c.SSEDescription)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting server_side_encryption: %s", err)
	}

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	req := &dax.UpdateClusterInput{
		ClusterName: aws.String(d.Id()),
	}

	requestUpdate := false
	awaitUpdate := false
	if d.HasChange(names.AttrDescription) {
		req.Description = aws.String(d.Get(names.AttrDescription).(string))
		requestUpdate = true
	}

	if d.HasChange(names.AttrSecurityGroupIDs) {
		if attr := d.Get(names.AttrSecurityGroupIDs).(*schema.Set); attr.Len() > 0 {
			req.SecurityGroupIds = flex.ExpandStringValueSet(attr)
			requestUpdate = true
		}
	}

	if d.HasChange(names.AttrParameterGroupName) {
		req.ParameterGroupName = aws.String(d.Get(names.AttrParameterGroupName).(string))
		requestUpdate = true
	}

	if d.HasChange("maintenance_window") {
		req.PreferredMaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("notification_topic_arn") {
		v := d.Get("notification_topic_arn").(string)
		req.NotificationTopicArn = aws.String(v)
		if v == "" {
			inactive := "inactive"
			req.NotificationTopicStatus = &inactive
		}
		requestUpdate = true
	}

	if requestUpdate {
		log.Printf("[DEBUG] Modifying DAX Cluster (%s), opts:\n%v", d.Id(), req)
		_, err := conn.UpdateCluster(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DAX cluster (%s), error: %s", d.Id(), err)
		}
		awaitUpdate = true
	}

	if d.HasChange("replication_factor") {
		oraw, nraw := d.GetChange("replication_factor")
		o := oraw.(int)
		n := nraw.(int)
		if n < o {
			log.Printf("[INFO] Decreasing nodes in DAX cluster %s from %d to %d", d.Id(), o, n)
			_, err := conn.DecreaseReplicationFactor(ctx, &dax.DecreaseReplicationFactorInput{
				ClusterName:          aws.String(d.Id()),
				NewReplicationFactor: int32(nraw.(int)),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "increasing nodes in DAX cluster %s, error: %s", d.Id(), err)
			}
			awaitUpdate = true
		}
		if n > o {
			log.Printf("[INFO] Increasing nodes in DAX cluster %s from %d to %d", d.Id(), o, n)
			_, err := conn.IncreaseReplicationFactor(ctx, &dax.IncreaseReplicationFactorInput{
				ClusterName:          aws.String(d.Id()),
				NewReplicationFactor: int32(nraw.(int)),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "increasing nodes in DAX cluster %s, error: %s", d.Id(), err)
			}
			awaitUpdate = true
		}
	}

	if awaitUpdate {
		log.Printf("[DEBUG] Waiting for update: %s", d.Id())
		pending := []string{"modifying"}
		stateConf := &retry.StateChangeConf{
			Pending:    pending,
			Target:     []string{"available"},
			Refresh:    clusterStateRefreshFunc(ctx, conn, d.Id(), "available", pending),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
			Delay:      30 * time.Second,
		}

		_, sterr := stateConf.WaitForStateContext(ctx)
		if sterr != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DAX (%s) to update: %s", d.Id(), sterr)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func setClusterNodeData(d *schema.ResourceData, c awstypes.Cluster) error {
	sortedNodes := make([]awstypes.Node, len(c.Nodes))
	copy(sortedNodes, c.Nodes)
	sort.Sort(byNodeId(sortedNodes))

	nodeData := make([]map[string]interface{}, 0, len(sortedNodes))

	for _, node := range sortedNodes {
		nodeData = append(nodeData, map[string]interface{}{
			names.AttrID:               aws.ToString(node.NodeId),
			names.AttrAddress:          aws.ToString(node.Endpoint.Address),
			names.AttrPort:             node.Endpoint.Port,
			names.AttrAvailabilityZone: aws.ToString(node.AvailabilityZone),
		})
	}

	return d.Set("nodes", nodeData)
}

type byNodeId []awstypes.Node

func (b byNodeId) Len() int      { return len(b) }
func (b byNodeId) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byNodeId) Less(i, j int) bool {
	return b[i].NodeId != nil && b[j].NodeId != nil &&
		aws.ToString(b[i].NodeId) < aws.ToString(b[j].NodeId)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DAXClient(ctx)

	req := &dax.DeleteClusterInput{
		ClusterName: aws.String(d.Id()),
	}
	err := retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteCluster(ctx, req)
		if errs.IsA[*awstypes.InvalidClusterStateFault](err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteCluster(ctx, req)
	}

	if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DAX cluster: %s", err)
	}

	log.Printf("[DEBUG] Waiting for deletion: %v", d.Id())
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"creating", "available", "deleting", "incompatible-parameters", "incompatible-network"},
		Target:     []string{},
		Refresh:    clusterStateRefreshFunc(ctx, conn, d.Id(), "", []string{}),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, sterr := stateConf.WaitForStateContext(ctx)
	if sterr != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DAX (%s) to delete: %s", d.Id(), sterr)
	}

	return diags
}

func clusterStateRefreshFunc(ctx context.Context, conn *dax.Client, clusterID, givenState string, pending []string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeClusters(ctx, &dax.DescribeClustersInput{
			ClusterNames: []string{clusterID},
		})

		if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
			log.Printf("[DEBUG] Detect deletion")
			return nil, "", nil
		}

		if err != nil {
			log.Printf("[ERROR] clusterStateRefreshFunc: %s", err)
			return nil, "", err
		}

		if len(resp.Clusters) == 0 {
			return nil, "", fmt.Errorf("Error: no DAX clusters found for id (%s)", clusterID)
		}

		var c awstypes.Cluster
		for _, cluster := range resp.Clusters {
			if aws.ToString(cluster.ClusterName) == clusterID {
				log.Printf("[DEBUG] Found matching DAX cluster: %s", *cluster.ClusterName)
				c = cluster
			}
		}

		if reflect.ValueOf(c).IsZero() {
			return nil, "", fmt.Errorf("Error: no matching DAX cluster for id (%s)", clusterID)
		}

		// DescribeCluster returns a response without status late on in the
		// deletion process - assume cluster is still deleting until we
		// get ClusterNotFoundFault
		if c.Status == nil {
			log.Printf("[DEBUG] DAX Cluster %s has no status attribute set - assume status is deleting", clusterID)
			return c, "deleting", nil
		}

		log.Printf("[DEBUG] DAX Cluster (%s) status: %v", clusterID, *c.Status)

		// return the current state if it's in the pending array
		for _, p := range pending {
			log.Printf("[DEBUG] DAX: checking pending state (%s) for cluster (%s), cluster status: %s", pending, clusterID, *c.Status)
			s := aws.ToString(c.Status)
			if p == s {
				log.Printf("[DEBUG] Return with status: %v", *c.Status)
				return c, p, nil
			}
		}

		// return given state if it's not in pending
		if givenState != "" {
			log.Printf("[DEBUG] DAX: checking given state (%s) of cluster (%s) against cluster status (%s)", givenState, clusterID, *c.Status)
			// check to make sure we have the node count we're expecting
			if int32(len(c.Nodes)) != aws.ToInt32(c.TotalNodes) {
				log.Printf("[DEBUG] Node count is not what is expected: %d found, %d expected", len(c.Nodes), *c.TotalNodes)
				return nil, "creating", nil
			}

			log.Printf("[DEBUG] Node count matched (%d)", len(c.Nodes))
			// loop the nodes and check their status as well
			for _, n := range c.Nodes {
				log.Printf("[DEBUG] Checking cache node for status: %v", n)
				if n.NodeStatus != nil && aws.ToString(n.NodeStatus) != "available" {
					log.Printf("[DEBUG] Node (%s) is not yet available, status: %s", *n.NodeId, *n.NodeStatus)
					return nil, "creating", nil
				}
				log.Printf("[DEBUG] Cache node not in expected state")
			}
			log.Printf("[DEBUG] DAX returning given state (%s), cluster: %v", givenState, c)
			return c, givenState, nil
		}
		log.Printf("[DEBUG] current status: %v", *c.Status)
		return c, *c.Status, nil
	}
}
