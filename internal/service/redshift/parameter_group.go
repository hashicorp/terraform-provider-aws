package redshift

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceParameterGroupCreate,
		Read:   resourceParameterGroupRead,
		Update: resourceParameterGroupUpdate,
		Delete: resourceParameterGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				ForceNew: false,
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
				Set: resourceAwsRedshiftParameterHash,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceParameterGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	createOpts := redshift.CreateClusterParameterGroupInput{
		ParameterGroupName:   aws.String(d.Get("name").(string)),
		ParameterGroupFamily: aws.String(d.Get("family").(string)),
		Description:          aws.String(d.Get("description").(string)),
		Tags:                 tags.IgnoreAws().RedshiftTags(),
	}

	log.Printf("[DEBUG] Create Redshift Parameter Group: %#v", createOpts)
	_, err := conn.CreateClusterParameterGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Redshift Parameter Group: %s", err)
	}

	d.SetId(aws.StringValue(createOpts.ParameterGroupName))

	if v := d.Get("parameter").(*schema.Set); v.Len() > 0 {
		parameters := ExpandParameters(v.List())

		modifyOpts := redshift.ModifyClusterParameterGroupInput{
			ParameterGroupName: aws.String(d.Id()),
			Parameters:         parameters,
		}

		if _, err := conn.ModifyClusterParameterGroup(&modifyOpts); err != nil {
			return fmt.Errorf("error adding Redshift Parameter Group (%s) parameters: %s", d.Id(), err)
		}
	}

	return resourceParameterGroupRead(d, meta)
}

func resourceParameterGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	describeOpts := redshift.DescribeClusterParameterGroupsInput{
		ParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeClusterParameterGroups(&describeOpts)
	if err != nil {
		return err
	}

	if len(describeResp.ParameterGroups) != 1 ||
		aws.StringValue(describeResp.ParameterGroups[0].ParameterGroupName) != d.Id() {
		d.SetId("")
		return fmt.Errorf("Unable to find Parameter Group: %#v", describeResp.ParameterGroups)
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
	tags := tftags.RedshiftKeyValueTags(describeResp.ParameterGroups[0].Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	describeParametersOpts := redshift.DescribeClusterParametersInput{
		ParameterGroupName: aws.String(d.Id()),
		Source:             aws.String("user"),
	}

	describeParametersResp, err := conn.DescribeClusterParameters(&describeParametersOpts)
	if err != nil {
		return err
	}

	d.Set("parameter", FlattenParameters(describeParametersResp.Parameters))
	return nil
}

func resourceParameterGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

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
			_, err := conn.ModifyClusterParameterGroup(&modifyOpts)
			if err != nil {
				return fmt.Errorf("Error modifying Redshift Parameter Group: %s", err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.RedshiftUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Parameter Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceParameterGroupRead(d, meta)
}

func resourceParameterGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftConn

	_, err := conn.DeleteClusterParameterGroup(&redshift.DeleteClusterParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	})
	if err != nil && tfawserr.ErrMessageContains(err, "RedshiftParameterGroupNotFoundFault", "") {
		return nil
	}
	return err
}

func resourceAwsRedshiftParameterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	// Store the value as a lower case string, to match how we store them in FlattenParameters
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["value"].(string))))

	return create.StringHashcode(buf.String())
}
