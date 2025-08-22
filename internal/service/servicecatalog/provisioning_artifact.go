// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_provisioning_artifact", name="Provisioning Artifact")
func resourceProvisioningArtifact() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProvisioningArtifactCreate,
		ReadWithoutTimeout:   resourceProvisioningArtifactRead,
		UpdateWithoutTimeout: resourceProvisioningArtifactUpdate,
		DeleteWithoutTimeout: resourceProvisioningArtifactDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ProvisioningArtifactReadyTimeout),
			Read:   schema.DefaultTimeout(ProvisioningArtifactReadTimeout),
			Update: schema.DefaultTimeout(ProvisioningArtifactUpdateTimeout),
			Delete: schema.DefaultTimeout(ProvisioningArtifactDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      acceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(acceptLanguage_Values(), false),
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"disable_template_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"guidance": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ProvisioningArtifactGuidanceDefault,
				ValidateDiagFunc: enum.Validate[awstypes.ProvisioningArtifactGuidance](),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provisioning_artifact_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template_physical_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"template_url",
					"template_physical_id",
				},
			},
			"template_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"template_url",
					"template_physical_id",
				},
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ProvisioningArtifactType](),
			},
		},
	}
}

func resourceProvisioningArtifactCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	parameters := make(map[string]any)
	parameters[names.AttrDescription] = d.Get(names.AttrDescription)
	parameters["disable_template_validation"] = d.Get("disable_template_validation")
	parameters[names.AttrName] = d.Get(names.AttrName)
	parameters["template_physical_id"] = d.Get("template_physical_id")
	parameters["template_url"] = d.Get("template_url")
	parameters[names.AttrType] = d.Get(names.AttrType)

	input := &servicecatalog.CreateProvisioningArtifactInput{
		IdempotencyToken: aws.String(id.UniqueId()),
		Parameters:       expandProvisioningArtifactParameters(parameters),
		ProductId:        aws.String(d.Get("product_id").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	var output *servicecatalog.CreateProvisioningArtifactOutput
	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error

		output, err = conn.CreateProvisioningArtifact(ctx, input)

		if errs.IsAErrorMessageContains[*awstypes.InvalidParametersException](err, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateProvisioningArtifact(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Provisioning Artifact: %s", err)
	}

	if output == nil || output.ProvisioningArtifactDetail == nil || output.ProvisioningArtifactDetail.Id == nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Provisioning Artifact: empty response")
	}

	d.SetId(provisioningArtifactID(aws.ToString(output.ProvisioningArtifactDetail.Id), d.Get("product_id").(string)))

	// Active and Guidance are not fields of CreateProvisioningArtifact but are fields of UpdateProvisioningArtifact.
	// In order to set these to non-default values, you must create and then update.

	return append(diags, resourceProvisioningArtifactUpdate(ctx, d, meta)...)
}

func resourceProvisioningArtifactRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	artifactID, productID, err := provisioningArtifactParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Service Catalog Provisioning Artifact ID (%s): %s", d.Id(), err)
	}

	output, err := waitProvisioningArtifactReady(ctx, conn, artifactID, productID, d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Service Catalog Provisioning Artifact (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Provisioning Artifact (%s): %s", d.Id(), err)
	}

	if output == nil || output.ProvisioningArtifactDetail == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Provisioning Artifact (%s): empty response", d.Id())
	}

	if v, ok := output.Info["ImportFromPhysicalId"]; ok {
		d.Set("template_physical_id", v)
	}

	if v, ok := output.Info["LoadTemplateFromURL"]; ok {
		d.Set("template_url", v)
	}

	pad := output.ProvisioningArtifactDetail

	d.Set("active", pad.Active)
	if pad.CreatedTime != nil {
		d.Set(names.AttrCreatedTime, pad.CreatedTime.Format(time.RFC3339))
	}
	d.Set(names.AttrDescription, pad.Description)
	d.Set("guidance", pad.Guidance)
	d.Set(names.AttrName, pad.Name)
	d.Set("product_id", productID)
	d.Set("provisioning_artifact_id", artifactID)
	d.Set(names.AttrType, pad.Type)

	return diags
}

func resourceProvisioningArtifactUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	if d.HasChanges("accept_language", "active", names.AttrDescription, "guidance", names.AttrName, "product_id") {
		artifactID, productID, err := provisioningArtifactParseID(d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing Service Catalog Provisioning Artifact ID (%s): %s", d.Id(), err)
		}

		input := &servicecatalog.UpdateProvisioningArtifactInput{
			ProductId:              aws.String(productID),
			ProvisioningArtifactId: aws.String(artifactID),
			Active:                 aws.Bool(d.Get("active").(bool)),
		}

		if v, ok := d.GetOk("accept_language"); ok {
			input.AcceptLanguage = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("guidance"); ok {
			input.Guidance = awstypes.ProvisioningArtifactGuidance(v.(string))
		}

		if v, ok := d.GetOk(names.AttrName); ok {
			input.Name = aws.String(v.(string))
		}

		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
			_, err := conn.UpdateProvisioningArtifact(ctx, input)

			if errs.IsAErrorMessageContains[*awstypes.InvalidParametersException](err, "profile does not exist") {
				return retry.RetryableError(err)
			}

			if err != nil {
				return retry.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateProvisioningArtifact(ctx, input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Service Catalog Provisioning Artifact (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceProvisioningArtifactRead(ctx, d, meta)...)
}

func resourceProvisioningArtifactDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	artifactID, productID, err := provisioningArtifactParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Service Catalog Provisioning Artifact ID (%s): %s", d.Id(), err)
	}

	input := &servicecatalog.DeleteProvisioningArtifactInput{
		ProductId:              aws.String(productID),
		ProvisioningArtifactId: aws.String(artifactID),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	_, err = conn.DeleteProvisioningArtifact(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Provisioning Artifact (%s): %s", d.Id(), err)
	}

	if err := waitProvisioningArtifactDeleted(ctx, conn, artifactID, productID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Provisioning Artifact (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
