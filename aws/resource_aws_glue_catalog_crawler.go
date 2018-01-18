package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsGlueCatalogCrawler() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueCatalogCrawlerCreate,
		Read:   resourceAwsGlueCatalogCrawlerRead,
		Update: resourceAwsGlueCatalogCrawlerUpdate,
		Delete: resourceAwsGlueCatalogCrawlerDelete,
		Exists: resourceAwsGlueCatalogCrawlerExists,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			//"description": {
			//	Type:     schema.TypeString,
			//	Optional: true,
			//},
			//"schedule": {
			//	Type: schema.TypeString,
			//	//TODO: Write a validate function on cron
			//	//ValidateFunc: validateCron,
			//},
			//"classifiers": {
			//	Type:     schema.TypeList,
			//	Optional: true,
			//},
			//"schema_change_policy": {
			//	Type:     schema.TypeSet,
			//	Optional: true,
			//	Elem: &schema.Resource{
			//		Schema: map[string]*schema.Schema{
			//			"delete_behavior": {
			//				Type:     schema.TypeString,
			//				Optional: true,
			//				//ValidateFunc: validateDeletion,
			//				//TODO: Write a validate function to ensure value matches enum
			//			},
			//			"update_behavior": {
			//				Type:     schema.TypeString,
			//				Optional: true,
			//				//ValidateFunc: validateUpdate,
			//				//TODO: Write a validate function to ensure value matches enum
			//			},
			//		},
			//	},
			//},
			//"table_prefix": {
			//	Type:     schema.TypeString,
			//	Optional: true,
			//},
			"targets": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"jdbc_targets": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection_name": {
										Type: schema.TypeString,
									},
									"path": {
										Type: schema.TypeString,
									},
									"exclusions": {
										Type: schema.TypeList,
									},
								},
							},
						},
						"s3_targets": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"path": {
										Type: schema.TypeList,
									},
									"exclusions": {
										Type: schema.TypeList,
									},
								},
							},
						},
					},
				},
			},
			//"configuration": {
			//	Type:             schema.TypeString,
			//	Optional:         true,
			//	DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			//	ValidateFunc:     validateJsonString,
			//},
		},
	}
}

func resourceAwsGlueCatalogCrawlerCreate(resource *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	name := resource.Get("name").(string)

	_, err := glueConn.CreateCrawler(createCrawlerInput(name, resource))

	if err != nil {
		return fmt.Errorf("error creating Glue crawler: %s", err)
	}
	resource.SetId(fmt.Sprintf("%s", name))

	return resourceAwsGlueCatalogCrawlerUpdate(resource, meta)
}

func createCrawlerInput(crawlerName string, resource *schema.ResourceData) *glue.CreateCrawlerInput {
	crawlerInput := &glue.CreateCrawlerInput{
		Name:         aws.String(crawlerName),
		DatabaseName: aws.String(resource.Get("database_name").(string)),
		Role:         aws.String(resource.Get("role").(string)),
		Targets:      createCrawlerTargets(resource.Get("targets")),
	}
	//if description, ok := resource.GetOk("description"); ok {
	//	crawlerInput.Description = aws.String(description.(string))
	//}
	//if schedule, ok := resource.GetOk("schedule"); ok {
	//	crawlerInput.Description = aws.String(schedule.(string))
	//}
	//if classifiers, ok := resource.GetOk("classifiers"); ok {
	//	crawlerInput.Classifiers = expandStringList(classifiers.(*schema.Set).List())
	//}
	//if v, ok := resource.GetOk("schema_change_policy"); ok {
	//	crawlerInput.SchemaChangePolicy = createSchemaPolicy(v)
	//}
	//if tablePrefix, ok := resource.GetOk("table_prefix"); ok {
	//	crawlerInput.TablePrefix = aws.String(tablePrefix.(string))
	//}
	//if targets, ok := resource.GetOk("targets"); ok {
	//	crawlerInput.Targets = createCrawlerTargets(targets)
	//}
	//if configuration, ok := resource.GetOk("configuration"); ok {
	//	crawlerInput.Configuration = aws.String(configuration.(string))
	//}
	return crawlerInput
}

