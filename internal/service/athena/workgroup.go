// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package athena

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcty "github.com/hashicorp/terraform-provider-aws/internal/cty"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_athena_workgroup", name="WorkGroup")
// @Tags(identifierAttribute="arn")
func resourceWorkGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkGroupCreate,
		ReadWithoutTimeout:   resourceWorkGroupRead,
		UpdateWithoutTimeout: resourceWorkGroupUpdate,
		DeleteWithoutTimeout: resourceWorkGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: managedQueryResultsValidation,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConfiguration: {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bytes_scanned_cutoff_per_query": {
							Type:     schema.TypeInt,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.IntAtLeast(10485760),
								validation.IntInSlice([]int{0}),
							),
						},
						"customer_content_encryption_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKMSKey: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[a-z\-]*:kms:([a-z0-9\-]+):\d{12}:key/?[a-zA-Z_0-9+=,.@\-_/]+$|^arn:aws[a-z\-]*:kms:([a-z0-9\-]+):\d{12}:alias/?[a-zA-Z_0-9+=,.@\-_/]+$|^alias/[a-zA-Z0-9/_-]+$|[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`), "must be a valid KMS Key ARN or Alias"),
									},
								},
							},
						},
						"enable_minimum_encryption_configuration": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"enforce_workgroup_configuration": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						names.AttrEngineVersion: {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"effective_engine_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"selected_engine_version": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "AUTO",
									},
								},
							},
						},
						"execution_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"identity_center_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable_identity_center": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"identity_center_instance_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"query_results_s3_access_grants_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.AuthenticationType](),
									},
									"create_user_level_prefix": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"enable_s3_access_grants": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"managed_query_results_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Optional: true,
									},
									names.AttrEncryptionConfiguration: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrKMSKey: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
								},
							},
						},
						"monitoring_configuration": {
							Type:             schema.TypeList,
							Optional:         true,
							MaxItems:         1,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cloud_watch_logging_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
											// Empty block and enabled = false is equivalent
											o, n := d.GetChange("configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration")
											if o != nil && n != nil {
												if v, ok := o.([]any); ok && len(v) == 0 {
													if v, ok := n.([]any); ok && len(v) > 0 && v[0] != nil {
														vm := v[0].(map[string]any)
														if v, ok := vm[names.AttrEnabled].(bool); ok && !v {
															return true
														}
													}
												}
											}
											return false
										},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Required: true,
													DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
														if old == "" && new == "false" {
															return true
														}
														return false
													},
												},
												"log_group": {
													Type:             schema.TypeString,
													Optional:         true,
													DiffSuppressFunc: diffSuppressWorkGroupConfigurationMonitoringCloudWatchLogging,
													ValidateFunc: validation.All(
														validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9._/-]+$`), "must contain only alphanumeric characters, periods, underscores, hyphens, and slashes"),
														validation.StringLenBetween(1, 512),
													),
												},
												"log_stream_name_prefix": {
													Type:             schema.TypeString,
													Optional:         true,
													DiffSuppressFunc: diffSuppressWorkGroupConfigurationMonitoringCloudWatchLogging,
													ValidateFunc: validation.All(
														validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9._/-]+$`), "must contain only alphanumeric characters, periods, underscores, hyphens, and slashes"),
														validation.StringLenBetween(1, 512),
													),
												},
												"log_type": {
													Type:             schema.TypeSet,
													Optional:         true,
													DiffSuppressFunc: diffSuppressWorkGroupConfigurationMonitoringCloudWatchLogging,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrKey: {
																Type:     schema.TypeString,
																Required: true,
															},
															names.AttrValues: {
																Type:     schema.TypeSet,
																Required: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
														},
													},
												},
											},
										},
									},
									"managed_logging_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
											// Empty block and enabled = false is equivalent
											o, n := d.GetChange("configuration.0.monitoring_configuration.0.managed_logging_configuration")
											if o != nil && n != nil {
												if v, ok := o.([]any); ok && len(v) == 0 {
													if v, ok := n.([]any); ok && len(v) > 0 && v[0] != nil {
														vm := v[0].(map[string]any)
														if v, ok := vm[names.AttrEnabled].(bool); ok && !v {
															return true
														}
													}
												}
											}
											return false
										},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Required: true,
													DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
														if old == "" && new == "false" {
															return true
														}
														return false
													},
												},
												names.AttrKMSKey: {
													Type:             schema.TypeString,
													Optional:         true,
													DiffSuppressFunc: diffSuppressWorkGroupConfigurationMonitoringManagedLogging,
													ValidateFunc:     validation.StringMatch(regexache.MustCompile(`^arn:aws[a-z\-]*:kms:([a-z0-9\-]+):\d{12}:key/?[a-zA-Z_0-9+=,.@\-_/]+$|^arn:aws[a-z\-]*:kms:([a-z0-9\-]+):\d{12}:alias/?[a-zA-Z_0-9+=,.@\-_/]+$|^alias/[a-zA-Z0-9/_-]+$|[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`), "must be a valid KMS Key ARN or Alias"),
												},
											},
										},
									},
									"s3_logging_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
											// Empty block and enabled = false is equivalent
											o, n := d.GetChange("configuration.0.monitoring_configuration.0.s3_logging_configuration")
											if o != nil && n != nil {
												if v, ok := o.([]any); ok && len(v) == 0 {
													if v, ok := n.([]any); ok && len(v) > 0 && v[0] != nil {
														vm := v[0].(map[string]any)
														if v, ok := vm[names.AttrEnabled].(bool); ok && !v {
															return true
														}
													}
												}
											}
											return false
										},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Required: true,
													DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
														if old == "" && new == "false" {
															return true
														}
														return false
													},
												},
												names.AttrKMSKey: {
													Type:             schema.TypeString,
													Optional:         true,
													DiffSuppressFunc: diffSuppressWorkGroupConfigurationMonitoringS3Logging,
													ValidateFunc:     validation.StringMatch(regexache.MustCompile(`^arn:aws[a-z\-]*:kms:([a-z0-9\-]+):\d{12}:key/?[a-zA-Z_0-9+=,.@\-_/]+$|^arn:aws[a-z\-]*:kms:([a-z0-9\-]+):\d{12}:alias/?[a-zA-Z_0-9+=,.@\-_/]+$|^alias/[a-zA-Z0-9/_-]+$|[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`), "must be a valid KMS Key ARN or Alias"),
												},
												"log_location": {
													Type:             schema.TypeString,
													Optional:         true,
													DiffSuppressFunc: diffSuppressWorkGroupConfigurationMonitoringS3Logging,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 1024),
														validation.StringMatch(regexache.MustCompile(`^s3://[a-z0-9][a-z0-9\-]*[a-z0-9](/.*)?$`), "must be a valid S3 URI starting with s3://"),
													),
												},
											},
										},
									},
								},
							},
						},
						"publish_cloudwatch_metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"result_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"acl_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_acl_option": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.S3AclOption](),
												},
											},
										},
									},
									names.AttrEncryptionConfiguration: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_option": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.EncryptionOption](),
												},
												names.AttrKMSKeyARN: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									names.AttrExpectedBucketOwner: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"output_location": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"requester_pays_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must contain only alphanumeric characters, periods, underscores, and hyphens"),
				),
			},
			names.AttrState: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.WorkGroupStateEnabled,
				ValidateDiagFunc: enum.Validate[types.WorkGroupState](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func diffSuppressWorkGroupConfigurationMonitoringCloudWatchLogging(_, old, new string, d *schema.ResourceData) bool {
	// cloud_watch_logging_configuration is disabled, ignore other attribute diffs
	if _, ok := d.GetOk("configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration.0.enabled"); !ok {
		return true
	}
	return false
}

func diffSuppressWorkGroupConfigurationMonitoringManagedLogging(_, old, new string, d *schema.ResourceData) bool {
	// managed_logging_configuration is disabled, ignore other attribute diffs
	if _, ok := d.GetOk("configuration.0.monitoring_configuration.0.managed_logging_configuration.0.enabled"); !ok {
		return true
	}
	return false
}

func diffSuppressWorkGroupConfigurationMonitoringS3Logging(_, old, new string, d *schema.ResourceData) bool {
	// s3_logging_configuration is disabled, ignore other attribute diffs
	if _, ok := d.GetOk("configuration.0.monitoring_configuration.0.s3_logging_configuration.0.enabled"); !ok {
		return true
	}
	return false
}

func resourceWorkGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := athena.CreateWorkGroupInput{
		Configuration: expandWorkGroupConfiguration(d.Get(names.AttrConfiguration).([]any)),
		Name:          aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	// To avoid eventual consistency issues with IAM roles,
	// retry on InvalidRequestException with a specific message.
	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *types.InvalidRequestException](ctx, timeout,
		func(ctx context.Context) (any, error) {
			return conn.CreateWorkGroup(ctx, &input)
		},
		"Make sure that athena.amazonaws.com has been allowed for sts:AssumeRole",
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Athena WorkGroup (%s): %s", name, err)
	}

	d.SetId(name)

	if v := types.WorkGroupState(d.Get(names.AttrState).(string)); v == types.WorkGroupStateDisabled {
		input := athena.UpdateWorkGroupInput{
			State:     v,
			WorkGroup: aws.String(d.Id()),
		}

		_, err := conn.UpdateWorkGroup(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkGroupRead(ctx, d, meta)...)
}

func resourceWorkGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.AthenaClient(ctx)

	wg, err := findWorkGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Athena WorkGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Athena WorkGroup (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, workGroupARN(ctx, c, d.Id()))
	if err := d.Set(names.AttrConfiguration, flattenWorkGroupConfiguration(wg.Configuration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}
	d.Set(names.AttrDescription, wg.Description)
	d.Set(names.AttrForceDestroy, d.Get(names.AttrForceDestroy))
	d.Set(names.AttrName, wg.Name)
	d.Set(names.AttrState, wg.State)

	return diags
}

func resourceWorkGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := athena.UpdateWorkGroupInput{
			WorkGroup: aws.String(d.Get(names.AttrName).(string)),
		}

		if d.HasChange(names.AttrConfiguration) {
			input.ConfigurationUpdates = expandWorkGroupConfigurationUpdates(d.Get(names.AttrConfiguration).([]any))
			if d.HasChange("configuration.0.customer_content_encryption_configuration") {
				// When customer_content_encryption_configuration is changed and the expander returns nil,
				// the content_encryption_configuration needs to be removed.
				// To remove it, set RemoveCustomerContentEncryptionConfiguration to true.
				if input.ConfigurationUpdates == nil || input.ConfigurationUpdates.CustomerContentEncryptionConfiguration == nil {
					input.ConfigurationUpdates.RemoveCustomerContentEncryptionConfiguration = aws.Bool(true)
				}
			}
			if d.HasChange("configuration.0.monitoring_configuration.0.cloud_watch_logging_configuration") {
				// When monitoring_configuration block is removed or monitoring_configuration.cloud_watch_logging_configuration is removed,
				// the cloud_watch_logging_configuration is disabled.
				if input.ConfigurationUpdates == nil || (input.ConfigurationUpdates.MonitoringConfiguration == nil || input.ConfigurationUpdates.MonitoringConfiguration.CloudWatchLoggingConfiguration == nil) {
					input.ConfigurationUpdates.MonitoringConfiguration.CloudWatchLoggingConfiguration.Enabled = aws.Bool(false)
				}
			}
			if d.HasChange("configuration.0.monitoring_configuration.0.managed_logging_configuration") {
				// When monitoring_configuration block is removed or monitoring_configuration.managed_logging_configuration is removed,
				// the managed_logging_configuration is disabled.
				if input.ConfigurationUpdates == nil || (input.ConfigurationUpdates.MonitoringConfiguration == nil || input.ConfigurationUpdates.MonitoringConfiguration.ManagedLoggingConfiguration == nil) {
					input.ConfigurationUpdates.MonitoringConfiguration.ManagedLoggingConfiguration.Enabled = aws.Bool(false)
				}
			}
			if d.HasChange("configuration.0.monitoring_configuration.0.s3_logging_configuration") {
				// When monitoring_configuration block is removed or monitoring_configuration.s3_logging_configuration is removed,
				// the s3_logging_configuration is disabled.
				if input.ConfigurationUpdates == nil || (input.ConfigurationUpdates.MonitoringConfiguration == nil || input.ConfigurationUpdates.MonitoringConfiguration.S3LoggingConfiguration == nil) {
					input.ConfigurationUpdates.MonitoringConfiguration.S3LoggingConfiguration.Enabled = aws.Bool(false)
				}
			}
			if d.HasChange("configuration.0.requester_pays_enabled") {
				// When requester_pays_enabled is returned as nil, set it to false to disable it.
				if input.ConfigurationUpdates == nil || input.ConfigurationUpdates.RequesterPaysEnabled == nil {
					input.ConfigurationUpdates.RequesterPaysEnabled = aws.Bool(false)
				}
			}
			if d.HasChange("configuration.0.enable_minimum_encryption_configuration") {
				// When enable_minimum_encryption_configuration is returned as nil, set it to false to disable it.
				if input.ConfigurationUpdates == nil || input.ConfigurationUpdates.EnableMinimumEncryptionConfiguration == nil {
					input.ConfigurationUpdates.EnableMinimumEncryptionConfiguration = aws.Bool(false)
				}
			}

			if d.HasChanges("configuration.0.result_configuration") {
				path := cty.GetAttrPath(names.AttrConfiguration).IndexInt(0).GetAttr("result_configuration").IndexInt(0)

				rs := d.GetRawState()
				stateResultConfiguration, _, err := tfcty.PathSafeApply(path, rs)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading state value at %q: %s", errs.PathString(path), err)
				}

				rc := d.GetRawConfig()
				configResultConfiguration, _, err := tfcty.PathSafeApply(path, rc)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading config value at %q: %s", errs.PathString(path), err)
				}

				input.ConfigurationUpdates.ResultConfigurationUpdates, err = expandWorkGroupResultConfigurationUpdatesByDiff(stateResultConfiguration, configResultConfiguration)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "expanding result configuration updates: %s", err)
				}
			}
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrState) {
			input.State = types.WorkGroupState(d.Get(names.AttrState).(string))
		}

		// To avoid eventual consistency issues with IAM roles,
		// retry on InvalidRequestException with a specific message.
		const (
			timeout = 2 * time.Minute
		)
		_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *types.InvalidRequestException](ctx, timeout,
			func(ctx context.Context) (any, error) {
				return conn.UpdateWorkGroup(ctx, &input)
			},
			"Make sure that athena.amazonaws.com has been allowed for sts:AssumeRole",
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkGroupRead(ctx, d, meta)...)
}

func resourceWorkGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	log.Printf("[DEBUG] Deleting Athena WorkGroup (%s)", d.Id())
	input := athena.DeleteWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}
	if v, ok := d.GetOk(names.AttrForceDestroy); ok {
		input.RecursiveDeleteOption = aws.Bool(v.(bool))
	}
	_, err := conn.DeleteWorkGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Athena WorkGroup (%s): %s", d.Id(), err)
	}

	return diags
}

func findWorkGroupByName(ctx context.Context, conn *athena.Client, name string) (*types.WorkGroup, error) {
	input := athena.GetWorkGroupInput{
		WorkGroup: aws.String(name),
	}

	output, err := conn.GetWorkGroup(ctx, &input)

	if errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "is not found") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.WorkGroup == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.WorkGroup, nil
}

func expandWorkGroupConfiguration(l []any) *types.WorkGroupConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	configuration := &types.WorkGroupConfiguration{}

	if v, ok := m["bytes_scanned_cutoff_per_query"].(int); ok && v > 0 {
		configuration.BytesScannedCutoffPerQuery = aws.Int64(int64(v))
	}

	if v, ok := m["customer_content_encryption_configuration"]; ok {
		configuration.CustomerContentEncryptionConfiguration = expandWorkGroupCustomerContentEncryptionConfiguration(v.([]any))
	}

	// Depending on other configurations, enable_minimum_encryption_configuration
	// must not be specified, even when set to false.
	// Therefore, the value is set only when it is true to avoid an API error.
	if v, ok := m["enable_minimum_encryption_configuration"].(bool); ok && v {
		configuration.EnableMinimumEncryptionConfiguration = aws.Bool(v)
	}

	if v, ok := m["enforce_workgroup_configuration"].(bool); ok {
		configuration.EnforceWorkGroupConfiguration = aws.Bool(v)
	}

	if v, ok := m[names.AttrEngineVersion].([]any); ok && len(v) > 0 && v[0] != nil {
		configuration.EngineVersion = expandWorkGroupEngineVersion(v)
	}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		configuration.ExecutionRole = aws.String(v)
	}

	if v, ok := m["identity_center_configuration"]; ok {
		configuration.IdentityCenterConfiguration = expandWorkGroupIdentityCenterConfiguration(v.([]any))
	}

	if v, ok := m["query_results_s3_access_grants_configuration"]; ok {
		configuration.QueryResultsS3AccessGrantsConfiguration = expandWorkGroupQueryResultsS3AccessGrantsConfiguration(v.([]any))
	}

	if v, ok := m["managed_query_results_configuration"]; ok {
		configuration.ManagedQueryResultsConfiguration = expandWorkGroupManagedQueryResultsConfiguration(v.([]any))
	}

	if v, ok := m["monitoring_configuration"]; ok {
		configuration.MonitoringConfiguration = expandWorkGroupMonitoringConfiguration(v.([]any))
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"].(bool); ok {
		configuration.PublishCloudWatchMetricsEnabled = aws.Bool(v)
	}

	if v, ok := m["result_configuration"]; ok {
		configuration.ResultConfiguration = expandWorkGroupResultConfiguration(v.([]any))
	}

	// Depending on other configurations, requester_pays_enabled
	// must not be specified, even when set to false.
	// Therefore, the value is set only when it is true to avoid an API error.
	if v, ok := m["requester_pays_enabled"].(bool); ok && v {
		configuration.RequesterPaysEnabled = aws.Bool(v)
	}

	return configuration
}

