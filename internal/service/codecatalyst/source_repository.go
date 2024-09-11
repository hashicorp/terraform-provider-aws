// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecatalyst

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_codecatalyst_source_repository", name="Source Repository")
func ResourceSourceRepository() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSourceRepositoryCreate,
		ReadWithoutTimeout:   resourceSourceRepositoryRead,
		UpdateWithoutTimeout: resourceSourceRepositoryCreate,
		DeleteWithoutTimeout: resourceSourceRepositoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"space_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

const (
	ResNameSourceRepository = "Source Repository"
)

func resourceSourceRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)

	in := &codecatalyst.CreateSourceRepositoryInput{
		Name:        aws.String(d.Get(names.AttrName).(string)),
		ProjectName: aws.String(d.Get("project_name").(string)),
		SpaceName:   aws.String(d.Get("space_name").(string)),
	}

	out, err := conn.CreateSourceRepository(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionCreating, ResNameSourceRepository, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.Name == nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionCreating, ResNameSourceRepository, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Name))

	return append(diags, resourceSourceRepositoryRead(ctx, d, meta)...)
}

func resourceSourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)

	projectName := aws.String(d.Get("project_name").(string))
	spaceName := aws.String(d.Get("space_name").(string))

	out, err := findSourceRepositoryByName(ctx, conn, d.Id(), projectName, spaceName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeCatalyst SourceRepository (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionReading, ResNameSourceRepository, d.Id(), err)
	}

	d.Set(names.AttrName, out.Name)
	d.Set("project_name", out.ProjectName)
	d.Set("space_name", out.SpaceName)
	d.Set(names.AttrDescription, out.Description)

	return diags
}

func resourceSourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)

	log.Printf("[INFO] Deleting CodeCatalyst SourceRepository %s", d.Id())

	_, err := conn.DeleteSourceRepository(ctx, &codecatalyst.DeleteSourceRepositoryInput{
		Name:        aws.String(d.Id()),
		ProjectName: aws.String(d.Get("project_name").(string)),
		SpaceName:   aws.String(d.Get("space_name").(string)),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.CodeCatalyst, create.ErrActionDeleting, ResNameSourceRepository, d.Id(), err)
	}

	return diags
}

func findSourceRepositoryByName(ctx context.Context, conn *codecatalyst.Client, name string, projectName, spaceName *string) (*codecatalyst.GetSourceRepositoryOutput, error) {
	in := &codecatalyst.GetSourceRepositoryInput{
		Name:        aws.String(name),
		ProjectName: projectName,
		SpaceName:   spaceName,
	}
	out, err := conn.GetSourceRepository(ctx, in)
	if errs.IsA[*types.AccessDeniedException](err) || errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.Name == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
