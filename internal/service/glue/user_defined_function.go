// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_user_defined_function")
func ResourceUserDefinedFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserDefinedFunctionCreate,
		ReadWithoutTimeout:   resourceUserDefinedFunctionRead,
		UpdateWithoutTimeout: resourceUserDefinedFunctionUpdate,
		DeleteWithoutTimeout: resourceUserDefinedFunctionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"class_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"owner_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"owner_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PrincipalType](),
			},
			"resource_uris": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1000,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrResourceType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ResourceType](),
						},
						names.AttrURI: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUserDefinedFunctionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)
	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)
	dbName := d.Get(names.AttrDatabaseName).(string)
	funcName := d.Get(names.AttrName).(string)

	input := &glue.CreateUserDefinedFunctionInput{
		CatalogId:     aws.String(catalogID),
		DatabaseName:  aws.String(dbName),
		FunctionInput: expandUserDefinedFunctionInput(d),
	}

	_, err := conn.CreateUserDefinedFunction(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue User Defined Function: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", catalogID, dbName, funcName))

	return append(diags, resourceUserDefinedFunctionRead(ctx, d, meta)...)
}

func resourceUserDefinedFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, funcName, err := ReadUDFID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Glue User Defined Function (%s): %s", d.Id(), err)
	}

	input := &glue.UpdateUserDefinedFunctionInput{
		CatalogId:     aws.String(catalogID),
		DatabaseName:  aws.String(dbName),
		FunctionName:  aws.String(funcName),
		FunctionInput: expandUserDefinedFunctionInput(d),
	}

	if _, err := conn.UpdateUserDefinedFunction(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Glue User Defined Function (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUserDefinedFunctionRead(ctx, d, meta)...)
}

func resourceUserDefinedFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, funcName, err := ReadUDFID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue User Defined Function (%s): %s", d.Id(), err)
	}

	input := &glue.GetUserDefinedFunctionInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		FunctionName: aws.String(funcName),
	}

	out, err := conn.GetUserDefinedFunction(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			log.Printf("[WARN] Glue User Defined Function (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading Glue User Defined Function (%s): %s", d.Id(), err)
	}

	udf := out.UserDefinedFunction

	udfArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("userDefinedFunction/%s/%s", dbName, aws.ToString(udf.FunctionName)),
	}.String()

	d.Set(names.AttrARN, udfArn)
	d.Set(names.AttrName, udf.FunctionName)
	d.Set(names.AttrCatalogID, catalogID)
	d.Set(names.AttrDatabaseName, dbName)
	d.Set("owner_type", udf.OwnerType)
	d.Set("owner_name", udf.OwnerName)
	d.Set("class_name", udf.ClassName)
	if udf.CreateTime != nil {
		d.Set(names.AttrCreateTime, udf.CreateTime.Format(time.RFC3339))
	}
	if err := d.Set("resource_uris", flattenUserDefinedFunctionResourceURI(udf.ResourceUris)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue User Defined Function (%s): setting resource_uris: %s", d.Id(), err)
	}

	return diags
}

func resourceUserDefinedFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)
	catalogID, dbName, funcName, err := ReadUDFID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue User Defined Function (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Glue User Defined Function: %s", d.Id())
	_, err = conn.DeleteUserDefinedFunction(ctx, &glue.DeleteUserDefinedFunctionInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		FunctionName: aws.String(funcName),
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue User Defined Function (%s): %s", d.Id(), err)
	}
	return diags
}

func ReadUDFID(id string) (catalogID string, dbName string, funcName string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected CATALOG-ID:DATABASE-NAME:FUNCTION-NAME", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func expandUserDefinedFunctionInput(d *schema.ResourceData) *awstypes.UserDefinedFunctionInput {
	udf := &awstypes.UserDefinedFunctionInput{
		ClassName:    aws.String(d.Get("class_name").(string)),
		FunctionName: aws.String(d.Get(names.AttrName).(string)),
		OwnerName:    aws.String(d.Get("owner_name").(string)),
		OwnerType:    awstypes.PrincipalType(d.Get("owner_type").(string)),
	}

	if v, ok := d.GetOk("resource_uris"); ok && v.(*schema.Set).Len() > 0 {
		udf.ResourceUris = expandUserDefinedFunctionResourceURI(d.Get("resource_uris").(*schema.Set))
	}

	return udf
}

func expandUserDefinedFunctionResourceURI(conf *schema.Set) []awstypes.ResourceUri {
	result := make([]awstypes.ResourceUri, 0, conf.Len())

	for _, r := range conf.List() {
		uriRaw, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		uri := awstypes.ResourceUri{
			ResourceType: awstypes.ResourceType(uriRaw[names.AttrResourceType].(string)),
			Uri:          aws.String(uriRaw[names.AttrURI].(string)),
		}

		result = append(result, uri)
	}

	return result
}

func flattenUserDefinedFunctionResourceURI(uris []awstypes.ResourceUri) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(uris))

	for _, i := range uris {
		l := map[string]interface{}{
			names.AttrResourceType: string(i.ResourceType),
			names.AttrURI:          aws.ToString(i.Uri),
		}

		result = append(result, l)
	}
	return result
}
