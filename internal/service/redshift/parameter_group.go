package redshift

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterGroupCreate,
		ReadWithoutTimeout:   resourceParameterGroupRead,
		UpdateWithoutTimeout: resourceParameterGroupUpdate,
		DeleteWithoutTimeout: resourceParameterGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringMatch(regexp.MustCompile(`(?i)^[a-z]`), "first character must be a letter"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},

			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},

			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: resourceParameterHash,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	createOpts := redshift.CreateClusterParameterGroupInput{
		ParameterGroupName:   aws.String(d.Get("name").(string)),
		ParameterGroupFamily: aws.String(d.Get("family").(string)),
		Description:          aws.String(d.Get("description").(string)),
		Tags:                 Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Create Redshift Parameter Group: %#v", createOpts)
	_, err := conn.CreateClusterParameterGroupWithContext(ctx, &createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Parameter Group: %s", err)
	}

	d.SetId(aws.StringValue(createOpts.ParameterGroupName))

	if v := d.Get("parameter").(*schema.Set); v.Len() > 0 {
		parameters := ExpandParameters(v.List())

		modifyOpts := redshift.ModifyClusterParameterGroupInput{
			ParameterGroupName: aws.String(d.Id()),
			Parameters:         parameters,
		}

		if _, err := conn.ModifyClusterParameterGroupWithContext(ctx, &modifyOpts); err != nil {
			return sdkdiag.AppendErrorf(diags, "adding Redshift Parameter Group (%s) parameters: %s", d.Id(), err)
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	describeOpts := redshift.DescribeClusterParameterGroupsInput{
		ParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeClusterParameterGroupsWithContext(ctx, &describeOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Parameter Group (%s): %s", d.Id(), err)
	}

	if len(describeResp.ParameterGroups) != 1 ||
		aws.StringValue(describeResp.ParameterGroups[0].ParameterGroupName) != d.Id() {
		d.SetId("")
		return sdkdiag.AppendErrorf(diags, "Unable to find Parameter Group: %#v", describeResp.ParameterGroups)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("parametergroup:%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	d.Set("name", describeResp.ParameterGroups[0].ParameterGroupName)
	d.Set("family", describeResp.ParameterGroups[0].ParameterGroupFamily)
	d.Set("description", describeResp.ParameterGroups[0].Description)
	tags := KeyValueTags(describeResp.ParameterGroups[0].Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	describeParametersOpts := redshift.DescribeClusterParametersInput{
		ParameterGroupName: aws.String(d.Id()),
		Source:             aws.String("user"),
	}

	describeParametersResp, err := conn.DescribeClusterParametersWithContext(ctx, &describeParametersOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set("parameter", FlattenParameters(describeParametersResp.Parameters))
	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()

	if d.HasChange("parameter") {
		o, n := d.GetChange("parameter")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Expand the "parameter" set to aws-sdk-go compat []redshift.Parameter
		parameters := ExpandParameters(ns.Difference(os).List())

		if len(parameters) > 0 {
			modifyOpts := redshift.ModifyClusterParameterGroupInput{
				ParameterGroupName: aws.String(d.Get("name").(string)),
				Parameters:         parameters,
			}

			log.Printf("[DEBUG] Modify Redshift Parameter Group: %s", modifyOpts)
			_, err := conn.ModifyClusterParameterGroupWithContext(ctx, &modifyOpts)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying Redshift Parameter Group: %s", err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Parameter Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn()

	_, err := conn.DeleteClusterParameterGroupWithContext(ctx, &redshift.DeleteClusterParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	})
	if err != nil && tfawserr.ErrCodeEquals(err, "RedshiftParameterGroupNotFoundFault") {
		return diags
	}
	return sdkdiag.AppendErrorf(diags, "deleting Redshift Parameter Group (%s): %s", d.Id(), err)
}

func resourceParameterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	// Store the value as a lower case string, to match how we store them in FlattenParameters
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["value"].(string))))

	return create.StringHashcode(buf.String())
}