func expandWorkGroupCustomerContentEncryptionConfiguration(l []any) *types.CustomerContentEncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	// customerContentEncryptionConfiguration is created and returned
	// only when a KMS key is specified.
	// Otherwise, nil is returned to avoid an SDK error.
	//
	// When the KMS key is removed from the configuration during an update,
	// the returned nil is handled in the Update function via
	// RemoveCustomerContentEncryptionConfiguration.
	if v, ok := m[names.AttrKMSKey].(string); ok && v != "" {
		customerContentEncryptionConfiguration := &types.CustomerContentEncryptionConfiguration{
			KmsKey: aws.String(v),
		}
		return customerContentEncryptionConfiguration
	}
	return nil
}

func expandWorkGroupEngineVersion(l []any) *types.EngineVersion {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	engineVersion := &types.EngineVersion{}

	if v, ok := m["selected_engine_version"].(string); ok && v != "" {
		engineVersion.SelectedEngineVersion = aws.String(v)
	}

	return engineVersion
}

func expandWorkGroupConfigurationUpdates(l []any) *types.WorkGroupConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	configurationUpdates := &types.WorkGroupConfigurationUpdates{}

	if v, ok := m["bytes_scanned_cutoff_per_query"].(int); ok && v > 0 {
		configurationUpdates.BytesScannedCutoffPerQuery = aws.Int64(int64(v))
	} else {
		configurationUpdates.RemoveBytesScannedCutoffPerQuery = aws.Bool(true)
	}

	if v, ok := m["customer_content_encryption_configuration"]; ok {
		configurationUpdates.CustomerContentEncryptionConfiguration = expandWorkGroupCustomerContentEncryptionConfiguration(v.([]any))
	}

	// Depending on other configurations, enable_minimum_encryption_configuration
	// must not be specified, even when set to false.
	// Therefore, the value is set only when it is true to avoid an API error.
	// The returned nil is handled in the Update function when this setting
	// needs to be disabled.
	if v, ok := m["enable_minimum_encryption_configuration"].(bool); ok && v {
		configurationUpdates.EnableMinimumEncryptionConfiguration = aws.Bool(v)
	}

	if v, ok := m["enforce_workgroup_configuration"].(bool); ok {
		configurationUpdates.EnforceWorkGroupConfiguration = aws.Bool(v)
	}

	if v, ok := m[names.AttrEngineVersion].([]any); ok && len(v) > 0 && v[0] != nil {
		configurationUpdates.EngineVersion = expandWorkGroupEngineVersion(v)
	}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		configurationUpdates.ExecutionRole = aws.String(v)
	}

	if v, ok := m["query_results_s3_access_grants_configuration"]; ok {
		configurationUpdates.QueryResultsS3AccessGrantsConfiguration = expandWorkGroupQueryResultsS3AccessGrantsConfiguration(v.([]any))
	}

	if v, ok := m["managed_query_results_configuration"]; ok {
		configurationUpdates.ManagedQueryResultsConfigurationUpdates = expandWorkGroupManagedQueryResultsConfigurationUpdates(v.([]any))
	}

	if v, ok := m["monitoring_configuration"]; ok {
		configurationUpdates.MonitoringConfiguration = expandWorkGroupMonitoringConfiguration(v.([]any))
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"].(bool); ok {
		configurationUpdates.PublishCloudWatchMetricsEnabled = aws.Bool(v)
	}

	// Depending on other configurations, requester_pays_enabled
	// must not be specified, even when set to false.
	// Therefore, the value is set only when it is true to avoid an API error.
	// The returned nil is handled in the Update function when this setting
	// needs to be disabled.
	if v, ok := m["requester_pays_enabled"].(bool); ok && v {
		configurationUpdates.RequesterPaysEnabled = aws.Bool(v)
	}

	return configurationUpdates
}

