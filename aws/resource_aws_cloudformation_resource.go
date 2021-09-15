package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	cfschema "github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/cfjsonpatch"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/waiter"
)

func resourceAwsCloudFormationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsCloudFormationResourceCreate,
		DeleteContext: resourceAwsCloudFormationResourceDelete,
		ReadContext:   resourceAwsCloudFormationResourceRead,
		UpdateContext: resourceAwsCloudFormationResourceUpdate,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Hour),
			Delete: schema.DefaultTimeout(2 * time.Hour),
			Update: schema.DefaultTimeout(2 * time.Hour),
		},

		Schema: map[string]*schema.Schema{
			"desired_state": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"schema": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"type_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}`), "must be three alphanumeric sections separated by double colons (::)"),
			},
			"type_version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			resourceAwsCloudFormationResourceCustomizeDiffGetSchema,
			resourceAwsCloudFormationResourceCustomizeDiffSchemaDiff,
			customdiff.ComputedIf("resource_model", func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("desired_state")
			}),
		),
	}
}

func resourceAwsCloudFormationResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.CreateResourceInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("desired_state"); ok {
		input.DesiredState = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_name"); ok {
		input.TypeName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_version_id"); ok {
		input.TypeVersionId = aws.String(v.(string))
	}

	output, err := conn.CreateResourceWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating CloudFormation Resource: %w", err))
	}

	if output == nil || output.ProgressEvent == nil {
		return diag.FromErr(fmt.Errorf("error creating CloudFormation Resource: empty response"))
	}

	// Always try to capture the identifier before returning errors
	d.SetId(aws.StringValue(output.ProgressEvent.Identifier))

	output.ProgressEvent, err = waiter.ResourceRequestStatusProgressEventOperationStatusSuccess(ctx, conn, aws.StringValue(output.ProgressEvent.RequestToken), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for CloudForamtion Resource (%s) creation: %w", d.Id(), err))
	}

	// Some resources do not set the identifier until after creation
	if d.Id() == "" {
		d.SetId(aws.StringValue(output.ProgressEvent.Identifier))
	}

	return resourceAwsCloudFormationResourceRead(ctx, d, meta)
}

func resourceAwsCloudFormationResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.GetResourceInput{
		Identifier: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_name"); ok {
		input.TypeName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_version_id"); ok {
		input.TypeVersionId = aws.String(v.(string))
	}

	output, err := conn.GetResourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeResourceNotFoundException) {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading CloudFormation Resource (%s): not found after creation", d.Id()))
		}

		log.Printf("[WARN] CloudFormation Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading CloudFormation Resource (%s): %w", d.Id(), err))
	}

	if output == nil || output.ResourceDescription == nil {
		return diag.FromErr(fmt.Errorf("error reading CloudFormation Resource (%s): empty response", d.Id()))
	}

	d.Set("resource_model", output.ResourceDescription.ResourceModel)

	return nil
}

func resourceAwsCloudFormationResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cfconn

	if d.HasChange("desired_state") {
		oldRaw, newRaw := d.GetChange("desired_state")

		patchOperations, err := cfjsonpatch.PatchOperations(oldRaw, newRaw)

		if err != nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "JSON Patch Creation Unsuccessful",
					Detail:   fmt.Sprintf("Creating JSON Patch failed: %s", err.Error()),
				},
			}
		}

		input := &cloudformation.UpdateResourceInput{
			ClientToken:     aws.String(resource.UniqueId()),
			Identifier:      aws.String(d.Id()),
			PatchOperations: patchOperations,
		}

		if v, ok := d.GetOk("role_arn"); ok {
			input.RoleArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("type_name"); ok {
			input.TypeName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("type_version_id"); ok {
			input.TypeVersionId = aws.String(v.(string))
		}

		output, err := conn.UpdateResourceWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating CloudFormation Resource (%s): %w", d.Id(), err))
		}

		if output == nil || output.ProgressEvent == nil {
			return diag.FromErr(fmt.Errorf("error updating CloudFormation Resource (%s): empty reponse", d.Id()))
		}

		if _, err := waiter.ResourceRequestStatusProgressEventOperationStatusSuccess(ctx, conn, aws.StringValue(output.ProgressEvent.RequestToken), d.Timeout(schema.TimeoutDelete)); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for CloudFormation Resource (%s) update: %w", d.Id(), err))
		}
	}

	return resourceAwsCloudFormationResourceRead(ctx, d, meta)
}

func resourceAwsCloudFormationResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.DeleteResourceInput{
		ClientToken: aws.String(resource.UniqueId()),
		Identifier:  aws.String(d.Id()),
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_name"); ok {
		input.TypeName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_version_id"); ok {
		input.TypeVersionId = aws.String(v.(string))
	}

	output, err := conn.DeleteResourceWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting CloudFormation Resource (%s): %w", d.Id(), err))
	}

	if output == nil || output.ProgressEvent == nil {
		return diag.FromErr(fmt.Errorf("error deleting CloudFormation Resource (%s): empty response", d.Id()))
	}

	progressEvent, err := waiter.ResourceRequestStatusProgressEventOperationStatusSuccess(ctx, conn, aws.StringValue(output.ProgressEvent.RequestToken), d.Timeout(schema.TimeoutDelete))

	if progressEvent != nil && aws.StringValue(progressEvent.ErrorCode) == cloudformation.HandlerErrorCodeNotFound {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for CloudFormation Resource (%s) deletion: %w", d.Id(), err))
	}

	return nil
}

func resourceAwsCloudFormationResourceCustomizeDiffGetSchema(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	resourceSchema := diff.Get("schema").(string)
	typeName := diff.Get("type_name").(string)

	if resourceSchema != "" {
		return nil
	}

	input := &cloudformation.DescribeTypeInput{
		Type:     aws.String(cloudformation.RegistryTypeResource),
		TypeName: aws.String(typeName),
	}

	output, err := conn.DescribeTypeWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("error describing CloudFormation Type (%s): %w", typeName, err)
	}

	if output == nil {
		return fmt.Errorf("error describing CloudFormation Type (%s): empty reponse", typeName)
	}

	if err := diff.SetNew("schema", aws.StringValue(output.Schema)); err != nil {
		return fmt.Errorf("error setting schema diff: %w", err)
	}

	return nil
}

func resourceAwsCloudFormationResourceCustomizeDiffSchemaDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	oldDesiredStateRaw, newDesiredStateRaw := diff.GetChange("desired_state")
	newSchema := diff.Get("schema").(string)

	newDesiredState, ok := newDesiredStateRaw.(string)

	if !ok {
		return fmt.Errorf("unexpected new desired_state value type: %T", newDesiredStateRaw)
	}

	// desired_state can be empty if unknown
	if newDesiredState == "" {
		return nil
	}

	cfResourceSchema, err := cfschema.NewResourceJsonSchemaDocument(newSchema)

	if err != nil {
		return fmt.Errorf("error parsing CloudFormation Resource Schema JSON: %w", err)
	}

	if err := cfResourceSchema.ValidateConfigurationDocument(newDesiredState); err != nil {
		return fmt.Errorf("error validating desired_state against CloudFormation Resource Schema: %w", err)
	}

	// Do nothing further for new resources or if desired state is not changed
	if diff.Id() == "" || !diff.HasChange("desired_state") {
		return nil
	}

	cfResource, err := cfResourceSchema.Resource()

	if err != nil {
		return fmt.Errorf("error converting CloudFormation Resource Schema JSON: %w", err)
	}

	patchOperations, err := cfjsonpatch.PatchOperations(oldDesiredStateRaw, newDesiredStateRaw)

	if err != nil {
		return fmt.Errorf("error creating desired_state JSON Patch: %w", err)
	}

	for _, patchOperation := range patchOperations {
		if patchOperation == nil {
			continue
		}

		if cfResource.IsCreateOnlyPropertyPath(aws.StringValue(patchOperation.Path)) {
			if err := diff.ForceNew("desired_state"); err != nil {
				return fmt.Errorf("error setting desired_state ForceNew: %w", err)
			}

			break
		}
	}

	return nil
}
