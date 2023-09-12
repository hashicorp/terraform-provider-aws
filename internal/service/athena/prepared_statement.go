// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_athena_prepared_statement", name="Prepared Statement")
func ResourcePreparedStatement() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePreparedStatementCreate,
		ReadWithoutTimeout:   resourcePreparedStatementRead,
		UpdateWithoutTimeout: resourcePreparedStatementUpdate,
		DeleteWithoutTimeout: resourcePreparedStatementDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected WORKGROUP-NAME/STATEMENT-NAME", d.Id())
				}
				workGroupName := idParts[0]
				statementName := idParts[1]
				d.Set("workgroup", workGroupName)
				d.SetId(statementName)
				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_@:]{1,256}$`), ""),
			},
			"query_statement": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 262144),
			},
			"workgroup": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9._-]{1,128}$`), ""),
			},
		},
	}
}

const (
	ResNamePreparedStatement = "Prepared Statement"
)

func resourcePreparedStatementCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AthenaConn(ctx)
	in := &athena.CreatePreparedStatementInput{
		QueryStatement: aws.String(d.Get("query_statement").(string)),
		StatementName:  aws.String(d.Get("name").(string)),
		WorkGroup:      aws.String(d.Get("workgroup").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	_, err := conn.CreatePreparedStatementWithContext(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.Athena, create.ErrActionCreating, ResNamePreparedStatement, d.Get("name").(string), err)...)
	}

	d.SetId(d.Get("name").(string))

	return append(diags, resourcePreparedStatementRead(ctx, d, meta)...)
}

func resourcePreparedStatementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AthenaConn(ctx)

	out, err := findPreparedStatement(ctx, conn, d.Id(), d.Get("workgroup").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Athena PreparedStatement (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.Athena, create.ErrActionReading, ResNamePreparedStatement, d.Id(), err)...)
	}

	d.Set("description", out.Description)
	d.Set("query_statement", out.QueryStatement)
	d.Set("name", out.StatementName)
	d.Set("workgroup", out.WorkGroupName)

	return diags
}

func resourcePreparedStatementUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AthenaConn(ctx)

	update := false

	in := &athena.UpdatePreparedStatementInput{
		StatementName: aws.String(d.Get("name").(string)),
		WorkGroup:     aws.String(d.Get("workgroup").(string)),
	}

	if d.HasChanges("query_statement") {
		in.QueryStatement = aws.String(d.Get("query_statement").(string))
		update = true
	}

	if d.HasChanges("description") {
		in.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating Athena PreparedStatement (%s): %#v", d.Id(), in)
	_, err := conn.UpdatePreparedStatementWithContext(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.Athena, create.ErrActionUpdating, ResNamePreparedStatement, d.Id(), err)...)
	}

	return append(diags, resourcePreparedStatementRead(ctx, d, meta)...)
}

func resourcePreparedStatementDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AthenaConn(ctx)

	log.Printf("[INFO] Deleting Athena PreparedStatement %s", d.Id())

	_, err := conn.DeletePreparedStatementWithContext(ctx, &athena.DeletePreparedStatementInput{
		StatementName: aws.String(d.Get("name").(string)),
		WorkGroup:     aws.String(d.Get("workgroup").(string)),
	})

	if tfawserr.ErrCodeEquals(err, athena.ErrCodeResourceNotFoundException) {
		return diags
	}
	if err != nil {
		return append(diags, create.DiagError(names.Athena, create.ErrActionDeleting, ResNamePreparedStatement, d.Id(), err)...)
	}

	return diags
}

func findPreparedStatement(ctx context.Context, conn *athena.Athena, name string, workgroup string) (*athena.PreparedStatement, error) {
	in := &athena.GetPreparedStatementInput{
		StatementName: aws.String(name),
		WorkGroup:     aws.String(workgroup),
	}
	out, err := conn.GetPreparedStatementWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, athena.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.PreparedStatement == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.PreparedStatement, nil
}
