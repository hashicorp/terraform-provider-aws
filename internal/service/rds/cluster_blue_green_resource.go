// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	rds_sdkv2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/rds"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/exp/slices"
)

// @SDKResource("aws_rds_cluster_blue_green_deployment", name="ClusterBlueGreenDeployment")
// @Tags(identifierAttribute="arn")
func ResourceBlueGreenDeployment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterBlueGreenCreate,
		ReadWithoutTimeout:   resourceClusterBlueGreenRead,
		UpdateWithoutTimeout: resourceClusterBlueGreenUpdate,
		DeleteWithoutTimeout: resourceClusterBlueGreenDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
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
			"backup_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtMost(35),
			},
			"cleanup_resources": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cluster_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},

			"cluster_members": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_deployment": {
				Type:     schema.TypeBool,
				Required: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringMatch(regexache.MustCompile(fmt.Sprintf(`^%s.*$`, InstanceEngineCustomPrefix)), fmt.Sprintf("must begin with %s", InstanceEngineCustomPrefix)),
					validation.StringInSlice(ClusterEngine_Values(), false),
				),
			},
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"switchover_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,

			func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
				if !d.Get("create_deployment").(bool) {
					return nil
				}

				engine := d.Get("engine").(string)
				if !slices.Contains(dbClusterValidBlueGreenEngines(), engine) {
					return fmt.Errorf(`"blue_green_update.enabled" cannot be set when "engine" is %q.`, engine)
				}
				return nil
			},
		),
	}
}

func parseDBClusterARN(s string) (dbClusterARN, error) {
	arn, err := arn.Parse(s)
	if err != nil {
		return dbClusterARN{}, err
	}

	result := dbClusterARN{
		ARN: arn,
	}

	re := regexache.MustCompile(`^db:([0-9a-z-]+)$`)
	matches := re.FindStringSubmatch(arn.Resource)
	if matches == nil || len(matches) != 2 {
		return dbClusterARN{}, errors.New("DB Cluster ARN: invalid resource section")
	}
	result.dbClusterARN = matches[1]

	return result, nil
}