func createSchemaPolicy(v interface{}) *glue.SchemaChangePolicy {
	schemaAttributes := v.(map[string]interface{})
	schemaPolicy := &glue.SchemaChangePolicy{}

	if updateBehavior, ok := schemaAttributes["update_behavior"]; ok {
		schemaPolicy.UpdateBehavior = aws.String(updateBehavior.(string))
	}

	if deleteBehavior, ok := schemaAttributes["delete_behavior"]; ok {
		schemaPolicy.DeleteBehavior = aws.String(deleteBehavior.(string))
	}
	return schemaPolicy
}

func createCrawlerTargets(v interface{}) *glue.CrawlerTargets {
	attributes := v.(map[string]interface{})
	crawlerTargets := &glue.CrawlerTargets{}

	if jdbcTargetsResource, ok := attributes["jdbc_targets"]; ok {
		jdbcTargets := jdbcTargetsResource.(*schema.Set).List()
		var configsOut []*glue.JdbcTarget

		for _, jdbcTarget := range jdbcTargets {
			attributes := jdbcTarget.(map[string]interface{})

			target := &glue.JdbcTarget{
				ConnectionName: aws.String(attributes["connection_name"].(string)),
				Path:           aws.String(attributes["path"].(string)),
			}

			if exclusions, ok := attributes["exclusions"]; ok {
				target.Exclusions = expandStringList(exclusions.(*schema.Set).List())
			}

			configsOut = append(configsOut, target)
		}

		crawlerTargets.JdbcTargets = configsOut
	}

	if s3Targets, ok := attributes["s3_targets"]; ok {
		targets := s3Targets.(*schema.Set).List()
		var configsOut []*glue.S3Target

		for _, target := range targets {
			attributes := target.(map[string]interface{})

			target := &glue.S3Target{
				Path: aws.String(attributes["path"].(string)),
			}

			if exclusions, ok := attributes["exclusions"]; ok {
				target.Exclusions = expandStringList(exclusions.(*schema.Set).List())
			}

			configsOut = append(configsOut, target)
		}
		crawlerTargets.S3Targets = configsOut
	}

	return crawlerTargets
}

func resourceAwsGlueCatalogCrawlerUpdate(resource *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	name := resource.Get("name").(string)

	crawlerInput := glue.UpdateCrawlerInput(*createCrawlerInput(name, resource))

	if _, err := glueConn.UpdateCrawler(&crawlerInput); err != nil {
		return err
	}

	return resourceAwsGlueCatalogCrawlerRead(resource, meta)
}

func resourceAwsGlueCatalogCrawlerRead(resource *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	name := resource.Get("name").(string)

	input := &glue.GetCrawlerInput{
		Name: aws.String(name),
	}

	crawlerOutput, err := glueConn.GetCrawler(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Crawler (%s) not found, removing from state", resource.Id())
			resource.SetId("")
		}

		return fmt.Errorf("error reading Glue crawler: %s", err.Error())
	}

	resource.Set("name", crawlerOutput.Crawler.Name)
	resource.Set("database_name", crawlerOutput.Crawler.DatabaseName)
	resource.Set("role", crawlerOutput.Crawler.Role)
	//resource.Set("description", crawler.Description)
	//resource.Set("schedule", crawler.Schedule)

	//var classifiers []string
	//if len(crawler.Classifiers) > 0 {
	//	for _, value := range crawler.Classifiers {
	//		classifiers = append(classifiers, *value)
	//	}
	//}
	//resource.Set("classifiers", crawler.Classifiers)

	//if crawlerOutput.Crawler.SchemaChangePolicy != nil {
	//	schemaPolicy := map[string]string{
	//		"delete_behavior": *crawlerOutput.Crawler.SchemaChangePolicy.DeleteBehavior,
	//		"update_behavior": *crawlerOutput.Crawler.SchemaChangePolicy.UpdateBehavior,
	//	}
	//	resource.Set("schema_change_policy", schemaPolicy)
	//}
	//
	//resource.Set("table_prefix", crawlerOutput.Crawler.TablePrefix)

	return nil
}

func resourceAwsGlueCatalogCrawlerDelete(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	log.Printf("[DEBUG] deleting Glue crawler: %s", name)
	_, err := glueConn.DeleteCrawler(&glue.DeleteCrawlerInput{
		Name: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("error deleting Glue crawler: %s", err.Error())
	}
	return nil
}

func resourceAwsGlueCatalogCrawlerExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	glueConn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	input := &glue.GetCrawlerInput{
		Name: aws.String(name),
	}

	_, err := glueConn.GetCrawler(input)
	return err == nil, err
}
