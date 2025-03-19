// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_component", name="Component")
// @Tags(identifierAttribute="id")
func resourceComponent() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceComponentCreate,
		ReadWithoutTimeout:   resourceComponentRead,
		UpdateWithoutTimeout: resourceComponentUpdate,
		DeleteWithoutTimeout: resourceComponentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"change_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"data": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"data", names.AttrURI},
				ValidateFunc: validation.StringLenBetween(1, 16000),
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Platform](),
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"supported_os_versions": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				MinItems: 1,
				MaxItems: 25,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrURI: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"data", names.AttrURI},
			},
			names.AttrVersion: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceComponentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateComponentInput{
		ClientToken: aws.String(id.UniqueId()),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("change_description"); ok {
		input.ChangeDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data"); ok {
		input.Data = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("platform"); ok {
		input.Platform = awstypes.Platform(v.(string))
	}

	if v, ok := d.GetOk("supported_os_versions"); ok && v.(*schema.Set).Len() > 0 {
		input.SupportedOsVersions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrURI); ok {
		input.Uri = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.SemanticVersion = aws.String(v.(string))
	}

	output, err := conn.CreateComponent(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Component: %s", err)
	}

	d.SetId(aws.ToString(output.ComponentBuildVersionArn))

	return append(diags, resourceComponentRead(ctx, d, meta)...)
}

func resourceComponentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	component, err := findComponentByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Image Builder Component (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Component (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, component.Arn)
	d.Set("change_description", component.ChangeDescription)
	d.Set("data", component.Data)
	d.Set("date_created", component.DateCreated)
	d.Set(names.AttrDescription, component.Description)
	d.Set(names.AttrEncrypted, component.Encrypted)
	d.Set(names.AttrKMSKeyID, component.KmsKeyId)
	d.Set(names.AttrName, component.Name)
	d.Set(names.AttrOwner, component.Owner)
	d.Set("platform", component.Platform)
	d.Set("supported_os_versions", component.SupportedOsVersions)
	d.Set(names.AttrType, component.Type)
	d.Set(names.AttrVersion, component.Version)

	setTagsOut(ctx, component.Tags)

	return diags
}

func resourceComponentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceComponentRead(ctx, d, meta)...)
}

func resourceComponentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Image Builder Component version %q", d.Id())
		return diags
	}

	log.Printf("[DEBUG] Deleting Image Builder Component: %s", d.Id())
	_, err := conn.DeleteComponent(ctx, &imagebuilder.DeleteComponentInput{
		ComponentBuildVersionArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Component (%s): %s", d.Id(), err)
	}

	return diags
}

func findComponentByARN(ctx context.Context, conn *imagebuilder.Client, arn string) (*awstypes.Component, error) {
	input := &imagebuilder.GetComponentInput{
		ComponentBuildVersionArn: aws.String(arn),
	}

	output, err := conn.GetComponent(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Component == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Component, nil
}
