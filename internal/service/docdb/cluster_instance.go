// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_docdb_cluster_instance", name="Cluster Instance")
// @Tags(identifierAttribute="arn")
func ResourceClusterInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterInstanceCreate,
		ReadWithoutTimeout:   resourceClusterInstanceRead,
		UpdateWithoutTimeout: resourceClusterInstanceUpdate,
		DeleteWithoutTimeout: resourceClusterInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// apply_immediately is used to determine when the update modifications take place.
			// See http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"ca_cert_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dbi_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_performance_insights": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "docdb",
				ValidateFunc: validEngine(),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
				ValidateFunc:  validIdentifier,
			},
			"identifier_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifierPrefix,
			},
			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"performance_insights_kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"preferred_backup_window": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(v interface{}) string {
					if v != nil {
						value := v.(string)
						return strings.ToLower(value)
					}
					return ""
				},
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			"promotion_tier": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 15),
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"writer": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	input := &docdb.CreateDBInstanceInput{
		DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
		DBClusterIdentifier:     aws.String(d.Get("cluster_identifier").(string)),
		Engine:                  aws.String(d.Get("engine").(string)),
		PromotionTier:           aws.Int64(int64(d.Get("promotion_tier").(int))),
		AutoMinorVersionUpgrade: aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
		Tags:                    getTagsIn(ctx),
	}

	if attr, ok := d.GetOk("availability_zone"); ok {
		input.AvailabilityZone = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("enable_performance_insights"); ok {
		input.EnablePerformanceInsights = aws.Bool(attr.(bool))
	}

	if v, ok := d.GetOk("identifier"); ok {
		input.DBInstanceIdentifier = aws.String(v.(string))
	} else {
		if v, ok := d.GetOk("identifier_prefix"); ok {
			input.DBInstanceIdentifier = aws.String(id.PrefixedUniqueId(v.(string)))
		} else {
			input.DBInstanceIdentifier = aws.String(id.PrefixedUniqueId("tf-"))
		}
	}

	if attr, ok := d.GetOk("performance_insights_kms_key_id"); ok {
		input.PerformanceInsightsKMSKeyId = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("preferred_maintenance_window"); ok {
		input.PreferredMaintenanceWindow = aws.String(attr.(string))
	}

	var resp *docdb.CreateDBInstanceOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		resp, err = conn.CreateDBInstanceWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.CreateDBInstanceWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster Instance: %s", err)
	}

	d.SetId(aws.StringValue(resp.DBInstance.DBInstanceIdentifier))

	// reuse db_instance refresh func
	stateConf := &retry.StateChangeConf{
		Pending:    resourceClusterInstanceCreateUpdatePendingStates,
		Target:     []string{"available"},
		Refresh:    resourceInstanceStateRefreshFunc(ctx, conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster Instance (%s) to become available: %s", d.Id(), err)
	}

	return append(diags, resourceClusterInstanceRead(ctx, d, meta)...)
}

func resourceClusterInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	db, err := resourceInstanceRetrieve(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DocumentDB Cluster Instance (%s): not found, removing from state.", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "retrieving DocumentDB Cluster Instance (%s): %s", d.Id(), err)
	}

	// Retrieve DB Cluster information, to determine if this Instance is a writer
	resp, err := conn.DescribeDBClustersWithContext(ctx, &docdb.DescribeDBClustersInput{
		DBClusterIdentifier: db.DBClusterIdentifier,
	})

	var dbc *docdb.DBCluster
	for _, c := range resp.DBClusters {
		if aws.StringValue(c.DBClusterIdentifier) == aws.StringValue(db.DBClusterIdentifier) {
			dbc = c
		}
	}

	if dbc == nil {
		return sdkdiag.AppendErrorf(diags, "finding DocumentDB Cluster (%s) for Cluster Instance (%s): %s",
			aws.StringValue(db.DBClusterIdentifier), aws.StringValue(db.DBInstanceIdentifier), err)
	}

	for _, m := range dbc.DBClusterMembers {
		if aws.StringValue(db.DBInstanceIdentifier) == aws.StringValue(m.DBInstanceIdentifier) {
			if *m.IsClusterWriter {
				d.Set("writer", true)
			} else {
				d.Set("writer", false)
			}
		}
	}

	if db.Endpoint != nil {
		d.Set("endpoint", db.Endpoint.Address)
		d.Set("port", db.Endpoint.Port)
	}

	if db.DBSubnetGroup != nil {
		d.Set("db_subnet_group_name", db.DBSubnetGroup.DBSubnetGroupName)
	}

	d.Set("arn", db.DBInstanceArn)
	d.Set("auto_minor_version_upgrade", db.AutoMinorVersionUpgrade)
	d.Set("availability_zone", db.AvailabilityZone)
	d.Set("cluster_identifier", db.DBClusterIdentifier)
	d.Set("dbi_resource_id", db.DbiResourceId)
	// The AWS API does not expose 'EnablePerformanceInsights' the line below should be uncommented
	// as soon as it is available in the DescribeDBClusters output.
	//d.Set("enable_performance_insights", db.EnablePerformanceInsights)
	d.Set("engine_version", db.EngineVersion)
	d.Set("engine", db.Engine)
	d.Set("identifier", db.DBInstanceIdentifier)
	d.Set("instance_class", db.DBInstanceClass)
	d.Set("kms_key_id", db.KmsKeyId)
	// The AWS API does not expose 'PerformanceInsightsKMSKeyId'  the line below should be uncommented
	// as soon as it is available in the DescribeDBClusters output.
	//d.Set("performance_insights_kms_key_id", db.PerformanceInsightsKMSKeyId)
	d.Set("preferred_backup_window", db.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", db.PreferredMaintenanceWindow)
	d.Set("promotion_tier", db.PromotionTier)
	d.Set("publicly_accessible", db.PubliclyAccessible)
	d.Set("storage_encrypted", db.StorageEncrypted)
	d.Set("ca_cert_identifier", db.CACertificateIdentifier)

	return diags
}

func resourceClusterInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)
	requestUpdate := false

	req := &docdb.ModifyDBInstanceInput{
		ApplyImmediately:     aws.Bool(d.Get("apply_immediately").(bool)),
		DBInstanceIdentifier: aws.String(d.Id()),
	}

	if d.HasChange("instance_class") {
		req.DBInstanceClass = aws.String(d.Get("instance_class").(string))
		requestUpdate = true
	}

	if d.HasChange("preferred_maintenance_window") {
		req.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("auto_minor_version_upgrade") {
		req.AutoMinorVersionUpgrade = aws.Bool(d.Get("auto_minor_version_upgrade").(bool))
		requestUpdate = true
	}

	if d.HasChange("promotion_tier") {
		req.PromotionTier = aws.Int64(int64(d.Get("promotion_tier").(int)))
		requestUpdate = true
	}

	if d.HasChange("ca_cert_identifier") {
		req.CACertificateIdentifier = aws.String(d.Get("ca_cert_identifier").(string))
		requestUpdate = true
	}

	if d.HasChange("enable_performance_insights") {
		req.EnablePerformanceInsights = aws.Bool(d.Get("enable_performance_insights").(bool))
		requestUpdate = true
	}

	if d.HasChange("performance_insights_kms_key_id") {
		req.PerformanceInsightsKMSKeyId = aws.String(d.Get("performance_insights_kms_key_id").(string))
		requestUpdate = true
	}

	if requestUpdate {
		err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
			_, err := conn.ModifyDBInstanceWithContext(ctx, req)
			if err != nil {
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return retry.RetryableError(err)
				}
				return retry.NonRetryableError(err)
			}
			return nil
		})
		if tfresource.TimedOut(err) {
			_, err = conn.ModifyDBInstanceWithContext(ctx, req)
		}
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DB Instance %s: %s", d.Id(), err)
		}

		// reuse db_instance refresh func
		stateConf := &retry.StateChangeConf{
			Pending:    resourceClusterInstanceCreateUpdatePendingStates,
			Target:     []string{"available"},
			Refresh:    resourceInstanceStateRefreshFunc(ctx, conn, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
			Delay:      30 * time.Second, // Wait 30 secs before starting
		}

		// Wait, catching any errors
		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster Instance (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterInstanceRead(ctx, d, meta)...)
}

func resourceClusterInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	opts := docdb.DeleteDBInstanceInput{DBInstanceIdentifier: aws.String(d.Id())}

	if _, err := conn.DeleteDBInstanceWithContext(ctx, &opts); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Cluster Instance (%s): %s", d.Id(), err)
	}

	// re-uses db_instance refresh func
	log.Println("[INFO] Waiting for DocumentDB Cluster Instance to be destroyed")
	stateConf := &retry.StateChangeConf{
		Pending:    resourceClusterInstanceDeletePendingStates,
		Target:     []string{},
		Refresh:    resourceInstanceStateRefreshFunc(ctx, conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster Instance (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func resourceInstanceStateRefreshFunc(ctx context.Context, conn *docdb.DocDB, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		v, err := resourceInstanceRetrieve(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return v, aws.StringValue(v.DBInstanceStatus), nil
	}
}

// resourceInstanceRetrieve fetches DBInstance information from the AWS
// API. It returns an error if there is a communication problem or unexpected
// error with AWS. When the DBInstance is not found, it returns no error and a
// nil pointer.
func resourceInstanceRetrieve(ctx context.Context, conn *docdb.DocDB, id string) (*docdb.DBInstance, error) {
	input := docdb.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(id),
	}
	out, err := conn.DescribeDBInstancesWithContext(ctx, &input)
	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBInstanceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	switch count := len(out.DBInstances); count {
	case 0:
		return nil, tfresource.NewEmptyResultError(input)
	case 1:
		return out.DBInstances[0], nil
	default:
		return nil, tfresource.NewTooManyResultsError(count, input)
	}
}

var resourceClusterInstanceCreateUpdatePendingStates = []string{
	"backing-up",
	"configuring-enhanced-monitoring",
	"configuring-iam-database-auth",
	"configuring-log-exports",
	"creating",
	"maintenance",
	"modifying",
	"rebooting",
	"renaming",
	"resetting-master-credentials",
	"starting",
	"storage-optimization",
	"upgrading",
}

var resourceClusterInstanceDeletePendingStates = []string{
	"configuring-log-exports",
	"modifying",
	"deleting",
}
