package aws

import (
	//"log"
	//"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
)

// diffTags takes our tags locally and the ones remotely and returns
// the set of tags that must be created, and the set of tags that must
// be destroyed.
func diffTagsQuickSight(oldTags, newTags []*quicksight.Tag) ([]*quicksight.Tag, []*quicksight.Tag) {
	// First, we're creating everything we have
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	// Build the list of what to remove
	var remove []*quicksight.Tag
	for _, t := range oldTags {
		old, ok := create[aws.StringValue(t.Key)]
		if !ok || old != aws.StringValue(t.Value) {
			// Delete it!
			remove = append(remove, t)
		} else if ok {
			delete(create, aws.StringValue(t.Key))
		}
	}

	return tagsFromMapQuickSight(create), remove
}

func tagsFromMapQuickSight(m map[string]interface{}) []*quicksight.Tag {
	result := make([]*quicksight.Tag, 0, len(m))
	for k, v := range m {
		t := &quicksight.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		result = append(result, t)
	}

	return result
}

func tagKeysQuickSight(ts []*quicksight.Tag) []*string {
	result := make([]*string, 0, len(ts))
	for _, t := range ts {
		result = append(result, t.Key)
	}
	return result
}
