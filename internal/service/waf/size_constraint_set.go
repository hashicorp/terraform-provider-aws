package waf

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceSizeConstraintSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSizeConstraintSetCreate,
		ReadWithoutTimeout:   resourceSizeConstraintSetRead,
		UpdateWithoutTimeout: resourceSizeConstraintSetUpdate,
		DeleteWithoutTimeout: resourceSizeConstraintSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: SizeConstraintSetSchema(),
	}
}

func resourceSizeConstraintSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn()

	log.Printf("[INFO] Creating SizeConstraintSet: %s", d.Get("name").(string))

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateSizeConstraintSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}

		return conn.CreateSizeConstraintSetWithContext(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SizeConstraintSet: %s", err)
	}
	resp := out.(*waf.CreateSizeConstraintSetOutput)

	d.SetId(aws.StringValue(resp.SizeConstraintSet.SizeConstraintSetId))

	return append(diags, resourceSizeConstraintSetUpdate(ctx, d, meta)...)
}

func resourceSizeConstraintSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn()
	log.Printf("[INFO] Reading SizeConstraintSet: %s", d.Get("name").(string))
	params := &waf.GetSizeConstraintSetInput{
		SizeConstraintSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetSizeConstraintSetWithContext(ctx, params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == waf.ErrCodeNonexistentItemException {
			log.Printf("[WARN] WAF SizeConstraintSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF Size Constraint Set (%s): %s", d.Get("name").(string), err)
	}

	d.Set("name", resp.SizeConstraintSet.Name)
	d.Set("size_constraints", FlattenSizeConstraints(resp.SizeConstraintSet.SizeConstraints))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("sizeconstraintset/%s", d.Id()),
	}
	d.Set("arn", arn.String())

	return diags
}

func resourceSizeConstraintSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn()

	if d.HasChange("size_constraints") {
		o, n := d.GetChange("size_constraints")
		oldConstraints, newConstraints := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateSizeConstraintSetResource(ctx, d.Id(), oldConstraints, newConstraints, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SizeConstraintSet: %s", err)
		}
	}

	return append(diags, resourceSizeConstraintSetRead(ctx, d, meta)...)
}

func resourceSizeConstraintSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFConn()

	oldConstraints := d.Get("size_constraints").(*schema.Set).List()

	if len(oldConstraints) > 0 {
		noConstraints := []interface{}{}
		err := updateSizeConstraintSetResource(ctx, d.Id(), oldConstraints, noConstraints, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting SizeConstraintSet: %s", err)
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteSizeConstraintSetInput{
			ChangeToken:         token,
			SizeConstraintSetId: aws.String(d.Id()),
		}
		return conn.DeleteSizeConstraintSetWithContext(ctx, req)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SizeConstraintSet: %s", err)
	}

	return diags
}

func updateSizeConstraintSetResource(ctx context.Context, id string, oldS, newS []interface{}, conn *waf.WAF) error {
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateSizeConstraintSetInput{
			ChangeToken:         token,
			SizeConstraintSetId: aws.String(id),
			Updates:             DiffSizeConstraints(oldS, newS),
		}

		log.Printf("[INFO] Updating WAF Size Constraint constraints: %s", req)
		return conn.UpdateSizeConstraintSetWithContext(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("Error updating SizeConstraintSet: %s", err)
	}

	return nil
}