func waitDBClusterAvailableSDKv2(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*rds.DBCluster, error) { //nolint:unparam
	options := tfresource.Options{
		PollInterval:              10 * time.Second,
		Delay:                     1 * time.Minute,
		ContinuousTargetOccurence: 3,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{
			ClusterStatusBackingUp,
			ClusterStatusConfiguringIAMDatabaseAuth,
			ClusterStatusCreating,
			ClusterStatusDeleting,
			ClusterStatusMigrating,
			ClusterStatusModifying,
			ClusterStatusPreparingDataMigration,
			ClusterStatusRebooting,
			ClusterStatusRenaming,
			ClusterStatusResettingMasterCredentials,
			ClusterStatusScalingCompute,
			ClusterStatusUpgrading,
		},
		Target:  []string{ClusterStatusAvailable},
		Refresh: statusDBInstanceSDKv2(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func dbClusterValidBlueGreenEngines() []string {
	return []string{
		ClusterEngineAuroraMySQL,
	}
}

func resourceClusterBlueGreenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	connv2 := meta.(*conns.AWSClient).RDSClient(ctx)
	conn := meta.(*conns.AWSClient).RDSConn(ctx)
	deadline := tfresource.NewDeadline(d.Timeout(schema.TimeoutUpdate))
	dbc, _ := FindDBClusterByID(ctx, conn, d.Get("cluster_identifier").(string)) //d.Id())

	fmt.Printf("[DEBUG] DBClusterARN UPDATE: %s", aws.StringValue(dbc.DBClusterArn))
	d.Set("arn", dbc.DBClusterArn)
	d.Set("cluster_identifier", dbc.DBClusterIdentifier)
	var clusterMembers []string
	for _, v := range dbc.DBClusterMembers {
		clusterMembers = append(clusterMembers, aws.StringValue(v.DBInstanceIdentifier))
	}
	d.Set("cluster_members", clusterMembers)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)

	setTagsOut(ctx, dbc.TagList)
	var cleaupWaiters []func(optFns ...tfresource.OptionsFunc)

	defer func() {
		if len(cleaupWaiters) == 0 {
			return
		}

		waiter, waiters := cleaupWaiters[0], cleaupWaiters[1:]
		waiter()
		for _, waiter := range waiters {
			// Skip the delay for subsequent waiters. Since we're waiting for all of the waiters
			// to complete, we don't need to run them concurrently, saving on network traffic.
			waiter(tfresource.WithDelay(0))
		}
	}()

	if d.Get("create_deployment").(bool) {
		log.Printf("[DEBUG] Describing blue/green deplyments...")
		log.Printf("[DEBUG] Implementing handler...")
		handler := newClusterHandler(connv2)
		log.Printf("[DEBUG] Creating input...")
		createIn := handler.createBlueGreenInput(d)

		orchestrator := newBlueGreenOrchestratorCluster(connv2)

		_, err := orchestrator.createDeploymentCluster(ctx, createIn)
		if err != nil {
			log.Printf("[DEBUG] Something went wrong on handler precondition: %s", err)
		}

		createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("blue-green-deployment-name"),
					Values: []string{d.Get("cluster_identifier").(string)},
				},
			},
		}
		// handler := newClusterHandler(connv2)
		// err := handler.precondition(ctx, d)

		bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)

		bluegreen := []string{}

		for _, value := range bluegreenDescribe.BlueGreenDeployments {
			bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
		}

		_, err = orchestrator.waitForDeploymentAvailable(ctx, aws.StringValue(&bluegreen[0]), deadline.Remaining())
		d.SetId(aws.StringValue(&bluegreen[0]))
	}

	if !d.Get("create_deployment").(bool) {
		log.Printf("Deployment disabled, checking for switchover...")
	}

	if !d.Get("switchover_enabled").(bool) {
		log.Printf("Switchover disabled. Make sure to delete dangling resources manually when done.")
	}

	defer func() {
		log.Printf("[DEBUG] Verifying cleanup mode...")
		createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("blue-green-deployment-name"),
					Values: []string{d.Get("cluster_identifier").(string)},
				},
			},
		}

		bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)
		bluegreen := []string{}
		// handler := newClusterHandler(connv2)
		// err := handler.precondition(ctx, d)

		for _, value := range bluegreenDescribe.BlueGreenDeployments {
			bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
		}

		if d.Get("switchover_enabled").(bool) {
			createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("blue-green-deployment-name"),
						Values: []string{d.Get("cluster_identifier").(string)},
					},
				},
			}
			// handler := newClusterHandler(connv2)
			// err := handler.precondition(ctx, d)

			bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)

			bluegreen := []string{}

			for _, value := range bluegreenDescribe.BlueGreenDeployments {
				bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
			}

			orchestrator := newBlueGreenOrchestratorCluster(connv2)
			log.Printf("[DEBUG] Updating Blue/Green deployment (%s): Switching over Blue/Green Deployment", bluegreen[0])

			input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
				BlueGreenDeploymentIdentifier: aws.String(d.Id()),
			}

			if d.Get("switchover_enabled").(bool) {

				_, err := orchestrator.switchover(ctx, aws.StringValue(&bluegreen[0]), deadline.Remaining())
				if aws.StringValue(bluegreenDescribe.BlueGreenDeployments[0].Status) != "SWITCHOVER_COMPLETED" {

					input.DeleteTarget = aws.Bool(d.Get("cleanup_resources").(bool))
				}
				_, err = waitBlueGreenDeploymenClusterSwitchoverCompleted(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining())
				if err != nil {
					log.Printf("[DEBUG] Error switching over: %s", err)
				}
			}

			if !d.Get("switchover_enabled").(bool) {
				log.Printf("[DEBUG] Switchover disabled so we are finished. Make sure to manually delete previous resources when done.")
			}
		}

		// dep, err := waitBlueGreenDeploymenClusterSwitchoverCompleted(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining())

		identifier := d.Get("cluster_identifier").(string)

		_, err := waitDBClusterAvailableSDKv2(ctx, connv2, identifier, deadline.Remaining())

		if err != nil {
			log.Printf("[ERROR] Updating RDS DB Cluster: creating Blue/Green Deployment: waiting for Green environment")
		}

		if err != nil {
			fmt.Printf("[ERROR] updating RDS DB Cluster: creating Blue/Green Deployment: waiting for Green environment")
		}

		log.Printf("[DEBUG] Updating RDS DB Deployment (%s): Deleting Blue/Green Deployment", d.Id())

		if len(d.Id()) < 1 {
			log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: deployment disappeared: %s", d.Get("cluster_identifier").(string), d.Id())
		}

		// Ensure that the Blue/Green Deployment is always cleaned up

		input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
			BlueGreenDeploymentIdentifier: aws.String(d.Id()),
		}

		_, err = connv2.DeleteBlueGreenDeployment(ctx, input)

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Id(), err)
			return
		}

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Id(), err)
			return
		}

		cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
			_, err = waitBlueGreenClusterDeploymentDeleted(ctx, connv2, d.Id(), deadline.Remaining(), optFns...)
			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", d.Id(), err)
			}
		})

		sourceARN, err := parseDBClusterARN(d.Get("cluster_identifier").(string))
		if err != nil {
			log.Printf("[ERROR] updating RDS DB Cluster (%s): deleting Blue/Green Deployment source: %s", d.Get("cluster_identifier").(string), err)
		}

		if d.Get("deletion_protection").(bool) {
			input := &rds_sdkv2.ModifyDBClusterInput{
				ApplyImmediately:    true,
				DBClusterIdentifier: aws.String(sourceARN.dbClusterARN),
				DeletionProtection:  aws.Bool(false),
			}

			err := dbClusterModify(ctx, connv2, sourceARN.dbClusterARN, input, deadline.Remaining())
			if err != nil {
				fmt.Printf("[ERROR] Updating RDS DB Instance (%s): deleting Blue/Green Deployment source: disabling deletion protection: %s", d.Get("identifier").(string), err)
			}
		}

		deleteInput := &rds_sdkv2.DeleteDBClusterInput{
			DBClusterIdentifier: aws.String(sourceARN.dbClusterARN),
		}

		_, err = tfresource.RetryWhen(ctx, 5*time.Minute,
			func() (any, error) {
				return connv2.DeleteDBCluster(ctx, deleteInput)
			},
			func(err error) (bool, error) {
				// Retry for IAM eventual consistency.
				if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
					return true, err
				}

				if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterCombination, "disable deletion pro") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			log.Printf("[ERROR] updating RDS DB Instance (%s): deleting Blue/Green Deployment source: %s", d.Get("icluster_dentifier").(string), err)
		}

		cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
			_, err = waitDBInstanceDeleted(ctx, meta.(*conns.AWSClient).RDSConn(ctx), aws.StringValue(&sourceARN.dbClusterARN), deadline.Remaining(), optFns...)
			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): deleting Blue/Green Deployment source: waiting for completion: %s", d.Get("identifier").(string), err)
			}
		})

		if diags.HasError() {
			return
		}
	}()
	return diags
}

func resourceClusterBlueGreenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	dbc, _ := FindDBClusterByID(ctx, conn, d.Get("cluster_identifier").(string)) //d.Id())

	fmt.Printf("[DEBUG] DBClusterARN READ: %s", aws.StringValue(dbc.DBClusterArn))
	d.Set("arn", dbc.DBClusterArn)
	d.Set("cluster_identifier", dbc.DBClusterIdentifier)
	var clusterMembers []string
	for _, v := range dbc.DBClusterMembers {
		clusterMembers = append(clusterMembers, aws.StringValue(v.DBInstanceIdentifier))
	}
	d.Set("cluster_members", clusterMembers)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)

	setTagsOut(ctx, dbc.TagList)

	return diags
}

func statusBlueGreenClusterDeployment(ctx context.Context, conn *rds_sdkv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findBlueGreenDeploymentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitBlueGreenDeploymentClusterAvailable(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"PROVISIONING"},
		Target:  []string{"AVAILABLE"},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		return output, err
	}

	return nil, err
}

func waitBlueGreenDeploymenClusterSwitchoverDeleting(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"SWITCHOVER_COMPLETED"},
		Target:  []string{"SWITCHOVER_IN_PROGRESS"},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		if status := aws.StringValue(output.Status); status == "INVALID_CONFIGURATION" || status == "SWITCHOVER_FAILED" {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusDetails)))
		}

		return output, err
	}

	return nil, err
}

func waitBlueGreenDeploymenClusterSwitchoverInProgress(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"SWITCHOVER_IN_PROGRESS"},
		Target:  []string{"SWITCHOVER_COMPLETED"},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		if status := aws.StringValue(output.Status); status == "INVALID_CONFIGURATION" || status == "SWITCHOVER_FAILED" {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusDetails)))
		}

		return output, err
	}

	return nil, err
}