func expandWorkGroupResultConfigurationUpdatesByDiff(state, config cty.Value) (*types.ResultConfigurationUpdates, error) {
	stateHasValue := tfcty.HasValue(state)
	configHasValue := tfcty.HasValue(config)

	if stateHasValue && !configHasValue {
		// Removing `result_configuration`
		// Set all `Remove` fields for simplcity
		return &types.ResultConfigurationUpdates{
			RemoveAclConfiguration:        aws.Bool(true),
			RemoveEncryptionConfiguration: aws.Bool(true),
			RemoveExpectedBucketOwner:     aws.Bool(true),
			RemoveOutputLocation:          aws.Bool(true),
		}, nil
	} else if configHasValue {
		result := &types.ResultConfigurationUpdates{}

		aclPath := cty.GetAttrPath("acl_configuration").IndexInt(0)
		stateACLConfiguration, _, err := tfcty.PathSafeApply(aclPath, state)
		if err != nil {
			return nil, fmt.Errorf("reading state value at %q: %w", errs.PathString(aclPath), err)
		}
		configACLConfiguration, _, err := tfcty.PathSafeApply(aclPath, config)
		if err != nil {
			return nil, fmt.Errorf("reading config value at %q: %w", errs.PathString(aclPath), err)
		}
		expandWorkGroupACLConfigurationByDiff(stateACLConfiguration, configACLConfiguration, result)

		encryptionPath := cty.GetAttrPath(names.AttrEncryptionConfiguration).IndexInt(0)
		stateEncryptionConfiguration, _, err := tfcty.PathSafeApply(encryptionPath, state)
		if err != nil {
			return nil, fmt.Errorf("reading state value at %q: %w", errs.PathString(encryptionPath), err)
		}
		configEncryptionConfiguration, _, err := tfcty.PathSafeApply(encryptionPath, config)
		if err != nil {
			return nil, fmt.Errorf("reading config value at %q: %w", errs.PathString(encryptionPath), err)
		}
		expandWorkGroupEncryptionConfigurationByDiff(stateEncryptionConfiguration, configEncryptionConfiguration, result)

		if outputLocation := config.GetAttr("output_location"); outputLocation.IsNull() {
			result.RemoveOutputLocation = aws.Bool(true)
		} else {
			result.OutputLocation = aws.String(outputLocation.AsString())
		}

		if expectedBucketOwner := config.GetAttr(names.AttrExpectedBucketOwner); expectedBucketOwner.IsNull() {
			result.RemoveExpectedBucketOwner = aws.Bool(true)
		} else {
			result.ExpectedBucketOwner = aws.String(expectedBucketOwner.AsString())
		}

		return result, nil
	}

	return nil, nil
}

