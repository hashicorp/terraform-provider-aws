// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codestarconnections

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codestarconnections"
	"github.com/aws/aws-sdk-go-v2/service/codestarconnections/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codestarconnections_connection", name="Connection")
// @Tags(identifierAttribute="arn")
func resourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{"provider_type"},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provider_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.ProviderType](),
				ConflictsWith:    []string{"host_arn"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeStarConnectionsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &codestarconnections.CreateConnectionInput{
		ConnectionName: aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("host_arn"); ok {
		input.HostArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provider_type"); ok {
		input.ProviderType = types.ProviderType(v.(string))
	}

	output, err := conn.CreateConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeStar Connections Connection (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ConnectionArn))

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeStarConnectionsClient(ctx)

	connection, err := findConnectionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeStar Connections Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeStar Connections Connection (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(connection.ConnectionArn)
	d.SetId(arn)
	d.Set(names.AttrARN, connection.ConnectionArn)
	d.Set("connection_status", connection.ConnectionStatus)
	d.Set("host_arn", connection.HostArn)
	d.Set(names.AttrName, connection.ConnectionName)
	d.Set("provider_type", connection.ProviderType)

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeStarConnectionsClient(ctx)

	log.Printf("[DEBUG] Deleting CodeStar Connections Connection: %s", d.Id())
	_, err := conn.DeleteConnection(ctx, &codestarconnections.DeleteConnectionInput{
		ConnectionArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeStar Connections Connection (%s): %s", d.Id(), err)
	}

	return diags
}

func findConnectionByARN(ctx context.Context, conn *codestarconnections.Client, arn string) (*types.Connection, error) {
	input := &codestarconnections.GetConnectionInput{
		ConnectionArn: aws.String(arn),
	}

	output, err := conn.GetConnection(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Connection == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Connection, nil
}
