// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_athena_data_catalog", name="Data Catalog")
// @Tags(identifierAttribute="arn")
func resourceDataCatalog() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataCatalogCreate,
		ReadWithoutTimeout:   resourceDataCatalogRead,
		UpdateWithoutTimeout: resourceDataCatalogUpdate,
		DeleteWithoutTimeout: resourceDataCatalogDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 129),
					validation.StringMatch(regexache.MustCompile(`[\w@-]*`), ""),
				),
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ValidateDiagFunc: validation.AllDiag(
					validation.MapKeyLenBetween(1, 255),
					validation.MapValueLenBetween(0, 51200),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.DataCatalogType](),
			},
		},
	}
}

func resourceDataCatalogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &athena.CreateDataCatalogInput{
		Name:        aws.String(name),
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		Tags:        getTagsIn(ctx),
		Type:        types.DataCatalogType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrParameters); ok && len(v.(map[string]interface{})) > 0 {
		input.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	_, err := conn.CreateDataCatalog(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Athena Data Catalog (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDataCatalogRead(ctx, d, meta)...)
}

func resourceDataCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	dataCatalog, err := findDataCatalogByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Athena Data Catalog (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Athena Data Catalog (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "athena",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("datacatalog/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, dataCatalog.Description)
	d.Set(names.AttrName, dataCatalog.Name)
	d.Set(names.AttrType, dataCatalog.Type)

	// NOTE: This is a workaround for the fact that the API sets default values for parameters that are not set.
	// Because the API sets default values, what's returned by the API is different than what's set by the user.
	if v, ok := d.GetOk(names.AttrParameters); ok && len(v.(map[string]interface{})) > 0 {
		parameters := make(map[string]string, 0)

		for key, val := range v.(map[string]interface{}) {
			if v, ok := dataCatalog.Parameters[key]; ok {
				parameters[key] = v
			} else {
				parameters[key] = val.(string)
			}
		}

		d.Set(names.AttrParameters, parameters)
	} else {
		d.Set(names.AttrParameters, nil)
	}

	return diags
}

func resourceDataCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &athena.UpdateDataCatalogInput{
			Name:        aws.String(d.Id()),
			Type:        types.DataCatalogType(d.Get(names.AttrType).(string)),
			Description: aws.String(d.Get(names.AttrDescription).(string)),
		}

		if d.HasChange(names.AttrParameters) {
			if v, ok := d.GetOk(names.AttrParameters); ok && len(v.(map[string]interface{})) > 0 {
				input.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
			}
		}

		_, err := conn.UpdateDataCatalog(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Athena Data Catalog (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataCatalogRead(ctx, d, meta)...)
}

func resourceDataCatalogDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	log.Printf("[DEBUG] Deleting Athena Data Catalog: (%s)", d.Id())
	_, err := conn.DeleteDataCatalog(ctx, &athena.DeleteDataCatalogInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Athena Data Catalog (%s): %s", d.Id(), err)
	}

	return diags
}

func findDataCatalogByName(ctx context.Context, conn *athena.Client, name string) (*types.DataCatalog, error) {
	input := &athena.GetDataCatalogInput{
		Name: aws.String(name),
	}

	output, err := conn.GetDataCatalog(ctx, input)

	if errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "was not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DataCatalog == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DataCatalog, nil
}