func expandWorkGroupIdentityCenterConfiguration(l []any) *types.IdentityCenterConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	identityCenterConfiguration := &types.IdentityCenterConfiguration{}

	if v, ok := m["enable_identity_center"].(bool); ok {
		identityCenterConfiguration.EnableIdentityCenter = aws.Bool(v)
	}

	if v, ok := m["identity_center_instance_arn"].(string); ok && v != "" {
		identityCenterConfiguration.IdentityCenterInstanceArn = aws.String(v)
	}

	return identityCenterConfiguration
}

func expandWorkGroupQueryResultsS3AccessGrantsConfiguration(l []any) *types.QueryResultsS3AccessGrantsConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	queryResultsS3AccessGrantsConfiguration := &types.QueryResultsS3AccessGrantsConfiguration{}

	if v, ok := m["authentication_type"].(string); ok && v != "" {
		queryResultsS3AccessGrantsConfiguration.AuthenticationType = types.AuthenticationType(v)
	}

	if v, ok := m["create_user_level_prefix"].(bool); ok {
		queryResultsS3AccessGrantsConfiguration.CreateUserLevelPrefix = aws.Bool(v)
	}

	if v, ok := m["enable_s3_access_grants"].(bool); ok {
		queryResultsS3AccessGrantsConfiguration.EnableS3AccessGrants = aws.Bool(v)
	}

	return queryResultsS3AccessGrantsConfiguration
}

func expandWorkGroupResultConfiguration(l []any) *types.ResultConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	resultConfiguration := &types.ResultConfiguration{}

	if v, ok := m[names.AttrEncryptionConfiguration]; ok {
		resultConfiguration.EncryptionConfiguration = expandWorkGroupEncryptionConfiguration(v.([]any))
	}

	if v, ok := m["output_location"].(string); ok && v != "" {
		resultConfiguration.OutputLocation = aws.String(v)
	}

	if v, ok := m[names.AttrExpectedBucketOwner].(string); ok && v != "" {
		resultConfiguration.ExpectedBucketOwner = aws.String(v)
	}

	if v, ok := m["acl_configuration"]; ok {
		resultConfiguration.AclConfiguration = expandResultConfigurationACLConfig(v.([]any))
	}

	return resultConfiguration
}

func expandWorkGroupMonitoringConfiguration(l []any) *types.MonitoringConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	monitoringConfiguration := &types.MonitoringConfiguration{}

	if v, ok := m["cloud_watch_logging_configuration"]; ok {
		monitoringConfiguration.CloudWatchLoggingConfiguration = expandWorkGroupMonitoringConfigurationCloudWatchLoggingConfiguration(v.([]any))
	}

	if v, ok := m["managed_logging_configuration"]; ok {
		monitoringConfiguration.ManagedLoggingConfiguration = expandWorkGroupMonitoringConfigurationManagedLoggingConfiguration(v.([]any))
	}

	if v, ok := m["s3_logging_configuration"]; ok {
		monitoringConfiguration.S3LoggingConfiguration = expandWorkGroupMonitoringConfigurationS3LoggingConfiguration(v.([]any))
	}

	return monitoringConfiguration
}

