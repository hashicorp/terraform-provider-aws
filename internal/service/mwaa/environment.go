// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mwaa

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mwaa"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	propagationTimeout = 2 * time.Minute
)

// @SDKResource("aws_mwaa_environment", name="Environment")
// @Tags(identifierAttribute="arn")
func ResourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnvironmentCreate,
		ReadWithoutTimeout:   resourceEnvironmentRead,
		UpdateWithoutTimeout: resourceEnvironmentUpdate,
		DeleteWithoutTimeout: resourceEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"airflow_configuration_options": {
				Type:      schema.TypeMap,
				Optional:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
			"airflow_version": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dag_s3_path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment_class": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"kms_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
				ForceNew:     true,
			},
			"last_updated": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"error": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"error_code": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"error_message": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"logging_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dag_processing_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     environmentModuleLoggingConfigurationSchema(),
						},
						"scheduler_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     environmentModuleLoggingConfigurationSchema(),
						},
						"task_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     environmentModuleLoggingConfigurationSchema(),
						},
						"webserver_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     environmentModuleLoggingConfigurationSchema(),
						},
						"worker_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     environmentModuleLoggingConfigurationSchema(),
						},
					},
				},
			},
			"max_workers": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"min_workers": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							MinItems: 2,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"plugins_s3_object_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"plugins_s3_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"requirements_s3_object_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"requirements_s3_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"schedulers": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"service_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_bucket_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"startup_script_s3_object_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"startup_script_s3_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"webserver_access_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(mwaa.WebserverAccessMode_Values(), false),
			},
			"webserver_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_window_start": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIf("airflow_version", func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				o, n := d.GetChange("airflow_version")

				if oldVersion, err := gversion.NewVersion(o.(string)); err == nil {
					if newVersion, err := gversion.NewVersion(n.(string)); err == nil {
						// https://docs.aws.amazon.com/mwaa/latest/userguide/airflow-versions.html#airflow-versions-upgrade:
						// 	Amazon MWAA supports minor version upgrades.
						// 	This means you can upgrade your environment from version x.4.z to x.5.z.
						// 	However, you cannot upgrade your environment to a new major version of Apache Airflow.
						// 	For example, upgrading from version 1.y.z to 2.y.z is not supported.
						return oldVersion.Segments()[0] < newVersion.Segments()[0]
					}
				}

				return false
			}),
			verify.SetTagsDiff,
		),
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MWAAConn(ctx)

	name := d.Get("name").(string)
	input := &mwaa.CreateEnvironmentInput{
		DagS3Path:            aws.String(d.Get("dag_s3_path").(string)),
		ExecutionRoleArn:     aws.String(d.Get("execution_role_arn").(string)),
		Name:                 aws.String(name),
		NetworkConfiguration: expandEnvironmentNetworkConfigurationCreate(d.Get("network_configuration").([]interface{})),
		SourceBucketArn:      aws.String(d.Get("source_bucket_arn").(string)),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("airflow_configuration_options"); ok {
		input.AirflowConfigurationOptions = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("airflow_version"); ok {
		input.AirflowVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("environment_class"); ok {
		input.EnvironmentClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key"); ok {
		input.KmsKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_configuration"); ok {
		input.LoggingConfiguration = expandEnvironmentLoggingConfiguration(v.([]interface{}))
	}

	// input.MaxWorkers = aws.Int64(int64(90))
	if v, ok := d.GetOk("max_workers"); ok {
		input.MaxWorkers = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("min_workers"); ok {
		input.MinWorkers = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("plugins_s3_object_version"); ok {
		input.PluginsS3ObjectVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("plugins_s3_path"); ok {
		input.PluginsS3Path = aws.String(v.(string))
	}

	if v, ok := d.GetOk("requirements_s3_object_version"); ok {
		input.RequirementsS3ObjectVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("requirements_s3_path"); ok {
		input.RequirementsS3Path = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedulers"); ok {
		input.Schedulers = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("startup_script_s3_object_version"); ok {
		input.StartupScriptS3ObjectVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("startup_script_s3_path"); ok {
		input.StartupScriptS3Path = aws.String(v.(string))
	}

	if v, ok := d.GetOk("webserver_access_mode"); ok {
		input.WebserverAccessMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("weekly_maintenance_window_start"); ok {
		input.WeeklyMaintenanceWindowStart = aws.String(v.(string))
	}

	/*
		Execution roles created just before the MWAA Environment may result in ValidationExceptions
		due to IAM permission propagation delays.
	*/
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateEnvironmentWithContext(ctx, input)
	}, mwaa.ErrCodeValidationException, mwaa.ErrCodeInternalServerException)

	if err != nil {
		return diag.Errorf("creating MWAA Environment (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitEnvironmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for MWAA Environment (%s) create: %s", d.Id(), err)
	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MWAAConn(ctx)

	environment, err := FindEnvironmentByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MWAA Environment %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading MWAA Environment (%s): %s", d.Id(), err)
	}

	d.Set("airflow_configuration_options", aws.StringValueMap(environment.AirflowConfigurationOptions))
	d.Set("airflow_version", environment.AirflowVersion)
	d.Set("arn", environment.Arn)
	d.Set("created_at", aws.TimeValue(environment.CreatedAt).String())
	d.Set("dag_s3_path", environment.DagS3Path)
	d.Set("environment_class", environment.EnvironmentClass)
	d.Set("execution_role_arn", environment.ExecutionRoleArn)
	d.Set("kms_key", environment.KmsKey)
	if err := d.Set("last_updated", flattenLastUpdate(environment.LastUpdate)); err != nil {
		return diag.Errorf("setting last_updated: %s", err)
	}
	if err := d.Set("logging_configuration", flattenLoggingConfiguration(environment.LoggingConfiguration)); err != nil {
		return diag.Errorf("setting logging_configuration: %s", err)
	}
	d.Set("max_workers", environment.MaxWorkers)
	d.Set("min_workers", environment.MinWorkers)
	d.Set("name", environment.Name)
	if err := d.Set("network_configuration", flattenNetworkConfiguration(environment.NetworkConfiguration)); err != nil {
		return diag.Errorf("setting network_configuration: %s", err)
	}
	d.Set("plugins_s3_object_version", environment.PluginsS3ObjectVersion)
	d.Set("plugins_s3_path", environment.PluginsS3Path)
	d.Set("requirements_s3_object_version", environment.RequirementsS3ObjectVersion)
	d.Set("requirements_s3_path", environment.RequirementsS3Path)
	d.Set("schedulers", environment.Schedulers)
	d.Set("service_role_arn", environment.ServiceRoleArn)
	d.Set("source_bucket_arn", environment.SourceBucketArn)
	d.Set("startup_script_s3_object_version", environment.StartupScriptS3ObjectVersion)
	d.Set("startup_script_s3_path", environment.StartupScriptS3Path)
	d.Set("status", environment.Status)
	d.Set("webserver_access_mode", environment.WebserverAccessMode)
	d.Set("webserver_url", environment.WebserverUrl)
	d.Set("weekly_maintenance_window_start", environment.WeeklyMaintenanceWindowStart)

	setTagsOut(ctx, environment.Tags)

	return nil
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MWAAConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &mwaa.UpdateEnvironmentInput{
			Name: aws.String(d.Get("name").(string)),
		}

		if d.HasChange("airflow_configuration_options") {
			options, ok := d.GetOk("airflow_configuration_options")
			if !ok {
				options = map[string]interface{}{}
			}

			input.AirflowConfigurationOptions = flex.ExpandStringMap(options.(map[string]interface{}))
		}

		if d.HasChange("airflow_version") {
			input.AirflowVersion = aws.String(d.Get("airflow_version").(string))
		}

		if d.HasChange("dag_s3_path") {
			input.DagS3Path = aws.String(d.Get("dag_s3_path").(string))
		}

		if d.HasChange("environment_class") {
			input.EnvironmentClass = aws.String(d.Get("environment_class").(string))
		}

		if d.HasChange("execution_role_arn") {
			input.ExecutionRoleArn = aws.String(d.Get("execution_role_arn").(string))
		}

		if d.HasChange("logging_configuration") {
			input.LoggingConfiguration = expandEnvironmentLoggingConfiguration(d.Get("logging_configuration").([]interface{}))
		}

		if d.HasChange("max_workers") {
			input.MaxWorkers = aws.Int64(int64(d.Get("max_workers").(int)))
		}

		if d.HasChange("min_workers") {
			input.MinWorkers = aws.Int64(int64(d.Get("min_workers").(int)))
		}

		if d.HasChange("network_configuration") {
			input.NetworkConfiguration = expandEnvironmentNetworkConfigurationUpdate(d.Get("network_configuration").([]interface{}))
		}

		if d.HasChange("plugins_s3_object_version") {
			input.PluginsS3ObjectVersion = aws.String(d.Get("plugins_s3_object_version").(string))
		}

		if d.HasChange("plugins_s3_path") {
			input.PluginsS3Path = aws.String(d.Get("plugins_s3_path").(string))
		}

		if d.HasChange("requirements_s3_object_version") {
			input.RequirementsS3ObjectVersion = aws.String(d.Get("requirements_s3_object_version").(string))
		}

		if d.HasChange("requirements_s3_path") {
			input.RequirementsS3Path = aws.String(d.Get("requirements_s3_path").(string))
		}

		if d.HasChange("schedulers") {
			input.Schedulers = aws.Int64(int64(d.Get("schedulers").(int)))
		}

		if d.HasChange("source_bucket_arn") {
			input.SourceBucketArn = aws.String(d.Get("source_bucket_arn").(string))
		}

		if d.HasChange("startup_script_s3_object_version") {
			input.StartupScriptS3ObjectVersion = aws.String(d.Get("startup_script_s3_object_version").(string))
		}

		if d.HasChange("startup_script_s3_path") {
			input.StartupScriptS3Path = aws.String(d.Get("startup_script_s3_path").(string))
		}

		if d.HasChange("webserver_access_mode") {
			input.WebserverAccessMode = aws.String(d.Get("webserver_access_mode").(string))
		}

		if d.HasChange("weekly_maintenance_window_start") {
			input.WeeklyMaintenanceWindowStart = aws.String(d.Get("weekly_maintenance_window_start").(string))
		}

		_, err := conn.UpdateEnvironmentWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MWAA Environment (%s): %s", d.Id(), err)
		}

		if _, err := waitEnvironmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for MWAA Environment (%s) update: %s", d.Id(), err)
		}
	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MWAAConn(ctx)

	log.Printf("[INFO] Deleting MWAA Environment: %s", d.Id())
	_, err := conn.DeleteEnvironmentWithContext(ctx, &mwaa.DeleteEnvironmentInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, mwaa.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting MWAA Environment (%s): %s", d.Id(), err)
	}

	if _, err := waitEnvironmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for MWAA Environment (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func environmentModuleLoggingConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cloud_watch_log_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"log_level": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(mwaa.LoggingLevel_Values(), false),
			},
		},
	}
}

func FindEnvironmentByName(ctx context.Context, conn *mwaa.MWAA, name string) (*mwaa.Environment, error) {
	input := &mwaa.GetEnvironmentInput{
		Name: aws.String(name),
	}

	output, err := conn.GetEnvironmentWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, mwaa.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Environment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Environment, nil
}

func statusEnvironment(ctx context.Context, conn *mwaa.MWAA, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		environment, err := FindEnvironmentByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return environment, aws.StringValue(environment.Status), nil
	}
}

func waitEnvironmentCreated(ctx context.Context, conn *mwaa.MWAA, name string, timeout time.Duration) (*mwaa.Environment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{mwaa.EnvironmentStatusCreating},
		Target:  []string{mwaa.EnvironmentStatusAvailable},
		Refresh: statusEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*mwaa.Environment); ok {
		if v.LastUpdate != nil && v.LastUpdate.Error != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(v.LastUpdate.Error.ErrorCode), aws.StringValue(v.LastUpdate.Error.ErrorMessage)))
		}

		return v, err
	}

	return nil, err
}

