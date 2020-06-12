package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
)

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapServiceCatalog(m map[string]interface{}) []*servicecatalog.Tag {
	result := make([]*servicecatalog.Tag, 0, len(m))
	for k, v := range m {
		t := &servicecatalog.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		result = append(result, t)
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapServiceCatalog(ts []*servicecatalog.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	return result
}