func expandWorkGroupMonitoringConfigurationCloudWatchLoggingConfiguration(l []any) *types.CloudWatchLoggingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	cloudWatchLoggingConfiguration := &types.CloudWatchLoggingConfiguration{}

	if v, ok := m[names.AttrEnabled].(bool); ok {
		cloudWatchLoggingConfiguration.Enabled = aws.Bool(v)
	}

	if v, ok := m["log_group"].(string); ok && v != "" {
		cloudWatchLoggingConfiguration.LogGroup = aws.String(v)
	}

	if v, ok := m["log_stream_name_prefix"].(string); ok && v != "" {
		cloudWatchLoggingConfiguration.LogStreamNamePrefix = aws.String(v)
	}

	if v, ok := m["log_type"].(*schema.Set); ok && v.Len() > 0 {
		logTypes := make(map[string][]string)
		for _, item := range v.List() {
			m := item.(map[string]any)
			if key, ok := m[names.AttrKey].(string); ok && key != "" {
				if valueSet, ok := m[names.AttrValues].(*schema.Set); ok {
					values := make([]string, 0, valueSet.Len())
					for _, val := range valueSet.List() {
						values = append(values, val.(string))
					}
					logTypes[key] = values
				}
			}
		}
		cloudWatchLoggingConfiguration.LogTypes = logTypes
	}

	return cloudWatchLoggingConfiguration
}

func expandWorkGroupMonitoringConfigurationManagedLoggingConfiguration(l []any) *types.ManagedLoggingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	managedLoggingConfiguration := &types.ManagedLoggingConfiguration{}

	if v, ok := m[names.AttrEnabled].(bool); ok {
		managedLoggingConfiguration.Enabled = aws.Bool(v)
	}
	if v, ok := m[names.AttrKMSKey].(string); ok && v != "" {
		managedLoggingConfiguration.KmsKey = aws.String(v)
	}
	return managedLoggingConfiguration
}

func expandWorkGroupMonitoringConfigurationS3LoggingConfiguration(l []any) *types.S3LoggingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	s3LoggingConfiguration := &types.S3LoggingConfiguration{}

	if v, ok := m[names.AttrEnabled].(bool); ok {
		s3LoggingConfiguration.Enabled = aws.Bool(v)
	}
	if v, ok := m[names.AttrKMSKey].(string); ok && v != "" {
		s3LoggingConfiguration.KmsKey = aws.String(v)
	}
	if v, ok := m["log_location"].(string); ok && v != "" {
		s3LoggingConfiguration.LogLocation = aws.String(v)
	}
	return s3LoggingConfiguration
}

func expandWorkGroupEncryptionConfiguration(l []any) *types.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	encryptionConfiguration := types.EncryptionConfiguration{}

	if v, ok := m["encryption_option"]; ok && v.(string) != "" {
		encryptionConfiguration.EncryptionOption = types.EncryptionOption(v.(string))
	}

	if v, ok := m[names.AttrKMSKeyARN]; ok && v.(string) != "" {
		encryptionConfiguration.KmsKey = aws.String(v.(string))
	}

	return &encryptionConfiguration
}

func expandWorkGroupACLConfigurationByDiff(state, config cty.Value, result *types.ResultConfigurationUpdates) {
	stateHasValue := tfcty.HasValue(state)
	configHasValue := tfcty.HasValue(config)

	if stateHasValue && !configHasValue {
		result.RemoveAclConfiguration = aws.Bool(true)
		return
	}

	if !configHasValue {
		return
	}

	aclConfiguration := types.AclConfiguration{}

	if v := config.GetAttr("s3_acl_option"); !v.IsNull() && v.AsString() != "" {
		aclConfiguration.S3AclOption = types.S3AclOption(v.AsString())
	}

	result.AclConfiguration = &aclConfiguration
}

func expandWorkGroupEncryptionConfigurationByDiff(state, config cty.Value, result *types.ResultConfigurationUpdates) {
	stateHasValue := tfcty.HasValue(state)
	configHasValue := tfcty.HasValue(config)

	if stateHasValue && !configHasValue {
		result.RemoveEncryptionConfiguration = aws.Bool(true)
		return
	}

	if !configHasValue {
		return
	}

	encryptionConfiguration := types.EncryptionConfiguration{}

	if v := config.GetAttr("encryption_option"); !v.IsNull() && v.AsString() != "" {
		encryptionConfiguration.EncryptionOption = types.EncryptionOption(v.AsString())
	}

	if v := config.GetAttr(names.AttrKMSKeyARN); !v.IsNull() && v.AsString() != "" {
		encryptionConfiguration.KmsKey = aws.String(v.AsString())
	}

	result.EncryptionConfiguration = &encryptionConfiguration
}

func expandWorkGroupManagedQueryResultsConfiguration(l []any) *types.ManagedQueryResultsConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	managedQueryResultsConfiguration := &types.ManagedQueryResultsConfiguration{}

	if v, ok := m[names.AttrEnabled].(bool); ok {
		managedQueryResultsConfiguration.Enabled = v
	}

	if v, ok := m[names.AttrEncryptionConfiguration]; ok {
		managedQueryResultsConfiguration.EncryptionConfiguration = expandManagedQueryResultsEncryptionConfiguration(v.([]any))
	}

	return managedQueryResultsConfiguration
}

func expandManagedQueryResultsEncryptionConfiguration(l []any) *types.ManagedQueryResultsEncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	if v, ok := m[names.AttrKMSKey].(string); !ok || v == "" {
		return nil
	}
	managedQueryResultsEncryptionConfiguration := &types.ManagedQueryResultsEncryptionConfiguration{}
	managedQueryResultsEncryptionConfiguration.KmsKey = aws.String(m[names.AttrKMSKey].(string))

	return managedQueryResultsEncryptionConfiguration
}

