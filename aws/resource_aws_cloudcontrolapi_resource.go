package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudcontrolapi"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	cfschema "github.com/hashicorp/aws-cloudformation-resource-schema-sdk-go"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mattbaird/jsonpatch"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudcontrolapi/waiter"
)

func resourceAwsCloudControlApiResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsCloudControlApiResourceCreate,
		DeleteContext: resourceAwsCloudControlApiResourceDelete,
		ReadContext:   resourceAwsCloudControlApiResourceRead,
		UpdateContext: resourceAwsCloudControlApiResourceUpdate,

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
			"properties": {
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
			resourceAwsCloudControlApiResourceCustomizeDiffGetSchema,
			resourceAwsCloudControlApiResourceCustomizeDiffSchemaDiff,
			customdiff.ComputedIf("properties", func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("desired_state")
			}),
		),
	}
}

func resourceAwsCloudControlApiResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudcontrolapiconn

	input := &cloudcontrolapi.CreateResourceInput{
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

	return resourceAwsCloudControlApiResourceRead(ctx, d, meta)
}

func resourceAwsCloudControlApiResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudcontrolapiconn

	input := &cloudcontrolapi.GetResourceInput{
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

	if tfawserr.ErrCodeEquals(err, cloudcontrolapi.ErrCodeResourceNotFoundException) {
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

	d.Set("properties", output.ResourceDescription.Properties)

	return nil
}

func resourceAwsCloudControlApiResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudcontrolapiconn

	if d.HasChange("desired_state") {
		oldRaw, newRaw := d.GetChange("desired_state")

		patchDocument, err := patchDocument(oldRaw.(string), newRaw.(string))

		if err != nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "JSON Patch Creation Unsuccessful",
					Detail:   fmt.Sprintf("Creating JSON Patch failed: %s", err.Error()),
				},
			}
		}

		input := &cloudcontrolapi.UpdateResourceInput{
			ClientToken:   aws.String(resource.UniqueId()),
			Identifier:    aws.String(d.Id()),
			PatchDocument: aws.String(patchDocument),
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

	return resourceAwsCloudControlApiResourceRead(ctx, d, meta)
}

func resourceAwsCloudControlApiResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudcontrolapiconn

	input := &cloudcontrolapi.DeleteResourceInput{
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

	if progressEvent != nil && aws.StringValue(progressEvent.ErrorCode) == cloudcontrolapi.HandlerErrorCodeNotFound {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for CloudFormation Resource (%s) deletion: %w", d.Id(), err))
	}

	return nil
}

func resourceAwsCloudControlApiResourceCustomizeDiffGetSchema(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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

	if err := diff.SetNew("schema", output.Schema); err != nil {
		return fmt.Errorf("error setting schema diff: %w", err)
	}

	return nil
}

func resourceAwsCloudControlApiResourceCustomizeDiffSchemaDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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

	patches, err := jsonpatch.CreatePatch([]byte(oldDesiredStateRaw.(string)), []byte(newDesiredStateRaw.(string)))

	if err != nil {
		return fmt.Errorf("error creating desired_state JSON Patch: %w", err)
	}

	for _, patch := range patches {
		if cfResource.IsCreateOnlyPropertyPath(patch.Path) {
			if err := diff.ForceNew("desired_state"); err != nil {
				return fmt.Errorf("error setting desired_state ForceNew: %w", err)
			}

			break
		}
	}

	return nil
}

// patchDocument returns a JSON Patch document describing the difference between `old` and `new`.
func patchDocument(old, new string) (string, error) {
	patch, err := jsonpatch.CreatePatch([]byte(old), []byte(new))

	if err != nil {
		return "", err
	}

	b, err := json.Marshal(patch)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
