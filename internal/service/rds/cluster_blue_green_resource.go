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
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
		CreateWithoutTimeout: resourceClusterBlueGreenUpdate,
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

func resourceClusterBlueGreenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	connv2 := meta.(*conns.AWSClient).RDSClient(ctx)
	conn := meta.(*conns.AWSClient).RDSConn(ctx)
	deadline := tfresource.NewDeadline(d.Timeout(schema.TimeoutUpdate))
	dbc, _ := FindDBClusterByID(ctx, conn, d.Get("cluster_identifier").(string)) //d.Id())

	fmt.Printf("[DEBUG] DBClusterARN CREATE: %s", aws.StringValue(dbc.DBClusterArn))
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

	for _, value := range bluegreenDescribe.BlueGreenDeployments {
		bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
	}

	// _, err := waitBlueGreenClusterDeploymentAvailable(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining(), optFns...)
	// if err != nil {
	//	diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", aws.StringValue(&bluegreen[0]), err)
	// }

	// _, err = orchestrator.switchover(ctx, bluegreen[0], deadline.Remaining())

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
		log.Printf("[DEBUG] Checking if resource cleanup is enabled...")
		createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("blue-green-deployment-name"),
					Values: []string{"aurora-cluster-demo-3"},
				},
			},
		}
		bluegreen := createOut.BlueGreenDeploymentIdentifier

		log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment", d.Get("cluster_identifier").(string))

		if aws.StringValue(bluegreen) == "" {
			log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: deployment disappeared: %s", d.Get("cluster_identifier").(string), aws.StringValue(bluegreen))
			return
		}

		// Ensure that the Blue/Green Deployment is always cleaned up

		input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
			BlueGreenDeploymentIdentifier: bluegreen,
		}

		dep, err := waitBlueGreenDeploymenClusterSwitchoverCompleted(ctx, connv2, aws.StringValue(bluegreen), deadline.Remaining())

		if aws.StringValue(dep.Status) != "SWITCHOVER_COMPLETED" {
			log.Printf("[DEBUG] Setting cleanup mode to %b", aws.Bool(d.Get("cleanup_resources").(bool)))
			input.DeleteTarget = aws.Bool(d.Get("cleanup_resources").(bool))
		}

		_, err = connv2.DeleteBlueGreenDeployment(ctx, input)
		_, err = waitBlueGreenClusterDeploymentDeleted(ctx, connv2, aws.StringValue(bluegreen), deadline.Remaining())

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Get("cluster_identifier").(string), err)
			return
		}

		cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
			_, err = waitBlueGreenClusterDeploymentDeleted(ctx, connv2, aws.StringValue(bluegreen), deadline.Remaining(), optFns...)
			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", aws.StringValue(bluegreen), err)
			}
		})
	}()

	log.Printf("[DEBUG] Checking for create_deployment being true and switchover_enabled being true...")
	// We need to go from available state to switchover in progress if blue/green deploy is on pause
	if d.Get("switchover_enabled").(bool) {
		log.Printf("[DEBUG] switchover_enabled true...")
		log.Printf("[DEBUG] Switching over blue/green deployment...")
		orchestrator := newBlueGreenOrchestratorCluster(connv2)

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

		for _, value := range bluegreenDescribe.BlueGreenDeployments {
			bluegreen = append(bluegreen, *value.BlueGreenDeploymentIdentifier)
		}

		cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
			_, err := waitBlueGreenClusterDeploymentAvailable(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining(), optFns...)
			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", aws.StringValue(&bluegreen[0]), err)
			}
		})

		log.Printf("Switching over blue/green deployment...")
		_, err := orchestrator.switchover(ctx, bluegreen[0], deadline.Remaining())

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): switching over Blue/Green Deployment: waiting for completion: %s", aws.StringValue(&bluegreen[0]), err)
		}

		defer func() {
			log.Printf("[DEBUG] Cleaning up and not creating resources...")
			createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("blue-green-deployment-name"),
						Values: []string{"aurora-cluster-demo-3"},
					},
				},
			}

			bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)

			bluegreen := []string{}

			for _, value := range bluegreenDescribe.BlueGreenDeployments {
				bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
			}

			log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: %s", d.Get("cluster_identifier").(string), aws.StringValue(&bluegreen[0]))

			if aws.StringValue(&bluegreen[0]) == "" {
				log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: deployment disappeared: %s", d.Get("cluster_identifier").(string), aws.StringValue(&bluegreen[0]))
				return
			}

			// Ensure that the Blue/Green Deployment is always cleaned up

			input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
				BlueGreenDeploymentIdentifier: &bluegreen[0],
			}

			if aws.StringValue(bluegreenDescribe.BlueGreenDeployments[0].Status) != "SWITCHOVER_COMPLETED" {
				input.DeleteTarget = aws.Bool(d.Get("cleanup_resources").(bool))
			}

			_, err = connv2.DeleteBlueGreenDeployment(ctx, input)

			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Get("cluster_identifier").(string), err)
				return
			}

			cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
				_, err = waitBlueGreenClusterDeploymentDeleted(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining(), optFns...)
				if err != nil {
					diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", aws.StringValue(&bluegreen[0]), err)
				}
			})
		}()

		log.Printf("[DEBUG] Updating Blue/Green deployment (%s): Switching over Blue/Green Deployment", d.Get("cluster_identifier").(string))
		orchestrator.switchover(ctx, aws.StringValue(&bluegreen[0]), deadline.Remaining())
		_, err = waitBlueGreenDeploymenClusterSwitchoverCompleted(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining())

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Get("cluster_identifier").(string), err)
			return
		}

	}

	log.Printf("[DEBUG] Checking for create_deployment being false and switchover_enabled being true...")

	if !d.Get("create_deployment").(bool) && d.Get("switchover_enabled").(bool) {
		log.Printf("[DEBUG] Entering create_deployment being false and switchover_enabled being true...")

		orchestrator := newBlueGreenOrchestratorCluster(connv2)
		log.Printf("[DEBUG] Updating RDS DB Cluster: Creating Blue/Green Deployment")

		createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("blue-green-deployment-name"),
					Values: []string{d.Get("cluster_identifier").(string)},
				},
			},
		}

		bluegreenDescribe, err := connv2.DescribeBlueGreenDeployments(ctx, createOut)

		bluegreen := []string{}

		for _, value := range bluegreenDescribe.BlueGreenDeployments {
			bluegreen = append(bluegreen, *value.BlueGreenDeploymentIdentifier)
		}

		if bluegreen == nil {
			log.Printf("[DEBUG] Describing blue/green deployment: Error describing Blue/Green Deployment source")
		}
		defer func() {
			log.Printf("[DEBUG] Cleaning up and not creating resources...")
			createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("blue-green-deployment-name"),
						Values: []string{"aurora-cluster-demo-3"},
					},
				},
			}
			handler := newClusterHandler(connv2)
			err := handler.precondition(ctx, d)

			bluegreenDescribe, _ := connv2.DescribeBlueGreenDeployments(ctx, createOut)

			bluegreen := []string{}

			for _, value := range bluegreenDescribe.BlueGreenDeployments {
				bluegreen = append(bluegreen, aws.StringValue(value.BlueGreenDeploymentIdentifier))
			}

			log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: %s", d.Get("cluster_identifier").(string), aws.StringValue(&bluegreen[0]))

			if aws.StringValue(&bluegreen[0]) == "" {
				log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: deployment disappeared: %s", d.Get("cluster_identifier").(string), aws.StringValue(&bluegreen[0]))
				return
			}

			// Ensure that the Blue/Green Deployment is always cleaned up

			input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
				BlueGreenDeploymentIdentifier: &bluegreen[0],
			}

			if *bluegreenDescribe.BlueGreenDeployments[0].Status != "SWITCHOVER_COMPLETED" {
				input.DeleteTarget = aws.Bool(d.Get("cleanup_resources").(bool))
			}

			_, err = connv2.DeleteBlueGreenDeployment(ctx, input)

			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Get("cluster_identifier").(string), err)
				return
			}

			cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
				_, err = waitBlueGreenClusterDeploymentDeleted(ctx, connv2, aws.StringValue(&bluegreen[0]), deadline.Remaining(), optFns...)
				if err != nil {
					diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", aws.StringValue(&bluegreen[0]), err)
				}
			})
		}()
		_, err = orchestrator.switchover(ctx, bluegreen[0], deadline.Remaining())

		if err != nil {
			log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Error waiting for Blue/Green Deployment source switchover: %s", bluegreen[0], err)
		}
		_, err = orchestrator.waitForDeploymentAvailable(ctx, bluegreen[0], deadline.Remaining())

		if err != nil {
			log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Error running Blue/Green Deployment source switchover: %s", bluegreen[0], err)
		}
	}

	log.Printf("[DEBUG] Checking for create_deployment being false and switchover_enabled being false...")
	if !d.Get("create_deployment").(bool) && !d.Get("switchover_enabled").(bool) {
		log.Printf("[DEBUG]: Resources cannot be separately cleaned up after switchover completed. Nothing to do here")
	}

	return diags // append(diags, resourceClusterBlueGreenRead(ctx, d, meta)...)

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
			log.Printf("[ERROR] updating RDS DB Cluster (%s): deleting Blue/Green Deployment source: %s", d.Get("identifier").(string), err)
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
			log.Printf("[ERROR] updating RDS DB Instance (%s): deleting Blue/Green Deployment source: %s", d.Get("identifier").(string), err)
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
		createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("blue-green-deployment-name"),
					Values: []string{d.Get("cluster_identifier").(string)},
				},
			},
		}
		bluegreen := createOut.BlueGreenDeploymentIdentifier
		// orchestrator := newBlueGreenOrchestratorCluster(connv2)
		handler := newClusterHandler(connv2)
		err := handler.precondition(ctx, d)

		log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment", d.Get("cluster_identifier").(string))

		if bluegreen == nil {
			log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: deployment disappeared", d.Get("cluster_identifier").(string))
			return
		}

		// Ensure that the Blue/Green Deployment is always cleaned up

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Get("cluster_identifier").(string), err)
			return
		}
		cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
			dep, err := waitBlueGreenClusterDeploymentDeleted(ctx, connv2, aws.StringValue(bluegreen), deadline.Remaining(), optFns...)
			if err != nil {
				diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", aws.StringValue(bluegreen), err)
			}

			if d.Get("deletion_protection").(bool) {
				sourceARN, err := parseDBInstanceARN(aws.StringValue(dep.Source))
				input := &rds_sdkv2.ModifyDBClusterInput{
					ApplyImmediately:    true,
					DBClusterIdentifier: aws.String(sourceARN.Identifier),
					DeletionProtection:  aws.Bool(false),
				}

				err = dbClusterModify(ctx, connv2, d.Id(), input, deadline.Remaining())
				if err != nil {
					log.Printf("updating RDS DB Instance (%s): deleting Blue/Green Deployment source: disabling deletion protection: %s", d.Get("cluster_identifier").(string), err)
				}

			}
		},
		)
	}()

	createOut := &rds_sdkv2.DescribeBlueGreenDeploymentsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("blue-green-deployment-name"),
				Values: []string{d.Get("cluster_identifier").(string)},
			},
		},
	}

	bluegreen := createOut.BlueGreenDeploymentIdentifier
	handler := newClusterHandler(connv2)
	err := handler.precondition(ctx, d)

	log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment", d.Get("cluster_identifier").(string))

	if bluegreen == nil {
		log.Printf("[DEBUG] Updating RDS DB Cluster (%s): Deleting Blue/Green Deployment: deployment disappeared", d.Get("cluster_identifier").(string))
		return
	}

	// Ensure that the Blue/Green Deployment is always cleaned up

	input := &rds_sdkv2.DeleteBlueGreenDeploymentInput{
		BlueGreenDeploymentIdentifier: bluegreen,
	}

	input.DeleteTarget = aws.Bool(d.Get("cleanup_resources").(bool))

	_, err = connv2.DeleteBlueGreenDeployment(ctx, input)

	if err != nil {
		diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: %s", d.Get("cluster_identifier").(string), err)
		return
	}

	_, err = connv2.DeleteBlueGreenDeployment(ctx, input)

	cleaupWaiters = append(cleaupWaiters, func(optFns ...tfresource.OptionsFunc) {
		_, err = waitBlueGreenClusterDeploymentDeleted(ctx, connv2, aws.StringValue(bluegreen), deadline.Remaining(), optFns...)
		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "updating RDS DB Cluster (%s): deleting Blue/Green Deployment: waiting for completion: %s", aws.StringValue(bluegreen), err)
		}
	})

	return nil
}
