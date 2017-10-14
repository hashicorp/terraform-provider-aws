package aws

func resourceAwsAthenaNamedQuery() *schema.Resource {
 	return &schema.Resource{
 		Create: resourceAwsAthenaNamedQueryCreate,
 		Read:   resourceAwsAthenaNamedQueryRead,
 		Delete: resourceAwsAthenaNamedQueryDelete,

 		Schema: map[string]*schema.Schema{
      "name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
      "query": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
      "database": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
      "description": &schema.Schema{
				Type:     schema.TypeString,
        Optional:     true,
			},
    }
  }
}

func resourceAwsAthenaNamedQueryCreate(d *schema.ResourceData, meta interface{}) error {
  return nil
}

func resourceAwsAthenaNamedQueryRead(d *schema.ResourceData, meta interface{}) error {
  return nil
}

func resourceAwsAthenaNamedQueryDelete(d *schema.ResourceData, meta interface{}) error {
  return nil
}
