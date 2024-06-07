// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/schemas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_schemas_schema", name="Schema")
// @Tags(identifierAttribute="arn")
func resourceSchema() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSchemaCreate,
		ReadWithoutTimeout:   resourceSchemaRead,
		UpdateWithoutTimeout: resourceSchemaUpdate,
		DeleteWithoutTimeout: resourceSchemaDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContent: {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 385),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z_.@-]+`), ""),
				),
			},
			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Type](),
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSchemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	name := d.Get(names.AttrName).(string)
	registryName := d.Get("registry_name").(string)
	input := &schemas.CreateSchemaInput{
		Content:      aws.String(d.Get(names.AttrContent).(string)),
		RegistryName: aws.String(registryName),
		SchemaName:   aws.String(name),
		Tags:         getTagsIn(ctx),
		Type:         awstypes.Type(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	id := schemaCreateResourceID(name, registryName)
	_, err := conn.CreateSchema(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Schemas Schema (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceSchemaRead(ctx, d, meta)...)
}

func resourceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	name, registryName, err := schemaParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findSchemaByTwoPartKey(ctx, conn, name, registryName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Schemas Schema (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Schemas Schema (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.SchemaArn)
	d.Set(names.AttrContent, output.Content)
	d.Set(names.AttrDescription, output.Description)
	if output.LastModified != nil {
		d.Set("last_modified", aws.ToTime(output.LastModified).Format(time.RFC3339))
	} else {
		d.Set("last_modified", nil)
	}
	d.Set(names.AttrName, output.SchemaName)
	d.Set("registry_name", registryName)
	d.Set(names.AttrType, output.Type)
	d.Set(names.AttrVersion, output.SchemaVersion)
	if output.VersionCreatedDate != nil {
		d.Set("version_created_date", aws.ToTime(output.VersionCreatedDate).Format(time.RFC3339))
	} else {
		d.Set("version_created_date", nil)
	}

	return diags
}

func resourceSchemaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	if d.HasChanges(names.AttrContent, names.AttrDescription, names.AttrType) {
		name, registryName, err := schemaParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &schemas.UpdateSchemaInput{
			RegistryName: aws.String(registryName),
			SchemaName:   aws.String(name),
		}

		if d.HasChanges(names.AttrContent, names.AttrType) {
			input.Content = aws.String(d.Get(names.AttrContent).(string))
			input.Type = awstypes.Type(d.Get(names.AttrType).(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		_, err = conn.UpdateSchema(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EventBridge Schemas Schema (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSchemaRead(ctx, d, meta)...)
}

func resourceSchemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SchemasClient(ctx)

	name, registryName, err := schemaParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting EventBridge Schemas Schema (%s)", d.Id())
	_, err = conn.DeleteSchema(ctx, &schemas.DeleteSchemaInput{
		RegistryName: aws.String(registryName),
		SchemaName:   aws.String(name),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Schemas Schema (%s): %s", d.Id(), err)
	}

	return diags
}

const schemaResourceIDSeparator = "/"

func schemaCreateResourceID(schemaName, registryName string) string {
	parts := []string{schemaName, registryName}
	id := strings.Join(parts, schemaResourceIDSeparator)

	return id
}

func schemaParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, schemaResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SCHEMA_NAME%[2]sREGISTRY_NAME", id, schemaResourceIDSeparator)
}

func findSchemaByTwoPartKey(ctx context.Context, conn *schemas.Client, name, registryName string) (*schemas.DescribeSchemaOutput, error) {
	input := &schemas.DescribeSchemaInput{
		RegistryName: aws.String(registryName),
		SchemaName:   aws.String(name),
	}

	output, err := conn.DescribeSchema(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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
