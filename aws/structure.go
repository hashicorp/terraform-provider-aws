package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/beevik/etree"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/mitchellh/copystructure"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"gopkg.in/yaml.v2"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)






































































































































func pointersMapToStringList(pointers map[string]*string) map[string]interface{} {
	list := make(map[string]interface{}, len(pointers))
	for i, v := range pointers {
		list[i] = *v
	}
	return list
}

// diffStringMaps returns the set of keys and values that must be created, the set of keys
// and values that must be destroyed, and the set of keys and values that are unchanged.
func diffStringMaps(oldMap, newMap map[string]interface{}) (map[string]*string, map[string]*string, map[string]*string) {
	// First, we're creating everything we have
	add := map[string]*string{}
	for k, v := range newMap {
		add[k] = aws.String(v.(string))
	}

	// Build the maps of what to remove and what is unchanged
	remove := map[string]*string{}
	unchanged := map[string]*string{}
	for k, v := range oldMap {
		old, ok := add[k]
		if !ok || aws.StringValue(old) != v.(string) {
			// Delete it!
			remove[k] = aws.String(v.(string))
		} else if ok {
			unchanged[k] = aws.String(v.(string))
			// already present so remove from new
			delete(add, k)
		}
	}

	return add, remove, unchanged
}















































// TODO: refactor some of these helper functions and types in the terraform/helper packages

























// Takes a value containing YAML string and passes it through
// the YAML parser. Returns either a parsing
// error or original YAML string.
func checkYamlString(yamlString interface{}) (string, error) {
	var y interface{}

	if yamlString == nil || yamlString.(string) == "" {
		return "", nil
	}

	s := yamlString.(string)

	err := yaml.Unmarshal([]byte(s), &y)

	return s, err
}

func normalizeJsonOrYamlString(templateString interface{}) (string, error) {
	if looksLikeJsonString(templateString) {
		return structure.NormalizeJsonString(templateString.(string))
	}

	return checkYamlString(templateString)
}















func sliceContainsMap(l []interface{}, m map[string]interface{}) (int, bool) {
	for i, t := range l {
		if reflect.DeepEqual(m, t.(map[string]interface{})) {
			return i, true
		}
	}

	return -1, false
}


















































// checkIfNonKeyAttributesChanged returns true if non_key_attributes between old map and new map are different
func checkIfNonKeyAttributesChanged(oldMap, newMap map[string]interface{}) bool {
	oldNonKeyAttributes, oldNkaExists := oldMap["non_key_attributes"].(*schema.Set)
	newNonKeyAttributes, newNkaExists := newMap["non_key_attributes"].(*schema.Set)

	if oldNkaExists && newNkaExists {
		return !oldNonKeyAttributes.Equal(newNonKeyAttributes)
	}

	return oldNkaExists != newNkaExists
}



// Expanders + flatteners
































































































