// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/schemas"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_schemas_schema", name="Schema")
// @Tags(identifierAttribute="arn")
func ResourceSchema() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSchemaCreate,
		ReadWithoutTimeout:   resourceSchemaRead,
		UpdateWithoutTimeout: resourceSchemaUpdate,
		DeleteWithoutTimeout: resourceSchemaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"content": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},

			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 385),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z@]+`), ""),
				),
			},

			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(schemas.Type_Values(), true),
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"version_created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSchemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	name := d.Get("name").(string)
	registryName := d.Get("registry_name").(string)
	input := &schemas.CreateSchemaInput{
		Content:      aws.String(d.Get("content").(string)),
		RegistryName: aws.String(registryName),
		SchemaName:   aws.String(name),
		Tags:         getTagsIn(ctx),
		Type:         aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	id := SchemaCreateResourceID(name, registryName)

	log.Printf("[DEBUG] Creating EventBridge Schemas Schema: %s", input)
	_, err := conn.CreateSchemaWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Schemas Schema (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceSchemaRead(ctx, d, meta)...)
}

func resourceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	name, registryName, err := SchemaParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing EventBridge Schemas Schema ID: %s", err)
	}

	output, err := FindSchemaByNameAndRegistryName(ctx, conn, name, registryName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Schema (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Schemas Schema (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.SchemaArn)
	d.Set("content", output.Content)
	d.Set("description", output.Description)
	if output.LastModified != nil {
		d.Set("last_modified", aws.TimeValue(output.LastModified).Format(time.RFC3339))
	} else {
		d.Set("last_modified", nil)
	}
	d.Set("name", output.SchemaName)
	d.Set("registry_name", registryName)
	d.Set("type", output.Type)
	d.Set("version", output.SchemaVersion)
	if output.VersionCreatedDate != nil {
		d.Set("version_created_date", aws.TimeValue(output.VersionCreatedDate).Format(time.RFC3339))
	} else {
		d.Set("version_created_date", nil)
	}

	return diags
}

func resourceSchemaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	if d.HasChanges("content", "description", "type") {
		name, registryName, err := SchemaParseResourceID(d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing EventBridge Schemas Schema ID: %s", err)
		}

		input := &schemas.UpdateSchemaInput{
			RegistryName: aws.String(registryName),
			SchemaName:   aws.String(name),
		}

		if d.HasChanges("content", "type") {
			input.Content = aws.String(d.Get("content").(string))
			input.Type = aws.String(d.Get("type").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		log.Printf("[DEBUG] Updating EventBridge Schemas Schema: %s", input)
		_, err = conn.UpdateSchemaWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Schemas Schema (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSchemaRead(ctx, d, meta)...)
}

func resourceSchemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasConn(ctx)

	name, registryName, err := SchemaParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing EventBridge Schemas Schema ID: %s", err)
	}

	log.Printf("[INFO] Deleting EventBridge Schemas Schema (%s)", d.Id())
	_, err = conn.DeleteSchemaWithContext(ctx, &schemas.DeleteSchemaInput{
		RegistryName: aws.String(registryName),
		SchemaName:   aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, schemas.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Schemas Schema (%s): %s", d.Id(), err)
	}

	return diags
}
