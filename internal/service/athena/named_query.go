package athena

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceNamedQuery() *schema.Resource {
	return &schema.Resource{
		Create: resourceNamedQueryCreate,
		Read:   resourceNamedQueryRead,
		Delete: resourceNamedQueryDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"query": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"workgroup": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "primary",
			},
			"database": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNamedQueryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn

	input := &athena.CreateNamedQueryInput{
		Database:    aws.String(d.Get("database").(string)),
		Name:        aws.String(d.Get("name").(string)),
		QueryString: aws.String(d.Get("query").(string)),
	}
	if raw, ok := d.GetOk("workgroup"); ok {
		input.WorkGroup = aws.String(raw.(string))
	}
	if raw, ok := d.GetOk("description"); ok {
		input.Description = aws.String(raw.(string))
	}

	resp, err := conn.CreateNamedQuery(input)
	if err != nil {
		return err
	}
	d.SetId(aws.StringValue(resp.NamedQueryId))
	return resourceNamedQueryRead(d, meta)
}

func resourceNamedQueryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn

	input := &athena.GetNamedQueryInput{
		NamedQueryId: aws.String(d.Id()),
	}

	resp, err := conn.GetNamedQuery(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, athena.ErrCodeInvalidRequestException, d.Id()) {
			log.Printf("[WARN] Athena Named Query (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", resp.NamedQuery.Name)
	d.Set("query", resp.NamedQuery.QueryString)
	d.Set("workgroup", resp.NamedQuery.WorkGroup)
	d.Set("database", resp.NamedQuery.Database)
	d.Set("description", resp.NamedQuery.Description)
	return nil
}

func resourceNamedQueryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AthenaConn

	input := &athena.DeleteNamedQueryInput{
		NamedQueryId: aws.String(d.Id()),
	}

	_, err := conn.DeleteNamedQuery(input)
	return err
}
