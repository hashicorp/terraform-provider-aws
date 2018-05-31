package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
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
			"s3_target": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						//"exclusions": {
						//	Type:     schema.TypeSet,
						//	Optional: true,
						//	Elem:     schema.TypeString,
						//},
					},
				},
			},
			//"jdbc_targets": {
			//	Type:     schema.TypeList,
			//	Optional: true,
			//	MinItems: 1,
			//	Elem: &schema.Resource{
			//		Schema: map[string]*schema.Schema{
			//			"connection_name": {
			//				Type:     schema.TypeString,
			//				Required: true,
			//			},
			//			"path": {
			//				Type:     schema.TypeString,
			//				Required: true,
			//			},
			//			"exclusions": {
			//				Type:     schema.TypeSet,
			//				Optional: true,
			//				Elem:     schema.TypeString,
			//			},
			//		},
			//	},
			//},
			//"configuration": {
			//	Type:             schema.TypeString,
			//	Optional:         true,
			//	DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			//	ValidateFunc:     validateJsonString,
			//},
		},
	}
}

func resourceAwsGlueCatalogCrawlerCreate(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := glueConn.CreateCrawler(createCrawlerInput(name, d))
		if err != nil {
			if isAWSErr(err, "InvalidInputException", "Service is unable to assume role") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error creating Glue crawler: %s", err)
	}
	d.SetId(fmt.Sprintf("%s", name))

	return resourceAwsGlueCatalogCrawlerUpdate(d, meta)
}

func createCrawlerInput(crawlerName string, d *schema.ResourceData) *glue.CreateCrawlerInput {
	crawlerInput := &glue.CreateCrawlerInput{
		Name:         aws.String(crawlerName),
		DatabaseName: aws.String(d.Get("database_name").(string)),
		Role:         aws.String(d.Get("role").(string)),
		Targets:      createCrawlerTargets(d),
	}
	//if description, ok := d.GetOk("description"); ok {
	//	crawlerInput.Description = aws.String(description.(string))
	//}
	//if schedule, ok := d.GetOk("schedule"); ok {
	//	crawlerInput.Description = aws.String(schedule.(string))
	//}
	//if classifiers, ok := d.GetOk("classifiers"); ok {
	//	crawlerInput.Classifiers = expandStringList(classifiers.(*schema.Set).List())
	//}
	//if v, ok := d.GetOk("schema_change_policy"); ok {
	//	crawlerInput.SchemaChangePolicy = createSchemaPolicy(v)
	//}
	//if tablePrefix, ok := d.GetOk("table_prefix"); ok {
	//	crawlerInput.TablePrefix = aws.String(tablePrefix.(string))
	//}
	//if targets, ok := d.GetOk("targets"); ok {
	//	crawlerInput.Targets = createCrawlerTargets(targets)
	//}
	//if configuration, ok := d.GetOk("configuration"); ok {
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

func createCrawlerTargets(d *schema.ResourceData) *glue.CrawlerTargets {
	//attributes := d.(map[string]interface{})
	crawlerTargets := &glue.CrawlerTargets{}

	//if jdbcTargetsResource, ok := attributes["jdbc_targets"]; ok {
	//	jdbcTargets := jdbcTargetsResource.(*schema.Set).List()
	//	var configsOut []*glue.JdbcTarget
	//
	//	for _, jdbcTarget := range jdbcTargets {
	//		attributes := jdbcTarget.(map[string]interface{})
	//
	//		target := &glue.JdbcTarget{
	//			ConnectionName: aws.String(attributes["connection_name"].(string)),
	//			Path:           aws.String(attributes["path"].(string)),
	//		}
	//
	//		if exclusions, ok := attributes["exclusions"]; ok {
	//			target.Exclusions = expandStringList(exclusions.(*schema.Set).List())
	//		}
	//
	//		configsOut = append(configsOut, target)
	//	}
	//
	//	crawlerTargets.JdbcTargets = configsOut
	//}

	if v, ok := d.GetOk("s3_target"); ok {
		crawlerTargets.S3Targets = expandS3Targets(v.([]interface{}))
	}

	//if s3Targets, ok := attributes["s3_target"]; ok {
	//	targets := s3Targets.(*schema.Set).List()
	//	var configsOut []*glue.S3Target
	//
	//	for _, target := range targets {
	//		attributes := target.(map[string]interface{})
	//
	//		target := &glue.S3Target{
	//			Path: aws.String(attributes["path"].(string)),
	//		}
	//
	//		if exclusions, ok := attributes["exclusions"]; ok {
	//			target.Exclusions = expandStringList(exclusions.(*schema.Set).List())
	//		}
	//
	//		configsOut = append(configsOut, target)
	//	}
	//	crawlerTargets.S3Targets = configsOut
	//}

	return crawlerTargets
}

func expandS3Targets(cfgs []interface{}) []*glue.S3Target {
	if len(cfgs) < 1 {
		return []*glue.S3Target{}
	}

	perms := make([]*glue.S3Target, len(cfgs), len(cfgs))
	for i, rawCfg := range cfgs {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandS3Target(cfg)
	}
	return perms
}

func expandS3Target(cfg map[string]interface{}) *glue.S3Target {
	return &glue.S3Target{
		Path: aws.String(cfg["path"].(string)),
		//FromPort: aws.Int64(int64(cfg["from_port"].(int))),
		//IpRange:  aws.String(cfg["ip_range"].(string)),
		//Protocol: aws.String(cfg["protocol"].(string)),
		//ToPort:   aws.Int64(int64(cfg["to_port"].(int))),
	}
}

func resourceAwsGlueCatalogCrawlerUpdate(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	crawlerInput := glue.UpdateCrawlerInput(*createCrawlerInput(name, d))

	if _, err := glueConn.UpdateCrawler(&crawlerInput); err != nil {
		return err
	}

	return resourceAwsGlueCatalogCrawlerRead(d, meta)
}

func resourceAwsGlueCatalogCrawlerRead(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	input := &glue.GetCrawlerInput{
		Name: aws.String(name),
	}

	crawlerOutput, err := glueConn.GetCrawler(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Crawler (%s) not found, removing from state", d.Id())
			d.SetId("")
		}

		return fmt.Errorf("error reading Glue crawler: %s", err.Error())
	}

	d.Set("name", crawlerOutput.Crawler.Name)
	d.Set("database_name", crawlerOutput.Crawler.DatabaseName)
	d.Set("role", crawlerOutput.Crawler.Role)
	//d.Set("description", crawler.Description)
	//d.Set("schedule", crawler.Schedule)

	//var classifiers []string
	//if len(crawler.Classifiers) > 0 {
	//	for _, value := range crawler.Classifiers {
	//		classifiers = append(classifiers, *value)
	//	}
	//}
	//d.Set("classifiers", crawler.Classifiers)

	//if crawlerOutput.Crawler.SchemaChangePolicy != nil {
	//	schemaPolicy := map[string]string{
	//		"delete_behavior": *crawlerOutput.Crawler.SchemaChangePolicy.DeleteBehavior,
	//		"update_behavior": *crawlerOutput.Crawler.SchemaChangePolicy.UpdateBehavior,
	//	}
	//	d.Set("schema_change_policy", schemaPolicy)
	//}
	//
	//d.Set("table_prefix", crawlerOutput.Crawler.TablePrefix)

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
