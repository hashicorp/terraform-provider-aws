package redshiftserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceWorkgroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkgroupCreate,
		Read:   resourceWorkgroupRead,
		Update: resourceWorkgroupUpdate,
		Delete: resourceWorkgroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workgroup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workgroup_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkgroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := redshiftserverless.CreateWorkgroupInput{
		NamespaceName: aws.String(d.Get("namespace_name").(string)),
		WorkgroupName: aws.String(d.Get("workgroup_name").(string)),
	}

	if v, ok := d.GetOk("base_capacity"); ok {
		input.BaseCapacity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	input.Tags = Tags(tags.IgnoreAWS())

	out, err := conn.CreateWorkgroup(&input)

	if err != nil {
		return fmt.Errorf("error creating Redshift Serverless Workgroup : %w", err)
	}

	d.SetId(aws.StringValue(out.Workgroup.WorkgroupName))

	if _, err := waitWorkgroupAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Redshift Serverless Workgroup (%s) to be created: %w", d.Id(), err)
	}

	return resourceWorkgroupRead(d, meta)
}

func resourceWorkgroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	out, err := FindWorkgroupByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Workgroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Serverless Workgroup (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(out.WorkgroupArn)
	d.Set("arn", arn)
	d.Set("namespace_name", out.NamespaceName)
	d.Set("workgroup_name", out.WorkgroupName)
	d.Set("workgroup_id", out.WorkgroupId)
	d.Set("base_capacity", out.BaseCapacity)
	d.Set("security_group_ids", flex.FlattenStringSet(out.SecurityGroupIds))
	d.Set("subnet_ids", flex.FlattenStringSet(out.SubnetIds))

	tags, err := ListTags(conn, arn)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, "UnknownOperationException") {
			return nil
		}

		return fmt.Errorf("error listing tags for edshift Serverless Workgroup (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceWorkgroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &redshiftserverless.UpdateWorkgroupInput{
			WorkgroupName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("base_capacity"); ok {
			input.BaseCapacity = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("subnet_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		_, err := conn.UpdateWorkgroup(input)
		if err != nil {
			return fmt.Errorf("error updating Redshift Serverless Workgroup (%s): %w", d.Id(), err)
		}

		if _, err := waitWorkgroupAvailable(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for Redshift Serverless Workgroup (%s) to be updated: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Serverless Workgroup (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceWorkgroupRead(d, meta)
}

func resourceWorkgroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	deleteInput := redshiftserverless.DeleteWorkgroupInput{
		WorkgroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Redshift Serverless Workgroup: %s", d.Id())
	_, err := conn.DeleteWorkgroup(&deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
			return nil
		}
		return err
	}

	if _, err := waitWorkgroupDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Redshift Serverless Workgroup (%s) delete: %w", d.Id(), err)
	}

	return nil
}
