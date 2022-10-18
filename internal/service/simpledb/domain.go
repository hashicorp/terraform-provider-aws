package simpledb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		DeleteWithoutTimeout: resourceDomainDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SimpleDBConn

	name := d.Get("name").(string)
	input := &simpledb.CreateDomainInput{
		DomainName: aws.String(name),
	}

	_, err := conn.CreateDomainWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating SimpleDB Domain (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceDomainRead(ctx, d, meta)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SimpleDBConn

	_, err := FindDomainByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SimpleDB Domain %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading SimpleDB Domain (%s): %s", d.Id(), err)
	}

	d.Set("name", d.Id())

	return nil
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SimpleDBConn

	log.Printf("[DEBUG] Deleting SimpleDB Domain: %s", d.Id())
	_, err := conn.DeleteDomainWithContext(ctx, &simpledb.DeleteDomainInput{
		DomainName: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting SimpleDB Domain (%s): %s", d.Id(), err)
	}

	return nil
}

func FindDomainByName(ctx context.Context, conn *simpledb.SimpleDB, name string) (*simpledb.DomainMetadataOutput, error) {
	input := &simpledb.DomainMetadataInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DomainMetadataWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, simpledb.ErrCodeNoSuchDomain) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