func expandWorkGroupManagedQueryResultsConfigurationUpdates(l []any) *types.ManagedQueryResultsConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	managedQueryResultsConfigurationUpdates := &types.ManagedQueryResultsConfigurationUpdates{}
	if v, ok := m[names.AttrEnabled].(bool); ok {
		managedQueryResultsConfigurationUpdates.Enabled = aws.Bool(v)
		if !v {
			managedQueryResultsConfigurationUpdates.RemoveEncryptionConfiguration = aws.Bool(true)
			return managedQueryResultsConfigurationUpdates
		}
	}

	if v, ok := m[names.AttrEncryptionConfiguration]; ok {
		encConfig := expandManagedQueryResultsEncryptionConfiguration(v.([]any))
		if encConfig != nil {
			managedQueryResultsConfigurationUpdates.EncryptionConfiguration = encConfig
		} else {
			managedQueryResultsConfigurationUpdates.RemoveEncryptionConfiguration = aws.Bool(true)
		}
	}

	return managedQueryResultsConfigurationUpdates
}

func flattenWorkGroupConfiguration(configuration *types.WorkGroupConfiguration) []any {
	if configuration == nil {
		return []any{}
	}

	m := map[string]any{
		"bytes_scanned_cutoff_per_query":               aws.ToInt64(configuration.BytesScannedCutoffPerQuery),
		"customer_content_encryption_configuration":    flattenWorkGroupCustomerContentEncryptionConfiguration(configuration.CustomerContentEncryptionConfiguration),
		"enable_minimum_encryption_configuration":      aws.ToBool(configuration.EnableMinimumEncryptionConfiguration),
		"enforce_workgroup_configuration":              aws.ToBool(configuration.EnforceWorkGroupConfiguration),
		names.AttrEngineVersion:                        flattenWorkGroupEngineVersion(configuration.EngineVersion),
		"execution_role":                               aws.ToString(configuration.ExecutionRole),
		"identity_center_configuration":                flattenWorkGroupIdentityCenterConfiguration(configuration.IdentityCenterConfiguration),
		"query_results_s3_access_grants_configuration": flattenWorkGroupQueryResultsS3AccessGrantsConfiguration(configuration.QueryResultsS3AccessGrantsConfiguration),
		"managed_query_results_configuration":          flattenWorkGroupManagedQueryResultsConfiguration(configuration.ManagedQueryResultsConfiguration),
		"monitoring_configuration":                     flattenWorkGroupMonitoringConfiguration(configuration.MonitoringConfiguration),
		"publish_cloudwatch_metrics_enabled":           aws.ToBool(configuration.PublishCloudWatchMetricsEnabled),
		"result_configuration":                         flattenWorkGroupResultConfiguration(configuration.ResultConfiguration),
		"requester_pays_enabled":                       aws.ToBool(configuration.RequesterPaysEnabled),
	}

	return []any{m}
}

func flattenWorkGroupCustomerContentEncryptionConfiguration(encryptionConfiguration *types.CustomerContentEncryptionConfiguration) []any {
	if encryptionConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrKMSKey: aws.ToString(encryptionConfiguration.KmsKey),
	}

	return []any{m}
}

func flattenWorkGroupEngineVersion(engineVersion *types.EngineVersion) []any {
	if engineVersion == nil {
		return []any{}
	}

	m := map[string]any{
		"effective_engine_version": aws.ToString(engineVersion.EffectiveEngineVersion),
		"selected_engine_version":  aws.ToString(engineVersion.SelectedEngineVersion),
	}

	return []any{m}
}

func flattenWorkGroupIdentityCenterConfiguration(identityCenterConfiguration *types.IdentityCenterConfiguration) []any {
	if identityCenterConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		"enable_identity_center":       aws.ToBool(identityCenterConfiguration.EnableIdentityCenter),
		"identity_center_instance_arn": aws.ToString(identityCenterConfiguration.IdentityCenterInstanceArn),
	}

	return []any{m}
}

func flattenWorkGroupQueryResultsS3AccessGrantsConfiguration(queryResultsS3AccessGrantsConfiguration *types.QueryResultsS3AccessGrantsConfiguration) []any {
	if queryResultsS3AccessGrantsConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		"authentication_type":      string(queryResultsS3AccessGrantsConfiguration.AuthenticationType),
		"create_user_level_prefix": aws.ToBool(queryResultsS3AccessGrantsConfiguration.CreateUserLevelPrefix),
		"enable_s3_access_grants":  aws.ToBool(queryResultsS3AccessGrantsConfiguration.EnableS3AccessGrants),
	}

	return []any{m}
}

func flattenWorkGroupMonitoringConfiguration(monitoringConfiguration *types.MonitoringConfiguration) []any {
	if monitoringConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		"cloud_watch_logging_configuration": flattenWorkGroupMonitoringConfigurationCloudWatchLoggingConfiguration(monitoringConfiguration.CloudWatchLoggingConfiguration),
		"managed_logging_configuration":     flattenWorkGroupMonitoringConfigurationManagedLoggingConfiguration(monitoringConfiguration.ManagedLoggingConfiguration),
		"s3_logging_configuration":          flattenWorkGroupMonitoringConfigurationS3LoggingConfiguration(monitoringConfiguration.S3LoggingConfiguration),
	}

	return []any{m}
}