func waitEnvironmentUpdated(ctx context.Context, conn *mwaa.MWAA, name string, timeout time.Duration) (*mwaa.Environment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{mwaa.EnvironmentStatusUpdating, mwaa.EnvironmentStatusCreatingSnapshot},
		Target:  []string{mwaa.EnvironmentStatusAvailable},
		Refresh: statusEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*mwaa.Environment); ok {
		if v.LastUpdate != nil && v.LastUpdate.Error != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(v.LastUpdate.Error.ErrorCode), aws.StringValue(v.LastUpdate.Error.ErrorMessage)))
		}

		return v, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *mwaa.MWAA, name string, timeout time.Duration) (*mwaa.Environment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{mwaa.EnvironmentStatusDeleting},
		Target:  []string{},
		Refresh: statusEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*mwaa.Environment); ok {
		if v.LastUpdate != nil && v.LastUpdate.Error != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(v.LastUpdate.Error.ErrorCode), aws.StringValue(v.LastUpdate.Error.ErrorMessage)))
		}

		return v, err
	}

	return nil, err
}

func expandEnvironmentLoggingConfiguration(l []interface{}) *mwaa.LoggingConfigurationInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	input := &mwaa.LoggingConfigurationInput{}

	m := l[0].(map[string]interface{})

	if v, ok := m["dag_processing_logs"]; ok {
		input.DagProcessingLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["scheduler_logs"]; ok {
		input.SchedulerLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["task_logs"]; ok {
		input.TaskLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["webserver_logs"]; ok {
		input.WebserverLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["worker_logs"]; ok {
		input.WorkerLogs = expandEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	return input
}

func expandEnvironmentModuleLoggingConfiguration(l []interface{}) *mwaa.ModuleLoggingConfigurationInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	input := &mwaa.ModuleLoggingConfigurationInput{}
	m := l[0].(map[string]interface{})

	input.Enabled = aws.Bool(m["enabled"].(bool))
	input.LogLevel = aws.String(m["log_level"].(string))

	return input
}

func expandEnvironmentNetworkConfigurationCreate(l []interface{}) *mwaa.NetworkConfiguration {
	m := l[0].(map[string]interface{})

	return &mwaa.NetworkConfiguration{
		SecurityGroupIds: flex.ExpandStringSet(m["security_group_ids"].(*schema.Set)),
		SubnetIds:        flex.ExpandStringSet(m["subnet_ids"].(*schema.Set)),
	}
}

func expandEnvironmentNetworkConfigurationUpdate(l []interface{}) *mwaa.UpdateNetworkConfigurationInput {
	m := l[0].(map[string]interface{})

	return &mwaa.UpdateNetworkConfigurationInput{
		SecurityGroupIds: flex.ExpandStringSet(m["security_group_ids"].(*schema.Set)),
	}
}

func flattenLastUpdate(lastUpdate *mwaa.LastUpdate) []interface{} {
	if lastUpdate == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if lastUpdate.CreatedAt != nil {
		m["created_at"] = aws.TimeValue(lastUpdate.CreatedAt).String()
	}

	if lastUpdate.Error != nil {
		m["error"] = flattenLastUpdateError(lastUpdate.Error)
	}

	if lastUpdate.Status != nil {
		m["status"] = lastUpdate.Status
	}

	return []interface{}{m}
}

func flattenLastUpdateError(apiObject *mwaa.UpdateError) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if apiObject.ErrorCode != nil {
		m["error_code"] = apiObject.ErrorCode
	}

	if apiObject.ErrorMessage != nil {
		m["error_message"] = apiObject.ErrorMessage
	}

	return []interface{}{m}
}

func flattenLoggingConfiguration(loggingConfiguration *mwaa.LoggingConfiguration) []interface{} {
	if loggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if loggingConfiguration.DagProcessingLogs != nil {
		m["dag_processing_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.DagProcessingLogs)
	}

	if loggingConfiguration.SchedulerLogs != nil {
		m["scheduler_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.SchedulerLogs)
	}

	if loggingConfiguration.TaskLogs != nil {
		m["task_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.TaskLogs)
	}

	if loggingConfiguration.WebserverLogs != nil {
		m["webserver_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.WebserverLogs)
	}

	if loggingConfiguration.WorkerLogs != nil {
		m["worker_logs"] = flattenModuleLoggingConfiguration(loggingConfiguration.WorkerLogs)
	}

	return []interface{}{m}
}

func flattenModuleLoggingConfiguration(moduleLoggingConfiguration *mwaa.ModuleLoggingConfiguration) []interface{} {
	if moduleLoggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloud_watch_log_group_arn": aws.StringValue(moduleLoggingConfiguration.CloudWatchLogGroupArn),
		"enabled":                   aws.BoolValue(moduleLoggingConfiguration.Enabled),
		"log_level":                 aws.StringValue(moduleLoggingConfiguration.LogLevel),
	}

	return []interface{}{m}
}

func flattenNetworkConfiguration(networkConfiguration *mwaa.NetworkConfiguration) []interface{} {
	if networkConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"security_group_ids": flex.FlattenStringSet(networkConfiguration.SecurityGroupIds),
		"subnet_ids":         flex.FlattenStringSet(networkConfiguration.SubnetIds),
	}

	return []interface{}{m}
}