func waitBlueGreenDeploymenClusterSwitchoverAvailable(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"AVAILABLE"},
		Target:  []string{"SWITCHOVER_IN_PROGRESS"},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		if status := aws.StringValue(output.Status); status == "INVALID_CONFIGURATION" || status == "SWITCHOVER_FAILED" {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusDetails)))
		}

		return output, err
	}

	return nil, err
}

func waitBlueGreenDeploymenClusterSwitchoverCompleted(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"SWITCHOVER_IN_PROGRESS"},
		Target:  []string{"SWITCHOVER_COMPLETED"},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		if status := aws.StringValue(output.Status); status == "INVALID_CONFIGURATION" || status == "SWITCHOVER_FAILED" {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusDetails)))
		}

		return output, err
	}

	return nil, err
}

func waitBlueGreenClusterDeploymentAvailable(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"PROVISIONING", "SWITCHOVER_IN_PROGRESS", "SWITCHOVER_COMPLETED", "INVALID_CONFIGURATION", "SWITCHOVER_FAILED", "DELETING"},
		Target:  []string{"AVAILABLE"},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		return output, err
	}

	return nil, err
}

func waitBlueGreenClusterDeploymentDeleted(ctx context.Context, conn *rds_sdkv2.Client, id string, timeout time.Duration, optFns ...tfresource.OptionsFunc) (*types.BlueGreenDeployment, error) {
	options := tfresource.Options{
		PollInterval: 10 * time.Second,
		Delay:        1 * time.Minute,
	}
	for _, fn := range optFns {
		fn(&options)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{"PROVISIONING", "AVAILABLE", "SWITCHOVER_IN_PROGRESS", "SWITCHOVER_COMPLETED", "INVALID_CONFIGURATION", "SWITCHOVER_FAILED", "DELETING"},
		Target:  []string{},
		Refresh: statusBlueGreenDeployment(ctx, conn, id),
		Timeout: timeout,
	}
	options.Apply(stateConf)

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.BlueGreenDeployment); ok {
		return output, err
	}

	return nil, err
}
func dbClusterPopulateModify(input *rds_sdkv2.ModifyDBClusterInput, d *schema.ResourceData) bool {
	needsModify := false

	if d.HasChanges("allocated_storage", "iops") {
		needsModify = true
		input.AllocatedStorage = aws.Int32(int32(d.Get("allocated_storage").(int)))

		// Send Iops if it has changed or not (StorageType == "gp3" and AllocatedStorage < threshold).
		if d.HasChange("iops") || !isStorageTypeGP3BelowAllocatedStorageThreshold(d) {
			input.Iops = aws.Int32(int32(d.Get("iops").(int)))
		}
	}

	if d.HasChange("auto_minor_version_upgrade") {
		needsModify = true
		input.AutoMinorVersionUpgrade = aws.Bool(d.Get("auto_minor_version_upgrade").(bool))
	}

	if d.HasChange("backup_retention_period") {
		needsModify = true
		input.BackupRetentionPeriod = aws.Int32(int32(d.Get("backup_retention_period").(int)))
	}

	if d.HasChange("backup_window") {
		needsModify = true
		input.PreferredBackupWindow = aws.String(d.Get("backup_window").(string))
	}

	if d.HasChange("copy_tags_to_snapshot") {
		needsModify = true
		input.CopyTagsToSnapshot = aws.Bool(d.Get("copy_tags_to_snapshot").(bool))
	}

	if d.HasChange("deletion_protection") {
		needsModify = true
	}
	// Always set this. Fixes TestAccRDSCluster_BlueGreenDeployment_updateWithDeletionProtection
	input.DeletionProtection = aws.Bool(d.Get("deletion_protection").(bool))

	if d.HasChanges("domain", "domain_iam_role_name") {
		needsModify = true
		input.Domain = aws.String(d.Get("domain").(string))
		input.DomainIAMRoleName = aws.String(d.Get("domain_iam_role_name").(string))
	}

	if d.HasChange("enabled_cloudwatch_logs_exports") {
		needsModify = true
		oraw, nraw := d.GetChange("enabled_cloudwatch_logs_exports")
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)

		enable := n.Difference(o)
		disable := o.Difference(n)

		input.CloudwatchLogsExportConfiguration = &types.CloudwatchLogsExportConfiguration{
			EnableLogTypes:  flex.ExpandStringValueSet(enable),
			DisableLogTypes: flex.ExpandStringValueSet(disable),
		}
	}

	if d.HasChange("iam_database_authentication_enabled") {
		needsModify = true
		input.EnableIAMDatabaseAuthentication = aws.Bool(d.Get("iam_database_authentication_enabled").(bool))
	}

	if d.HasChange("identifier") {
		needsModify = true
		input.NewDBClusterIdentifier = aws.String(d.Get("cluster_identifier").(string))
	}

	if d.HasChange("maintenance_window") {
		needsModify = true
		input.PreferredMaintenanceWindow = aws.String(d.Get("maintenance_window").(string))
	}

	if d.HasChange("manage_master_user_password") {
		needsModify = true
		input.ManageMasterUserPassword = aws.Bool(d.Get("manage_master_user_password").(bool))
	}

	if d.HasChange("master_user_secret_kms_key_id") {
		needsModify = true
		if v, ok := d.GetOk("master_user_secret_kms_key_id"); ok {
			input.MasterUserSecretKmsKeyId = aws.String(v.(string))
			// InvalidParameterValue: A ManageMasterUserPassword value is required when MasterUserSecretKmsKeyId is specified.
			input.ManageMasterUserPassword = aws.Bool(d.Get("manage_master_user_password").(bool))
		}
	}

	if d.HasChange("monitoring_interval") {
		needsModify = true
		input.MonitoringInterval = aws.Int32(int32(d.Get("monitoring_interval").(int)))
	}

	if d.HasChange("monitoring_role_arn") {
		needsModify = true
		input.MonitoringRoleArn = aws.String(d.Get("monitoring_role_arn").(string))
	}

	if d.HasChange("network_type") {
		needsModify = true
		input.NetworkType = aws.String(d.Get("network_type").(string))
	}

	if d.HasChange("option_group_name") {
		needsModify = true
		input.OptionGroupName = aws.String(d.Get("option_group_name").(string))
	}

	if d.HasChange("password") {
		needsModify = true
		// With ManageMasterUserPassword set to true, the password is no longer needed, so we omit it from the API call.
		if v, ok := d.GetOk("password"); ok {
			input.MasterUserPassword = aws.String(v.(string))
		}
	}

	if d.HasChanges("performance_insights_enabled", "performance_insights_kms_key_id", "performance_insights_retention_period") {
		needsModify = true
		input.EnablePerformanceInsights = aws.Bool(d.Get("performance_insights_enabled").(bool))

		if v, ok := d.GetOk("performance_insights_kms_key_id"); ok {
			input.PerformanceInsightsKMSKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("performance_insights_retention_period"); ok {
			input.PerformanceInsightsRetentionPeriod = aws.Int32(int32(v.(int)))
		}
	}

	if d.HasChange("port") {
		needsModify = true
		input.Port = aws.Int32(int32(d.Get("port").(int)))
	}

	if d.HasChange("storage_type") {
		needsModify = true
		input.StorageType = aws.String(d.Get("storage_type").(string))

		if aws.StringValue(input.StorageType) == storageTypeIO1 {
			input.Iops = aws.Int32(int32(d.Get("iops").(int)))
		}
	}

	if d.HasChange("vpc_security_group_ids") {
		if v := d.Get("vpc_security_group_ids").(*schema.Set); v.Len() > 0 {
			needsModify = true
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
		}
	}

	return needsModify
}

