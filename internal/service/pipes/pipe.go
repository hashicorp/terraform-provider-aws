// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pipes

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pipes_pipe", name="Pipe")
// @Tags(identifierAttribute="arn")
func resourcePipe() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipeCreate,
		ReadWithoutTimeout:   resourcePipeRead,
		UpdateWithoutTimeout: resourcePipeUpdate,
		DeleteWithoutTimeout: resourcePipeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "Managed by Terraform",
				},
				"desired_state": {
					Type:             schema.TypeString,
					Optional:         true,
					Default:          string(awstypes.RequestedPipeStateRunning),
					ValidateDiagFunc: enum.Validate[awstypes.RequestedPipeState](),
				},
				"enrichment": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidARN,
				},
				"enrichment_parameters": enrichmentParametersSchema(),
				"log_configuration":     logConfigurationSchema(),
				names.AttrName: {
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					ForceNew:      true,
					ConflictsWith: []string{names.AttrNamePrefix},
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 64),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), ""),
					),
				},
				names.AttrNamePrefix: {
					Type:          schema.TypeString,
					Optional:      true,
					Computed:      true,
					ForceNew:      true,
					ConflictsWith: []string{names.AttrName},
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 64-id.UniqueIDSuffixLength),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), ""),
					),
				},
				names.AttrRoleARN: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				names.AttrSource: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					ValidateFunc: validation.Any(
						verify.ValidARN,
						validation.StringMatch(regexache.MustCompile(`^smk://(([0-9A-Za-z]|[0-9A-Za-z][0-9A-Za-z-]*[0-9A-Za-z])\.)*([0-9A-Za-z]|[0-9A-Za-z][0-9A-Za-z-]*[0-9A-Za-z]):[0-9]{1,5}|arn:(aws[0-9A-Za-z-]*):([0-9A-Za-z-]+):([a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\d{1})?:(\d{12})?:(.+)$`), ""),
					),
				},
				"source_parameters": sourceParametersSchema(),
				names.AttrTarget: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"target_parameters": targetParametersSchema(),
				names.AttrTags:      tftags.TagsSchema(),
				names.AttrTagsAll:   tftags.TagsSchemaComputed(),
			}
		},
	}
}

const (
	ResNamePipe = "Pipe"
)

func resourcePipeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PipesClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &pipes.CreatePipeInput{
		DesiredState: awstypes.RequestedPipeState(d.Get("desired_state").(string)),
		Name:         aws.String(name),
		RoleArn:      aws.String(d.Get(names.AttrRoleARN).(string)),
		Source:       aws.String(d.Get(names.AttrSource).(string)),
		Tags:         getTagsIn(ctx),
		Target:       aws.String(d.Get(names.AttrTarget).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enrichment"); ok && v != "" {
		input.Enrichment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enrichment_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EnrichmentParameters = expandPipeEnrichmentParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("source_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SourceParameters = expandPipeSourceParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("target_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TargetParameters = expandPipeTargetParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("log_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LogConfiguration = expandPipeLogConfigurationParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreatePipe(ctx, input)

	if err != nil {
		return create.AppendDiagError(diags, names.Pipes, create.ErrActionCreating, ResNamePipe, name, err)
	}

	d.SetId(aws.ToString(output.Name))

	if _, err := waitPipeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.Pipes, create.ErrActionWaitingForCreation, ResNamePipe, d.Id(), err)
	}

	return append(diags, resourcePipeRead(ctx, d, meta)...)
}

func resourcePipeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PipesClient(ctx)

	output, err := findPipeByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Pipes Pipe (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Pipes, create.ErrActionReading, ResNamePipe, d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("desired_state", output.DesiredState)
	d.Set("enrichment", output.Enrichment)
	if v := output.EnrichmentParameters; !types.IsZero(v) {
		if err := d.Set("enrichment_parameters", []interface{}{flattenPipeEnrichmentParameters(v)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting enrichment_parameters: %s", err)
		}
	} else {
		d.Set("enrichment_parameters", nil)
	}
	if v := output.LogConfiguration; !types.IsZero(v) {
		if err := d.Set("log_configuration", []interface{}{flattenPipeLogConfiguration(v)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting log_configuration: %s", err)
		}
	} else {
		d.Set("log_configuration", nil)
	}
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.Name)))
	d.Set(names.AttrRoleARN, output.RoleArn)
	d.Set(names.AttrSource, output.Source)
	if v := output.SourceParameters; !types.IsZero(v) {
		if err := d.Set("source_parameters", []interface{}{flattenPipeSourceParameters(v)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting source_parameters: %s", err)
		}
	} else {
		d.Set("source_parameters", nil)
	}
	d.Set(names.AttrTarget, output.Target)
	if v := output.TargetParameters; !types.IsZero(v) {
		if err := d.Set("target_parameters", []interface{}{flattenPipeTargetParameters(v)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_parameters: %s", err)
		}
	} else {
		d.Set("target_parameters", nil)
	}

	return diags
}

func resourcePipeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PipesClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &pipes.UpdatePipeInput{
			Description:  aws.String(d.Get(names.AttrDescription).(string)),
			DesiredState: awstypes.RequestedPipeState(d.Get("desired_state").(string)),
			Name:         aws.String(d.Id()),
			RoleArn:      aws.String(d.Get(names.AttrRoleARN).(string)),
			Target:       aws.String(d.Get(names.AttrTarget).(string)),
			// Reset state in case it's a deletion, have to set the input to an empty string otherwise it doesn't get overwritten.
			TargetParameters: &awstypes.PipeTargetParameters{
				InputTemplate: aws.String(""),
			},
		}

		if d.HasChange("enrichment") {
			input.Enrichment = aws.String(d.Get("enrichment").(string))
		}

		if d.HasChange("enrichment_parameters") {
			if v, ok := d.GetOk("enrichment_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.EnrichmentParameters = expandPipeEnrichmentParameters(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("log_configuration") {
			if v, ok := d.GetOk("log_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.LogConfiguration = expandPipeLogConfigurationParameters(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("source_parameters") {
			if v, ok := d.GetOk("source_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.SourceParameters = expandUpdatePipeSourceParameters(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("target_parameters") {
			if v, ok := d.GetOk("target_parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.TargetParameters = expandPipeTargetParameters(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		output, err := conn.UpdatePipe(ctx, input)

		if err != nil {
			return create.AppendDiagError(diags, names.Pipes, create.ErrActionUpdating, ResNamePipe, d.Id(), err)
		}

		if _, err := waitPipeUpdated(ctx, conn, aws.ToString(output.Name), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.Pipes, create.ErrActionWaitingForUpdate, ResNamePipe, d.Id(), err)
		}
	}

	return append(diags, resourcePipeRead(ctx, d, meta)...)
}

func resourcePipeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PipesClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Pipes Pipe: %s", d.Id())
	_, err := conn.DeletePipe(ctx, &pipes.DeletePipeInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Pipes, create.ErrActionDeleting, ResNamePipe, d.Id(), err)
	}

	if _, err := waitPipeDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.Pipes, create.ErrActionWaitingForDeletion, ResNamePipe, d.Id(), err)
	}

	return diags
}

func findPipeByName(ctx context.Context, conn *pipes.Client, name string) (*pipes.DescribePipeOutput, error) {
	input := &pipes.DescribePipeInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribePipe(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Arn == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusPipe(ctx context.Context, conn *pipes.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findPipeByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.CurrentState), nil
	}
}

func waitPipeCreated(ctx context.Context, conn *pipes.Client, id string, timeout time.Duration) (*pipes.DescribePipeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PipeStateCreating),
		Target:                    enum.Slice(awstypes.PipeStateRunning, awstypes.PipeStateStopped),
		Refresh:                   statusPipe(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*pipes.DescribePipeOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitPipeUpdated(ctx context.Context, conn *pipes.Client, id string, timeout time.Duration) (*pipes.DescribePipeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PipeStateUpdating),
		Target:                    enum.Slice(awstypes.PipeStateRunning, awstypes.PipeStateStopped),
		Refresh:                   statusPipe(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*pipes.DescribePipeOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitPipeDeleted(ctx context.Context, conn *pipes.Client, id string, timeout time.Duration) (*pipes.DescribePipeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PipeStateDeleting),
		Target:  []string{},
		Refresh: statusPipe(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*pipes.DescribePipeOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func suppressEmptyConfigurationBlock(key string) schema.SchemaDiffSuppressFunc {
	return func(k, o, n string, d *schema.ResourceData) bool {
		if k != key+".#" {
			return false
		}

		if o == "0" && n == "1" {
			v := d.Get(key).([]interface{})
			return len(v) == 0 || v[0] == nil || len(v[0].(map[string]interface{})) == 0
		}

		return false
	}
}
