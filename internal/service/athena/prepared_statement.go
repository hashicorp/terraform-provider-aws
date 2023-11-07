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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_athena_prepared_statement", name="Prepared Statement")
func resourcePreparedStatement() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePreparedStatementCreate,
		ReadWithoutTimeout:   resourcePreparedStatementRead,
		UpdateWithoutTimeout: resourcePreparedStatementUpdate,
		DeleteWithoutTimeout: resourcePreparedStatementDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourcePreparedStatementCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	workGroupName, statementName := d.Get("workgroup").(string), d.Get("name").(string)
	id := preparedStatementCreateResourceID(workGroupName, statementName)
	input := &athena.CreatePreparedStatementInput{
		QueryStatement: aws.String(d.Get("query_statement").(string)),
		StatementName:  aws.String(statementName),
		WorkGroup:      aws.String(workGroupName),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreatePreparedStatement(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Athena Prepared Statement (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePreparedStatementRead(ctx, d, meta)...)
}

func resourcePreparedStatementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	workGroupName, statementName, err := preparedStatementParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	preparedStatement, err := findPreparedStatementByTwoPartKey(ctx, conn, workGroupName, statementName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Athena Prepared Statement (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Athena Prepared Statement (%s): %s", d.Id(), err)
	}

	d.Set("description", preparedStatement.Description)
	d.Set("name", preparedStatement.StatementName)
	d.Set("query_statement", preparedStatement.QueryStatement)
	d.Set("workgroup", preparedStatement.WorkGroupName)

	return diags
}

func resourcePreparedStatementUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	workGroupName, statementName, err := preparedStatementParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &athena.UpdatePreparedStatementInput{
		StatementName: aws.String(statementName),
		WorkGroup:     aws.String(workGroupName),
	}

	if d.HasChanges("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChanges("query_statement") {
		input.QueryStatement = aws.String(d.Get("query_statement").(string))
	}

	_, err = conn.UpdatePreparedStatement(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Athena Prepared Statement (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePreparedStatementRead(ctx, d, meta)...)
}

func resourcePreparedStatementDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	workGroupName, statementName, err := preparedStatementParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Athena Prepared Statement: %s", d.Id())
	_, err = conn.DeletePreparedStatement(ctx, &athena.DeletePreparedStatementInput{
		StatementName: aws.String(statementName),
		WorkGroup:     aws.String(workGroupName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Athena Prepaed Statement (%s): %s", d.Id(), err)
	}

	return diags
}

func findPreparedStatementByTwoPartKey(ctx context.Context, conn *athena.Client, workGroupName, statementName string) (*types.PreparedStatement, error) {
	input := &athena.GetPreparedStatementInput{
		StatementName: aws.String(statementName),
		WorkGroup:     aws.String(workGroupName),
	}

	output, err := conn.GetPreparedStatement(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "WorkGroup is not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PreparedStatement == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PreparedStatement, nil
}

const preparedStatementResourceIDSeparator = "/"

func preparedStatementCreateResourceID(workGroupName, statementName string) string {
	parts := []string{workGroupName, statementName}
	id := strings.Join(parts, preparedStatementResourceIDSeparator)

	return id
}

func preparedStatementParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, preparedStatementResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected WORKGROUP-NAME%[2]sSTATEMENT-NAME", id, preparedStatementResourceIDSeparator)
}
