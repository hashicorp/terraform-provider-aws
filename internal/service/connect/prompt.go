// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_prompt", name="Prompt")
// @Tags(identifierAttribute="arn")
func ResourcePrompt() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourcePromptCreate,
		ReadWithoutTimeout:   ResourcePromptRead,
		UpdateWithoutTimeout: ResourcePromptUpdate,
		DeleteWithoutTimeout: ResourcePromptDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prompt_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 250),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 127),
			},
			"s3_uri": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
					validation.StringLenBetween(1, 512),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func ResourcePromptCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Get("instance_id").(string)
	name := d.Get("name").(string)

	input := &connect.CreatePromptInput{
		InstanceId: aws.String(instanceID),
		Name:       aws.String(name),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("s3_uri"); ok {
		input.S3Uri = aws.String(v.(string))
	}

	output, err := conn.CreatePromptWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Connect Prompt (%s): %s", name, err)
	}

	if output == nil {
		return diag.Errorf("creating Connect Prompt (%s): empty output", name)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.PromptId)))

	return ResourcePromptRead(ctx, d, meta)
}

func ResourcePromptRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, promptID, err := PromptParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribePromptWithContext(ctx, &connect.DescribePromptInput{
		InstanceId: aws.String(instanceID),
		PromptId:   aws.String(promptID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Prompt (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("getting Connect Prompt (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.Prompt == nil {
		return diag.Errorf("getting Connect Prompt (%s): empty response", d.Id())
	}

	d.Set("arn", resp.Prompt.PromptARN)
	d.Set("instance_id", aws.String(instanceID))
	d.Set("name", resp.Prompt.Name)
	d.Set("description", resp.Prompt.Description)
	d.Set("prompt_id", aws.String(promptID))

	setTagsOut(ctx, resp.Prompt.Tags)

	return nil
}

func ResourcePromptUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, promptID, err := PromptParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &connect.UpdatePromptInput{
		InstanceId: aws.String(instanceID),
		PromptId:   aws.String(promptID),
	}

	if d.HasChange("name") {
		input.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("s3_uri") {
		input.S3Uri = aws.String(d.Get("s3_uri").(string))
	}

	_, err = conn.UpdatePromptWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating Prompt (%s): %s", d.Id(), err)
	}

	return ResourcePromptRead(ctx, d, meta)
}

func ResourcePromptDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, promptID, err := PromptParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeletePromptWithContext(ctx, &connect.DeletePromptInput{
		InstanceId: aws.String(instanceID),
		PromptId:   aws.String(promptID),
	})

	if err != nil {
		return diag.Errorf("deleting Prompt (%s): %s", d.Id(), err)
	}

	return nil
}

func PromptParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:promptID", id)
	}

	return parts[0], parts[1], nil
}