func dbClusterModify(ctx context.Context, conn *rds_sdkv2.Client, resourceID string, input *rds_sdkv2.ModifyDBClusterInput, timeout time.Duration) error {
	_, err := tfresource.RetryWhen(ctx, timeout,
		func() (interface{}, error) {
			return conn.ModifyDBCluster(ctx, input)
		},
		func(err error) (bool, error) {
			// Retry for IAM eventual consistency.
			if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
				return true, err
			}

			if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterCombination, "previous storage change is being optimized") {
				return true, err
			}

			if errs.IsA[*types.InvalidDBClusterStateFault](err) {
				return true, err
			}

			return false, err
		},
	)
	if err != nil {
		return err
	}

	if _, err := waitDBClusterAvailableSDKv2(ctx, conn, resourceID, timeout); err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}
	return nil
}

func resourceClusterBlueGreenUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	connv2 := meta.(*conns.AWSClient).RDSClient(ctx)
	conn := meta.(*conns.AWSClient).RDSConn(ctx)
	deadline := tfresource.NewDeadline(d.Timeout(schema.TimeoutUpdate))
	dbc, _ := FindDBClusterByID(ctx, conn, d.Get("cluster_identifier").(string)) //d.Id())

	fmt.Printf("[DEBUG] DBClusterARN UPDATE: %s", aws.StringValue(dbc.DBClusterArn))
	d.Set("arn", dbc.DBClusterArn)
	d.Set("cluster_identifier", dbc.DBClusterIdentifier)
	var clusterMembers []string
	for _, v := range dbc.DBClusterMembers {
		clusterMembers = append(clusterMembers, aws.StringValue(v.DBInstanceIdentifier))
	}
	d.Set("cluster_members", clusterMembers)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)

	setTagsOut(ctx, dbc.TagList)
	var cleaupWaiters []func(optFns ...tfresource.OptionsFunc)

	defer func() {
		if len(cleaupWaiters) == 0 {
			return
		}

		waiter, waiters := cleaupWaiters[0], cleaupWaiters[1:]
		waiter()
		for _, waiter := range waiters {
			// Skip the delay for subsequent waiters. Since we're waiting for all of the waiters
			// to complete, we don't need to run them concurrently, saving on network traffic.
			waiter(tfresource.WithDelay(0))
		}
	}()

	if d.Get("create_deployment").(bool) {
		log.Printf("[DEBUG] Describing blue/green deplyments...")
		log.Printf("[DEBUG] Implementing handler...")
		handler := newClusterHandler(connv2)
		log.Printf("[DEBUG] Creating input...")
		createIn := handler.createBlueGreenInput(d)

		orchestrator := newBlueGreenOrchestratorCluster(connv2)

		_, err := orchestrator.createDeploymentCluster(ctx, createIn)
		if err != nil {
			log.Printf("[DEBUG] Something went wrong on handler precondition: %s", err)
		}

		createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("blue-green-deployment-name"),
					Values: []string{d.Get("cluster_identifier").(string)},
				},
			},
		}
		// handler := newClusterHandler(connv2)
		// err := handler.precondition(ctx, d)

		bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)

		bluegreen := []string{}

		for _, value := range bluegreenDescribe.BlueGreenDeployments {
			bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
		}

		_, err = orchestrator.waitForDeploymentAvailable(ctx, aws.StringValue(&bluegreen[0]), deadline.Remaining())
		d.SetId(aws.StringValue(&bluegreen[0]))
	}

	if !d.Get("create_deployment").(bool) {
		log.Printf("Deployment disabled, checking for switchover...")
	}

	if !d.Get("switchover_enabled").(bool) {
		log.Printf("Switchover disabled. Make sure to delete dangling resources manually when done.")
	}

	defer func() {
		log.Printf("[DEBUG] Verifying cleanup mode...")
		createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("blue-green-deployment-name"),
					Values: []string{d.Get("cluster_identifier").(string)},
				},
			},
		}

		bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)
		bluegreen := []string{}
		// handler := newClusterHandler(connv2)
		// err := handler.precondition(ctx, d)

		for _, value := range bluegreenDescribe.BlueGreenDeployments {
			bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
		}

		if d.Get("switchover_enabled").(bool) {
			createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("blue-green-deployment-name"),
						Values: []string{d.Get("cluster_identifier").(string)},
					},
				},
			}
			// handler := newClusterHandler(connv2)
			// err := handler.precondition(ctx, d)

			bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)

			bluegreen := []string{}

			for _, value := range bluegreenDescribe.BlueGreenDeployments {
				bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
			}

			orchestrator := newBlueGreenOrchestratorCluster(connv2)
			log.Printf("[DEBUG] Updating Blue/Green deployment (%s): Switching over Blue/Green Deployment", bluegreen[0])

			input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
				BlueGreenDeploymentIdentifier: aws.String(d.Id()),
			}

			if d.Get("switchover_enabled").(bool) {

				_, err := orchestrator.switchover(ctx, aws.StringValue(&bluegreen[0]), deadline.Remaining())
				if aws.StringValue(bluegreenDescribe.BlueGreenDeployments[0].Status) != "SWITCHOVER_COMPLETED" {

					input.DeleteTarget = aws.Bool(d.Get("cleanup_resources").(bool))
				}
				_, err = waitBlueGreenDeploymenClusterSwitchoverCompleted(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining())
				if err != nil {
					log.Printf("[DEBUG] Error switching over: %s", err)
				}
			}

			if !d.Get("switchover_enabled").(bool) {
				log.Printf("[DEBUG] Switchover disabled so we are finished. Make sure to manually delete previous resources when done.")
			}
		}

		// dep, err := waitBlueGreenDeploymenClusterSwitchoverCompleted(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining())

		identifier := d.Get("cluster_identifier").(string)

		_, err := waitDBClusterAvailableSDKv2(ctx, connv2, identifier, deadline.Remaining())

		if err != nil {
			log.Printf("[ERROR] Updating RDS DB Cluster: creating Blue/Green Deployment: waiting for Green environment")
		}

		if err != nil {
			fmt.Printf("[ERROR] updating RDS DB Cluster: creating Blue/Green Deployment: waiting for Green environment")
		}

		log.Printf("[DEBUG] Updating RDS DB Deployment (%s): Deleting Blue/Green Deployment", d.Id())

		if len(d.Id()) < 1 {
			log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: deployment disappeared: %s", d.Get("cluster_identifier").(string), d.Id())
		}

		// Ensure that the Blue/Green Deployment is always cleaned up

		input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
			BlueGreenDeploymentIdentifier: aws.String(d.Id()),
		}

		_, err = connv2.DeleteBlueGreenDeployment(ctx, input)

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Id(), err)
			return
		}

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Id(), err)
			return
		}

		cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
			_, err = waitBlueGreenClusterDeploymentDeleted(ctx, connv2, d.Id(), deadline.Remaining(), optFns...)
			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", d.Id(), err)
			}
		})

		sourceARN, err := parseDBClusterARN(d.Get("cluster_identifier").(string))
		if err != nil {
			log.Printf("[ERROR] updating RDS DB Cluster (%s): deleting Blue/Green Deployment source: %s", d.Get("cluster_identifier").(string), err)
		}

		if d.Get("deletion_protection").(bool) {
			input := &rds_sdkv2.ModifyDBClusterInput{
				ApplyImmediately:    true,
				DBClusterIdentifier: aws.String(sourceARN.dbClusterARN),
				DeletionProtection:  aws.Bool(false),
			}

			err := dbClusterModify(ctx, connv2, sourceARN.dbClusterARN, input, deadline.Remaining())
			if err != nil {
				fmt.Printf("[ERROR] Updating RDS DB Instance (%s): deleting Blue/Green Deployment source: disabling deletion protection: %s", d.Get("identifier").(string), err)
			}
		}

		deleteInput := &rds_sdkv2.DeleteDBClusterInput{
			DBClusterIdentifier: aws.String(sourceARN.dbClusterARN),
		}

		_, err = tfresource.RetryWhen(ctx, 5*time.Minute,
			func() (any, error) {
				return connv2.DeleteDBCluster(ctx, deleteInput)
			},
			func(err error) (bool, error) {
				// Retry for IAM eventual consistency.
				if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
					return true, err
				}

				if tfawserr_sdkv2.ErrMessageContains(err, errCodeInvalidParameterCombination, "disable deletion pro") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			log.Printf("[ERROR] updating RDS DB Instance (%s): deleting Blue/Green Deployment source: %s", d.Get("icluster_dentifier").(string), err)
		}

		cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
			_, err = waitDBInstanceDeleted(ctx, meta.(*conns.AWSClient).RDSConn(ctx), aws.StringValue(&sourceARN.dbClusterARN), deadline.Remaining(), optFns...)
			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Instance (%s): deleting Blue/Green Deployment source: waiting for completion: %s", d.Get("identifier").(string), err)
			}
		})

		if diags.HasError() {
			return
		}
	}()

	return diags
}

func resourceClusterBlueGreenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	connv2 := meta.(*conns.AWSClient).RDSClient(ctx)
	deadline := tfresource.NewDeadline(d.Timeout(schema.TimeoutUpdate))
	conn := meta.(*conns.AWSClient).RDSConn(ctx)
	dbc, _ := FindDBClusterByID(ctx, conn, d.Get("cluster_identifier").(string)) //d.Id())

	fmt.Printf("[DEBUG] DBClusterARN: %s", aws.StringValue(dbc.DBClusterArn))
	d.Set("arn", dbc.DBClusterArn)
	d.Set("cluster_identifier", dbc.DBClusterIdentifier)
	var clusterMembers []string
	for _, v := range dbc.DBClusterMembers {
		clusterMembers = append(clusterMembers, aws.StringValue(v.DBInstanceIdentifier))
	}
	d.Set("cluster_members", clusterMembers)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)

	setTagsOut(ctx, dbc.TagList)

	var cleaupWaiters []func(optFns ...tfresource.OptionsFunc)
	defer func() {
		if len(cleaupWaiters) == 0 {
			return
		}

		waiter, waiters := cleaupWaiters[0], cleaupWaiters[1:]
		waiter()
		for _, waiter := range waiters {
			// Skip the delay for subsequent waiters. Since we're waiting for all of the waiters
			// to complete, we don't need to run them concurrently, saving on network traffic.
			waiter(tfresource.WithDelay(0))
		}
	}()

	defer func() {
		log.Printf("[DEBUG] Verifying cleanup mode...")
		createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("blue-green-deployment-name"),
					Values: []string{d.Get("cluster_identifier").(string)},
				},
			},
		}

		bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)
		bluegreen := []string{}
		// handler := newClusterHandler(connv2)
		// err := handler.precondition(ctx, d)

		for _, value := range bluegreenDescribe.BlueGreenDeployments {
			bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
		}

		if d.Get("switchover_enabled").(bool) {
			createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("blue-green-deployment-name"),
						Values: []string{d.Get("cluster_identifier").(string)},
					},
				},
			}
			// handler := newClusterHandler(connv2)
			// err := handler.precondition(ctx, d)

			bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)

			bluegreen := []string{}

			for _, value := range bluegreenDescribe.BlueGreenDeployments {
				bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
			}

			orchestrator := newBlueGreenOrchestratorCluster(connv2)
			log.Printf("[DEBUG] Updating Blue/Green deployment (%s): Switching over Blue/Green Deployment", bluegreen[0])

			input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
				BlueGreenDeploymentIdentifier: aws.String(d.Id()),
			}

			if d.Get("switchover_enabled").(bool) {

				_, err := orchestrator.switchover(ctx, aws.StringValue(&bluegreen[0]), deadline.Remaining())
				if aws.StringValue(bluegreenDescribe.BlueGreenDeployments[0].Status) != "SWITCHOVER_COMPLETED" {

					input.DeleteTarget = aws.Bool(d.Get("cleanup_resources").(bool))
				}
				_, err = waitBlueGreenDeploymenClusterSwitchoverCompleted(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining())
				if err != nil {
					log.Printf("[DEBUG] Error switching over: %s", err)
				}
			}

			if !d.Get("switchover_enabled").(bool) {
				log.Printf("[DEBUG] Switchover disabled so we are finished. Make sure to manually delete previous resources when done.")
			}
		}

		// dep, err := waitBlueGreenDeploymenClusterSwitchoverCompleted(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining())

		identifier := d.Get("cluster_identifier").(string)

		_, err := waitDBClusterAvailableSDKv2(ctx, connv2, identifier, deadline.Remaining())

		if err != nil {
			log.Printf("[ERROR] Updating RDS DB Cluster: creating Blue/Green Deployment: waiting for Green environment")
		}

		if err != nil {
			fmt.Printf("[ERROR] updating RDS DB Cluster: creating Blue/Green Deployment: waiting for Green environment")
		}

		log.Printf("[DEBUG] Updating RDS DB Deployment (%s): Deleting Blue/Green Deployment", d.Id())

		if len(d.Id()) < 1 {
			log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: deployment disappeared: %s", d.Get("cluster_identifier").(string), d.Id())
		}

		// Ensure that the Blue/Green Deployment is always cleaned up

		input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
			BlueGreenDeploymentIdentifier: aws.String(d.Id()),
		}

		_, err = connv2.DeleteBlueGreenDeployment(ctx, input)

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Id(), err)
			return
		}

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Id(), err)
			return
		}

		cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
			_, err = waitBlueGreenClusterDeploymentDeleted(ctx, connv2, d.Id(), deadline.Remaining(), optFns...)
			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", d.Id(), err)
			}
		})
	}()
	return diags
}