func flattenWorkGroupMonitoringConfigurationCloudWatchLoggingConfiguration(cloudWatchLoggingConfiguration *types.CloudWatchLoggingConfiguration) []any {
	if cloudWatchLoggingConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEnabled:        aws.ToBool(cloudWatchLoggingConfiguration.Enabled),
		"log_group":              aws.ToString(cloudWatchLoggingConfiguration.LogGroup),
		"log_stream_name_prefix": aws.ToString(cloudWatchLoggingConfiguration.LogStreamNamePrefix),
	}

	if len(cloudWatchLoggingConfiguration.LogTypes) > 0 {
		logTypes := make([]any, 0, len(cloudWatchLoggingConfiguration.LogTypes))
		for key, values := range cloudWatchLoggingConfiguration.LogTypes {
			logTypeMap := map[string]any{
				names.AttrKey:    key,
				names.AttrValues: values,
			}
			logTypes = append(logTypes, logTypeMap)
		}
		m["log_type"] = logTypes
	}

	return []any{m}
}

func flattenWorkGroupMonitoringConfigurationManagedLoggingConfiguration(managedLoggingConfiguration *types.ManagedLoggingConfiguration) []any {
	if managedLoggingConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEnabled: aws.ToBool(managedLoggingConfiguration.Enabled),
		names.AttrKMSKey:  aws.ToString(managedLoggingConfiguration.KmsKey),
	}

	return []any{m}
}

func flattenWorkGroupMonitoringConfigurationS3LoggingConfiguration(s3LoggingConfiguration *types.S3LoggingConfiguration) []any {
	if s3LoggingConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEnabled: aws.ToBool(s3LoggingConfiguration.Enabled),
		names.AttrKMSKey:  aws.ToString(s3LoggingConfiguration.KmsKey),
		"log_location":    aws.ToString(s3LoggingConfiguration.LogLocation),
	}

	return []any{m}
}

func flattenWorkGroupResultConfiguration(resultConfiguration *types.ResultConfiguration) []any {
	if resultConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEncryptionConfiguration: flattenWorkGroupEncryptionConfiguration(resultConfiguration.EncryptionConfiguration),
		"output_location":                 aws.ToString(resultConfiguration.OutputLocation),
	}

	if resultConfiguration.ExpectedBucketOwner != nil {
		m[names.AttrExpectedBucketOwner] = aws.ToString(resultConfiguration.ExpectedBucketOwner)
	}

	if resultConfiguration.AclConfiguration != nil {
		m["acl_configuration"] = flattenWorkGroupACLConfiguration(resultConfiguration.AclConfiguration)
	}

	return []any{m}
}

func flattenWorkGroupEncryptionConfiguration(encryptionConfiguration *types.EncryptionConfiguration) []any {
	if encryptionConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		"encryption_option": encryptionConfiguration.EncryptionOption,
		names.AttrKMSKeyARN: aws.ToString(encryptionConfiguration.KmsKey),
	}

	return []any{m}
}

func flattenWorkGroupACLConfiguration(aclConfig *types.AclConfiguration) []any {
	if aclConfig == nil {
		return []any{}
	}

	m := map[string]any{
		"s3_acl_option": aclConfig.S3AclOption,
	}

	return []any{m}
}

func flattenWorkGroupManagedQueryResultsConfiguration(managedQueryResultsConfiguration *types.ManagedQueryResultsConfiguration) []any {
	if managedQueryResultsConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEnabled:                 aws.ToBool(&managedQueryResultsConfiguration.Enabled),
		names.AttrEncryptionConfiguration: flattenWorkGroupManagedQueryResultsEncryptionConfiguration(managedQueryResultsConfiguration.EncryptionConfiguration),
	}

	return []any{m}
}

func flattenWorkGroupManagedQueryResultsEncryptionConfiguration(managedQueryResultsEncryptionConfiguration *types.ManagedQueryResultsEncryptionConfiguration) []any {
	if managedQueryResultsEncryptionConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrKMSKey: aws.ToString(managedQueryResultsEncryptionConfiguration.KmsKey),
	}

	return []any{m}
}

func managedQueryResultsValidation(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	configRaw, configOk := diff.GetOk(names.AttrConfiguration)
	if !configOk {
		return nil
	}

	configList, ok := configRaw.([]any)
	if !ok || len(configList) == 0 || configList[0] == nil {
		return nil
	}

	config := configList[0].(map[string]any)

	mqrEnabled := false
	if mqrRaw, mqrOk := config["managed_query_results_configuration"]; mqrOk {
		if mqrList, ok := mqrRaw.([]any); ok && len(mqrList) > 0 && mqrList[0] != nil {
			if mqrConfig, ok := mqrList[0].(map[string]any); ok {
				if enabled, ok := mqrConfig[names.AttrEnabled].(bool); ok {
					mqrEnabled = enabled
				}
			}
		}
	}

	if !mqrEnabled {
		return nil
	}

	if rcRaw, rcOk := config["result_configuration"]; rcOk {
		if rcList, ok := rcRaw.([]any); ok && len(rcList) > 0 && rcList[0] != nil {
			if rcConfig, ok := rcList[0].(map[string]any); ok {
				if outputLoc, ok := rcConfig["output_location"].(string); ok && outputLoc != "" {
					return fmt.Errorf("configuration.result_configuration.output_location cannot be specified when configuration.managed_query_results_configuration.enabled is true")
				}
			}
		}
	}

	return nil
}

func workGroupARN(ctx context.Context, c *conns.AWSClient, workGroupName string) string {
	return c.RegionalARN(ctx, "athena", fmt.Sprintf("workgroup/%s", workGroupName))
}
