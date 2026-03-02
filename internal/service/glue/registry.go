// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package glue

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_registry", name="Registry")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(preIdentityVersion="v6.3.0")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/glue;glue.GetRegistryOutput")
func resourceRegistry() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryCreate,
		ReadWithoutTimeout:   resourceRegistryRead,
		UpdateWithoutTimeout: resourceRegistryUpdate,
		DeleteWithoutTimeout: resourceRegistryDelete,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"registry_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_$#-]+$`), ""),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRegistryCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	input := &glue.CreateRegistryInput{
		RegistryName: aws.String(d.Get("registry_name").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Registry: %+v", input)
	output, err := conn.CreateRegistry(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Registry: %s", err)
	}
	d.SetId(aws.ToString(output.RegistryArn))

	return append(diags, resourceRegistryRead(ctx, d, meta)...)
}

func resourceRegistryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	output, err := findRegistryByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Registry (%s): %s", d.Id(), err)
	}

	if output == nil {
		log.Printf("[WARN] Glue Registry (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := aws.ToString(output.RegistryArn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("registry_name", output.RegistryName)

	return diags
}

func resourceRegistryUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChanges(names.AttrDescription) {
		input := &glue.UpdateRegistryInput{
			RegistryId: createRegistryID(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Glue Registry: %#v", input)
		_, err := conn.UpdateRegistry(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Registry (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRegistryRead(ctx, d, meta)...)
}

func resourceRegistryDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Deleting Glue Registry: %s", d.Id())
	input := &glue.DeleteRegistryInput{
		RegistryId: createRegistryID(d.Id()),
	}

	_, err := conn.DeleteRegistry(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Glue Registry (%s): %s", d.Id(), err)
	}

	_, err = waitRegistryDeleted(ctx, conn, d.Id())
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for Glue Registry (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func findRegistryByID(ctx context.Context, conn *glue.Client, id string) (*glue.GetRegistryOutput, error) {
	input := &glue.GetRegistryInput{
		RegistryId: createRegistryID(id),
	}

	output, err := conn.GetRegistry(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusRegistry(conn *glue.Client, id string) retry.StateRefreshFunc {
	const registryStatusUnknown = "Unknown"
	return func(ctx context.Context) (any, string, error) {
		output, err := findRegistryByID(ctx, conn, id)
		if err != nil {
			return nil, registryStatusUnknown, err
		}

		if output == nil {
			return output, registryStatusUnknown, nil
		}

		return output, string(output.Status), nil
	}
}

// waitRegistryDeleted waits for a Registry to return Deleted
func waitRegistryDeleted(ctx context.Context, conn *glue.Client, registryID string) (*glue.GetRegistryOutput, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegistryStatusDeleting),
		Target:  []string{},
		Refresh: statusRegistry(conn, registryID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetRegistryOutput); ok {
		return output, err
	}

	return nil, err
}
