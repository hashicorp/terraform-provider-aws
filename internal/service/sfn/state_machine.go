// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sfn_state_machine", name="State Machine")
// @Tags(identifierAttribute="id")
func ResourceStateMachine() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStateMachineCreate,
		ReadWithoutTimeout:   resourceStateMachineRead,
		UpdateWithoutTimeout: resourceStateMachineUpdate,
		DeleteWithoutTimeout: resourceStateMachineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definition": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024*1024), // 1048576
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_execution_data": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"level": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(sfn.LogLevel_Values(), false),
						},
						"log_destination": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 80),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), "the name should only contain 0-9, A-Z, a-z, - and _"),
				),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 80-id.UniqueIDSuffixLength),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), "the name should only contain 0-9, A-Z, a-z, - and _"),
				),
			},
			"publish": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_machine_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tracing_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      sfn.StateMachineTypeStandard,
				ValidateFunc: validation.StringInSlice(sfn.StateMachineType_Values(), false),
			},
			"version_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStateMachineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &sfn.CreateStateMachineInput{
		Definition: aws.String(d.Get("definition").(string)),
		Name:       aws.String(name),
		Publish:    aws.Bool(d.Get("publish").(bool)),
		RoleArn:    aws.String(d.Get("role_arn").(string)),
		Tags:       getTagsIn(ctx),
		Type:       aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("logging_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LoggingConfiguration = expandLoggingConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("tracing_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TracingConfiguration = expandTracingConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	// This is done to deal with IAM eventual consistency.
	// Note: the instance may be in a deleting mode, hence the retry
	// when creating the step function. This can happen when we are
	// updating the resource (since there is no update API call).
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateStateMachineWithContext(ctx, input)
	}, sfn.ErrCodeStateMachineDeleting, "AccessDeniedException")

	if err != nil {
		return diag.Errorf("creating Step Functions State Machine (%s): %s", name, err)
	}

	arn := aws.StringValue(outputRaw.(*sfn.CreateStateMachineOutput).StateMachineArn)
	d.SetId(arn)
	d.Set("state_machine_version_arn", outputRaw.(*sfn.CreateStateMachineOutput).StateMachineVersionArn)

	return resourceStateMachineRead(ctx, d, meta)
}

func resourceStateMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	output, err := FindStateMachineByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Step Functions State Machine (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Step Functions State Machine (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.StateMachineArn)
	if output.CreationDate != nil {
		d.Set("creation_date", aws.TimeValue(output.CreationDate).Format(time.RFC3339))
	} else {
		d.Set("creation_date", nil)
	}
	d.Set("definition", output.Definition)
	d.Set("description", output.Description)

	if output.LoggingConfiguration != nil {
		if err := d.Set("logging_configuration", []interface{}{flattenLoggingConfiguration(output.LoggingConfiguration)}); err != nil {
			return diag.Errorf("setting logging_configuration: %s", err)
		}
	} else {
		d.Set("logging_configuration", nil)
	}
	d.Set("name", output.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(output.Name)))
	d.Set("role_arn", output.RoleArn)
	d.Set("revision_id", output.RevisionId)
	d.Set("status", output.Status)
	if output.TracingConfiguration != nil {
		if err := d.Set("tracing_configuration", []interface{}{flattenTracingConfiguration(output.TracingConfiguration)}); err != nil {
			return diag.Errorf("setting tracing_configuration: %s", err)
		}
	} else {
		d.Set("tracing_configuration", nil)
	}
	d.Set("type", output.Type)

	return nil
}

func resourceStateMachineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		// "You must include at least one of definition or roleArn or you will receive a MissingRequiredParameter error"
		input := &sfn.UpdateStateMachineInput{
			Definition:      aws.String(d.Get("definition").(string)),
			RoleArn:         aws.String(d.Get("role_arn").(string)),
			StateMachineArn: aws.String(d.Id()),
			Publish:         aws.Bool(d.Get("publish").(bool)),
		}

		if v, ok := d.GetOk("publish"); ok && v == true {
			input.VersionDescription = aws.String(d.Get("version_description").(string))
		}

		if d.HasChange("logging_configuration") {
			if v, ok := d.GetOk("logging_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.LoggingConfiguration = expandLoggingConfiguration(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("tracing_configuration") {
			if v, ok := d.GetOk("tracing_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.TracingConfiguration = expandTracingConfiguration(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		out, err := conn.UpdateStateMachineWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Step Functions State Machine (%s): %s", d.Id(), err)
		}

		// Handle eventual consistency after update.
		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError { // nosemgrep:ci.helper-schema-retry-RetryContext-without-TimeoutError-check
			output, err := FindStateMachineByARN(ctx, conn, d.Id())

			if err != nil {
				return retry.NonRetryableError(err)
			}

			if d.HasChange("definition") && !verify.JSONBytesEqual([]byte(aws.StringValue(output.Definition)), []byte(d.Get("definition").(string))) ||
				d.HasChange("role_arn") && aws.StringValue(output.RoleArn) != d.Get("role_arn").(string) ||
				//d.HasChange("publish") && aws.Bool(output.Publish) != d.Get("publish").(bool) ||
				d.HasChange("tracing_configuration.0.enabled") && output.TracingConfiguration != nil && aws.BoolValue(output.TracingConfiguration.Enabled) != d.Get("tracing_configuration.0.enabled").(bool) ||
				d.HasChange("logging_configuration.0.include_execution_data") && output.LoggingConfiguration != nil && aws.BoolValue(output.LoggingConfiguration.IncludeExecutionData) != d.Get("logging_configuration.0.include_execution_data").(bool) ||
				d.HasChange("logging_configuration.0.level") && output.LoggingConfiguration != nil && aws.StringValue(output.LoggingConfiguration.Level) != d.Get("logging_configuration.0.level").(string) {
				return retry.RetryableError(fmt.Errorf("Step Functions State Machine (%s) eventual consistency", d.Id()))
			}

			return nil
		})

		if err != nil {
			return diag.Errorf("waiting for Step Functions State Machine (%s) update: %s", d.Id(), err)
		}

		d.Set("state_machine_version_arn", out.StateMachineVersionArn)
	}

	return resourceStateMachineRead(ctx, d, meta)
}

func resourceStateMachineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SFNConn(ctx)

	log.Printf("[DEBUG] Deleting Step Functions State Machine: %s", d.Id())
	_, err := conn.DeleteStateMachineWithContext(ctx, &sfn.DeleteStateMachineInput{
		StateMachineArn: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting Step Functions State Machine (%s): %s", d.Id(), err)
	}

	if _, err := waitStateMachineDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Step Functions State Machine (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindStateMachineByARN(ctx context.Context, conn *sfn.SFN, arn string) (*sfn.DescribeStateMachineOutput, error) {
	input := &sfn.DescribeStateMachineInput{
		StateMachineArn: aws.String(arn),
	}

	output, err := conn.DescribeStateMachineWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sfn.ErrCodeStateMachineDoesNotExist) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusStateMachine(ctx context.Context, conn *sfn.SFN, stateMachineArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStateMachineByARN(ctx, conn, stateMachineArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitStateMachineDeleted(ctx context.Context, conn *sfn.SFN, stateMachineArn string, timeout time.Duration) (*sfn.DescribeStateMachineOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{sfn.StateMachineStatusActive, sfn.StateMachineStatusDeleting},
		Target:  []string{},
		Refresh: statusStateMachine(ctx, conn, stateMachineArn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sfn.DescribeStateMachineOutput); ok {
		return output, err
	}

	return nil, err
}

func expandLoggingConfiguration(tfMap map[string]interface{}) *sfn.LoggingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &sfn.LoggingConfiguration{}

	if v, ok := tfMap["include_execution_data"].(bool); ok {
		apiObject.IncludeExecutionData = aws.Bool(v)
	}

	if v, ok := tfMap["level"].(string); ok && v != "" {
		apiObject.Level = aws.String(v)
	}

	if v, ok := tfMap["log_destination"].(string); ok && v != "" {
		apiObject.Destinations = []*sfn.LogDestination{{
			CloudWatchLogsLogGroup: &sfn.CloudWatchLogsLogGroup{
				LogGroupArn: aws.String(v),
			},
		}}
	}

	return apiObject
}

func flattenLoggingConfiguration(apiObject *sfn.LoggingConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.IncludeExecutionData; v != nil {
		tfMap["include_execution_data"] = aws.BoolValue(v)
	}

	if v := apiObject.Level; v != nil {
		tfMap["level"] = aws.StringValue(v)
	}

	if v := apiObject.Destinations; len(v) > 0 {
		tfMap["log_destination"] = aws.StringValue(v[0].CloudWatchLogsLogGroup.LogGroupArn)
	}

	return tfMap
}

func expandTracingConfiguration(tfMap map[string]interface{}) *sfn.TracingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &sfn.TracingConfiguration{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func flattenTracingConfiguration(apiObject *sfn.TracingConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	return tfMap
}
