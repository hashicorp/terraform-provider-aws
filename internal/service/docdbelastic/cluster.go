package docdbelastic

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdbelastic"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdbelastic/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_docdbelastic_cluster")
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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(45 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"admin_user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"admin_user_password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Required:  true,
			},
			"auth_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(AuthValues(), false),
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validIdentifierPrefix,
				ForceNew:     true,
			},
			"shard_capacity": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: validation.IntInSlice([]int{
					2, 4, 8, 16, 32, 64,
				}),
			},
			"shard_count": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 32),
			},
			"client_token": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"preferred_maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(val interface{}) string {
					if val == nil {
						return ""
					}
					return strings.ToLower(val.(string))
				},
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"cluster_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameCluster = "Cluster"
)

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBElasticClient(ctx)

	in := &docdbelastic.CreateClusterInput{
		AdminUserName:     aws.String(d.Get("admin_user_name").(string)),
		AdminUserPassword: aws.String(d.Get("admin_user_password").(string)),
		AuthType:          awstypes.Auth(d.Get("auth_type").(string)),
		ClusterName:       aws.String(d.Get("cluster_name").(string)),
		ShardCapacity:     aws.Int32(int32(d.Get("shard_capacity").(int))),
		ShardCount:        aws.Int32(int32(d.Get("shard_count").(int))),
	}

	if v, ok := d.GetOk("clientToken"); ok {
		in.ClientToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		in.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_maintenance_window"); ok {
		in.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if attr := d.Get("subnet_ids").(*schema.Set); attr.Len() > 0 {
		in.SubnetIds = flex.ExpandStringValueSet(attr)
	}

	if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
		in.VpcSecurityGroupIds = flex.ExpandStringValueSet(attr)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateCluster(ctx, in)
	if err != nil {
		return create.DiagError(names.DocDBElastic, create.ErrActionCreating, ResNameCluster, d.Get("cluster_name").(string), err)
	}

	if out == nil || out.Cluster == nil {
		return create.DiagError(names.DocDBElastic, create.ErrActionCreating, ResNameCluster, d.Get("cluster_name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Cluster.ClusterArn))

	if _, err := waitClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.DocDBElastic, create.ErrActionWaitingForCreation, ResNameCluster, d.Id(), err)
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DocDBElasticClient(ctx)

	out, err := findClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DocDBElastic Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.DocDBElastic, create.ErrActionReading, ResNameCluster, d.Id(), err)
	}

	d.Set("cluster_arn", out.ClusterArn)
	d.Set("admin_user_name", out.AdminUserName)
	d.Set("auth_type", out.AuthType)

	d.Set("cluster_endpoint", out.ClusterEndpoint)
	d.Set("cluster_name", out.ClusterName)
	d.Set("kms_key_id", out.KmsKeyId)
	d.Set("preferred_maintenance_window", out.PreferredMaintenanceWindow)
	d.Set("shard_capacity", out.ShardCapacity)
	d.Set("shard_count", out.ShardCount)

	if err := d.Set("subnet_ids", out.SubnetIds); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids: %s", err)
	}

	if err := d.Set("vpc_security_group_ids", out.VpcSecurityGroupIds); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_security_group_ids: %s", err)
	}

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBElasticClient(ctx)

	update := false

	in := &docdbelastic.UpdateClusterInput{
		ClusterArn: aws.String(d.Id()),
	}

	if d.HasChanges("admin_user_password") {
		in.AdminUserPassword = aws.String(d.Get("admin_user_password").(string))
		update = true
	}
	if d.HasChanges("auth_type") {
		in.AuthType = awstypes.Auth(d.Get("auth_type").(string))
		update = true
	}
	if d.HasChanges("client_token") {
		in.ClientToken = aws.String(d.Get("client_token").(string))
		update = true
	}
	if d.HasChanges("preferred_maintenance_window") {
		in.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		update = true
	}
	if d.HasChanges("shard_capacity") {
		in.ShardCapacity = aws.Int32(int32(d.Get("shard_capacity").(int)))
		update = true
	}
	if d.HasChanges("shard_count") {
		in.ShardCount = aws.Int32(int32(d.Get("shard_count").(int)))
		update = true
	}
	if d.HasChanges("subnet_ids") {
		if attr := d.Get("subnet_ids").(*schema.Set); attr.Len() > 0 {
			in.SubnetIds = flex.ExpandStringValueSet(attr)
		} else {
			in.SubnetIds = []string{}
		}
		update = true
	}
	if d.HasChanges("vpc_security_group_ids") {
		if attr := d.Get("vpc_security_group_ids").(*schema.Set); attr.Len() > 0 {
			in.VpcSecurityGroupIds = flex.ExpandStringValueSet(attr)
		} else {
			in.VpcSecurityGroupIds = []string{}
		}
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating DocDBElastic Cluster (%s): %#v", d.Id(), in)
	out, err := conn.UpdateCluster(ctx, in)
	if err != nil {
		return create.DiagError(names.DocDBElastic, create.ErrActionUpdating, ResNameCluster, d.Id(), err)
	}

	if _, err := waitClusterUpdated(ctx, conn, aws.ToString(out.Cluster.ClusterArn), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.DocDBElastic, create.ErrActionWaitingForUpdate, ResNameCluster, d.Id(), err)
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBElasticClient(ctx)

	log.Printf("[INFO] Deleting DocDBElastic Cluster %s", d.Id())

	_, err := conn.DeleteCluster(ctx, &docdbelastic.DeleteClusterInput{
		ClusterArn: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.DocDBElastic, create.ErrActionDeleting, ResNameCluster, d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.DocDBElastic, create.ErrActionWaitingForDeletion, ResNameCluster, d.Id(), err)
	}

	return nil
}

const (
	statusCreatePending = "CREATING"
	statusChangePending = "UPDATING"
	statusDeleting      = "DELETING"
	statusNormal        = "ACTIVE"
)

func waitClusterCreated(ctx context.Context, conn *docdbelastic.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusCreatePending},
		Target:                    []string{statusNormal},
		Refresh:                   statusCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Cluster); ok {
		return out, err
	}

	return nil, err
}

func waitClusterUpdated(ctx context.Context, conn *docdbelastic.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusNormal},
		Refresh:                   statusCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Cluster); ok {
		return out, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *docdbelastic.Client, id string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Cluster); ok {
		return out, err
	}

	return nil, err
}

func statusCluster(ctx context.Context, conn *docdbelastic.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findClusterByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		status := string(out.Status)
		return out, aws.ToString(&status), nil
	}
}

func findClusterByID(ctx context.Context, conn *docdbelastic.Client, id string) (*awstypes.Cluster, error) {
	in := &docdbelastic.GetClusterInput{
		ClusterArn: aws.String(id),
	}
	out, err := conn.GetCluster(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Cluster == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Cluster, nil
}

func AuthValues() []string {
	authValues := awstypes.Auth("").Values()
	stringValues := make([]string, len(authValues))
	for i, value := range authValues {
		stringValues[i] = string(value)
	}

	return stringValues
}
