// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplestructfield

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplestructfield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplestructfield/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_examplestructfield_database", name="Database")
func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDatabaseCreate,
		ReadWithoutTimeout:   resourceDatabaseRead,
		UpdateWithoutTimeout: resourceDatabaseUpdate,
		DeleteWithoutTimeout: resourceDatabaseDelete,

		Schema: map[string]*schema.Schema{
			"engine_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parent_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.ExampleStructFieldClient(ctx)

	if v, ok := d.GetOk("parent_id"); ok {
		if _, err := newParentLookup(c.ExampleCrossServiceClient(ctx)).Find(ctx, v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Thing (%s): %s", v, err)
		}
	}

	name := d.Get("name").(string)
	input := examplestructfield.CreateDatabaseInput{
		EngineVersion: aws.String(d.Get("engine_version").(string)),
		Name:          aws.String(name),
	}

	output, err := conn.CreateDatabase(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Database (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Database.DatabaseId))

	return append(diags, resourceDatabaseRead(ctx, d, meta)...)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleStructFieldClient(ctx)

	database, err := findDatabaseByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Database (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Database (%s): %s", d.Id(), err)
	}

	d.Set("engine_version", database.EngineVersion)
	d.Set("name", database.Name)

	return diags
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleStructFieldClient(ctx)

	if d.HasChange("engine_version") {
		orchestrator := newDeploymentOrchestrator(conn)
		defer orchestrator.CleanUp(ctx)

		input := examplestructfield.CreateDeploymentInput{
			DatabaseId:          aws.String(d.Id()),
			TargetEngineVersion: aws.String(d.Get("engine_version").(string)),
		}

		deployment, err := orchestrator.CreateDeployment(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Database (%s): %s", d.Id(), err)
		}

		if err := orchestrator.Switchover(ctx, aws.ToString(deployment.DeploymentId)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Database (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDatabaseRead(ctx, d, meta)...)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleStructFieldClient(ctx)

	log.Printf("[DEBUG] Deleting Database: %s", d.Id())
	input := examplestructfield.DeleteDatabaseInput{
		DatabaseId: aws.String(d.Id()),
	}
	_, err := conn.DeleteDatabase(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Database (%s): %s", d.Id(), err)
	}

	return diags
}

func findDatabaseByID(ctx context.Context, conn *examplestructfield.Client, id string) (*awstypes.Database, error) {
	input := examplestructfield.DescribeDatabaseInput{
		DatabaseId: aws.String(id),
	}

	output, err := conn.DescribeDatabase(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Database == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Database, nil
}
