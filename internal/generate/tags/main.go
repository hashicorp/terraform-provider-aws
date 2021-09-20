//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

const filename = `tags_gen.go`

var (
	getTag           = flag.String("GetTag", "", "whether to generate GetTag")
	listTags         = flag.String("ListTags", "", "whether to generate ListTags")
	serviceTagsMap   = flag.String("ServiceTagsMap", "", "whether to generate service tags for map")
	serviceTagsSlice = flag.String("ServiceTagsSlice", "", "whether to generate service tags for slice")
	updateTags       = flag.String("UpdateTags", "", "whether to generate UpdateTags")

	listTagsInFiltIDName  = flag.String("ListTagsInFiltIDName", "", "listTagsInFiltIDName")
	listTagsInIDElem      = flag.String("ListTagsInIDElem", "ResourceArn", "listTagsInIDElem")
	listTagsInIDNeedSlice = flag.String("ListTagsInIDNeedSlice", "", "listTagsInIDNeedSlice")
	listTagsOp            = flag.String("ListTagsOp", "ListTagsForResource", "listTagsOp")
	listTagsOutTagsElem   = flag.String("ListTagsOutTagsElem", "Tags", "listTagsOutTagsElem")
	tagInCustomVal        = flag.String("TagInCustomVal", "", "tagInCustomVal")
	tagInIDElem           = flag.String("TagInIDElem", "ResourceArn", "tagInIDElem")
	tagInIDNeedSlice      = flag.String("TagInIDNeedSlice", "", "tagInIDNeedSlice")
	tagInTagsElem         = flag.String("TagInTagsElem", "Tags", "tagInTagsElem")
	tagKeyType            = flag.String("TagKeyType", "", "tagKeyType")
	tagOp                 = flag.String("TagOp", "TagResource", "tagOp")
	tagOpBatchSize        = flag.String("TagOpBatchSize", "", "tagOpBatchSize")
	tagResTypeElem        = flag.String("TagResTypeElem", "", "tagResTypeElem")
	tagType               = flag.String("TagType", "Tag", "tagType")
	tagType2              = flag.String("TagType2", "", "tagType")
	TagTypeAddBoolElem    = flag.String("TagTypeAddBoolElem", "", "TagTypeAddBoolElem")
	tagTypeIDElem         = flag.String("TagTypeIDElem", "", "tagTypeIDElem")
	tagTypeKeyElem        = flag.String("TagTypeKeyElem", "Key", "tagTypeKeyElem")
	tagTypeValElem        = flag.String("TagTypeValElem", "Value", "tagTypeValElem")
	untagInCustomVal      = flag.String("UntagInCustomVal", "", "untagInCustomVal")
	untagInNeedTagKeyType = flag.String("UntagInNeedTagKeyType", "", "untagInNeedTagKeyType")
	untagInNeedTagType    = flag.String("UntagInNeedTagType", "", "untagInNeedTagType")
	untagInTagsElem       = flag.String("UntagInTagsElem", "TagKeys", "untagInTagsElem")
	untagOp               = flag.String("UntagOp", "UntagResource", "untagOp")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

type TemplateData struct {
	AWSService     string
	ClientType     string
	ServicePackage string

	ListTagsInFiltIDName    string
	ListTagsInIDElem        string
	ListTagsInIDNeedSlice   string
	ListTagsOp              string
	ListTagsOutTagsElem     string
	ParentNotFoundError     string
	RetryCreateOnNotFound   string
	TagInCustomVal          string
	TagInIDElem             string
	TagInIDNeedSlice        string
	TagInTagsElem           string
	TagKeyType              string
	TagOp                   string
	TagOpBatchSize          string
	TagPackage              string
	TagResTypeElem          string
	TagType                 string
	TagType2                string
	TagTypeAddBoolElem      string
	TagTypeAddBoolElemSnake string
	TagTypeIDElem           string
	TagTypeKeyElem          string
	TagTypeValElem          string
	UntagInCustomVal        string
	UntagInNeedTagKeyType   string
	UntagInNeedTagType      string
	UntagInTagsElem         string
	UntagOp                 string
}

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	servicePackage := os.Getenv("GOPACKAGE")
	awsService, err := awsServiceName(servicePackage)

	if err != nil {
		log.Fatalf("encountered: %s", err)
	}

	awsServiceUpper, err := awsServiceNameUpper(servicePackage)

	if err != nil {
		log.Fatalf("encountered: %s", err)
	}

	clientType := fmt.Sprintf("*%s.%s", awsService, awsServiceUpper)

	tagPackage := awsService

	if tagPackage == "wafregional" {
		tagPackage = "waf"
	}

	templateData := TemplateData{
		AWSService:     awsService,
		ClientType:     clientType,
		ServicePackage: servicePackage,

		ListTagsInFiltIDName:    *listTagsInFiltIDName,
		ListTagsInIDElem:        *listTagsInIDElem,
		ListTagsInIDNeedSlice:   *listTagsInIDNeedSlice,
		ListTagsOp:              *listTagsOp,
		ListTagsOutTagsElem:     *listTagsOutTagsElem,
		TagInCustomVal:          *tagInCustomVal,
		TagInIDElem:             *tagInIDElem,
		TagInIDNeedSlice:        *tagInIDNeedSlice,
		TagInTagsElem:           *tagInTagsElem,
		TagKeyType:              *tagKeyType,
		TagOp:                   *tagOp,
		TagOpBatchSize:          *tagOpBatchSize,
		TagPackage:              tagPackage,
		TagResTypeElem:          *tagResTypeElem,
		TagType:                 *tagType,
		TagType2:                *tagType2,
		TagTypeAddBoolElem:      *TagTypeAddBoolElem,
		TagTypeAddBoolElemSnake: ToSnakeCase(*TagTypeAddBoolElem),
		TagTypeIDElem:           *tagTypeIDElem,
		TagTypeKeyElem:          *tagTypeKeyElem,
		TagTypeValElem:          *tagTypeValElem,
		UntagInCustomVal:        *untagInCustomVal,
		UntagInNeedTagKeyType:   *untagInNeedTagKeyType,
		UntagInNeedTagType:      *untagInNeedTagType,
		UntagInTagsElem:         *untagInTagsElem,
		UntagOp:                 *untagOp,
	}

	if *getTag != "" || *listTags != "" || *serviceTagsMap != "" || *serviceTagsSlice != "" || *updateTags != "" {
		writeTemplate(headerBody, "header", templateData)
	}

	if *getTag != "" {
		writeTemplate(gettagBody, "gettag", templateData)
	}

	if *listTags != "" {
		writeTemplate(listtagsBody, "listtags", templateData)
	}

	if *serviceTagsMap != "" {
		writeTemplate(servicetagsmapBody, "servicetagsmap", templateData)
	}

	if *serviceTagsSlice != "" {
		writeTemplate(servicetagssliceBody, "servicetagsslice", templateData)
	}

	if *updateTags != "" {
		writeTemplate(updatetagsBody, "updatetags", templateData)
	}
}

func writeTemplate(body string, templateName string, td TemplateData) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file (%s): %s", filename, err)
	}

	tplate, err := template.New(templateName).Parse(body)
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}

	var buffer bytes.Buffer
	err = tplate.Execute(&buffer, td)
	if err != nil {
		log.Fatalf("error executing template: %s", err)
	}

	contents, err := format.Source(buffer.Bytes())
	if err != nil {
		log.Fatalf("error formatting generated file: %s", err)
	}

	if _, err := f.Write(contents); err != nil {
		f.Close() // ignore error; Write error takes precedence
		log.Fatalf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		log.Fatalf("error closing file (%s): %s", filename, err)
	}
}

var headerBody = `
// Code generated by internal/generate/tags/main.go; DO NOT EDIT.

package {{ .ServicePackage }}

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/{{ .AWSService }}"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
    "github.com/hashicorp/aws-sdk-go-base/tfawserr"
    "github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

`

var gettagBody = `
// GetTag fetches an individual {{ .ServicePackage }} service tag for a resource.
// Returns whether the key value and any errors. A NotFoundError is used to signal that no value was found.
// This function will optimise the handling over ListTags, if possible.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
{{- if or ( .TagInIDElem ) ( .TagTypeAddBoolElem ) }}
func GetTag(conn {{ .ClientType }}, identifier string{{ if .TagResTypeElem }}, resourceType string{{ end }}, key string) (*tftags.TagData, error) {
{{- else }}
func GetTag(conn {{ .ClientType }}, identifier string{{ if .TagResTypeElem }}, resourceType string{{ end }}, key string) (*string, error) {
{{- end }}
	{{- if .ListTagsInFiltIDName }}
	input := &{{ .AWSService  }}.{{ .ListTagsOp }}Input{
		Filters: []*{{ .AWSService  }}.Filter{
			{
				Name:   aws.String("{{ .ListTagsInFiltIDName }}"),
				Values: []*string{aws.String(identifier)},
			},
			{
				Name:   aws.String("key"),
				Values: []*string{aws.String(key)},
			},
		},
	}

	output, err := conn.{{ .ListTagsOp }}(input)

	if err != nil {
		return nil, err
	}

	listTags := KeyValueTags(output.{{ .ListTagsOutTagsElem }}{{ if .TagTypeIDElem }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}{{ end }})
	{{- else }}
	listTags, err := ListTags(conn, identifier{{ if .TagResTypeElem }}, resourceType{{ end }})

	if err != nil {
		return nil, err
	}
	{{- end }}

	if !listTags.KeyExists(key) {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	{{ if or ( .TagInIDElem ) ( .TagTypeAddBoolElem) }}
	return listTags.KeyTagData(key), nil
	{{- else }}
	return listTags.KeyValue(key), nil
	{{- end }}
}

`

var listtagsBody = `
// ListTags lists {{ .ServicePackage }} service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(conn {{ .ClientType }}, identifier string{{ if .TagResTypeElem }}, resourceType string{{ end }}) (tftags.KeyValueTags, error) {
	input := &{{ .TagPackage  }}.{{ .ListTagsOp }}Input{
		{{- if .ListTagsInFiltIDName }}
		Filters: []*{{ .AWSService  }}.Filter{
			{
				Name:   aws.String("{{ .ListTagsInFiltIDName }}"),
				Values: []*string{aws.String(identifier)},
			},
		},
		{{- else }}
		{{- if .ListTagsInIDNeedSlice }}
		{{ .ListTagsInIDElem }}: aws.StringSlice([]string{identifier}),
		{{- else }}
		{{ .ListTagsInIDElem }}: aws.String(identifier),
		{{- end }}
		{{- if .TagResTypeElem }}
		{{ .TagResTypeElem }}:         aws.String(resourceType),
		{{- end }}
		{{- end }}
	}

	output, err := conn.{{ .ListTagsOp }}(input)

	{{ .ParentNotFoundError }}

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.{{ .ListTagsOutTagsElem }}{{ if .TagTypeIDElem }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}{{ end }}), nil
}

`

var servicetagsmapBody = `
// map[string]*string handling

// Tags returns {{ .ServicePackage }} service tags.
func Tags(tags tftags.KeyValueTags) map[string]*string {
	return aws.StringMap(tags.Map())
}

// KeyValueTags creates KeyValueTags from {{ .ServicePackage }} service tags.
func KeyValueTags(tags map[string]*string) tftags.KeyValueTags {
	return tftags.New(tags)
}
`

var servicetagssliceBody = `
// []*SERVICE.Tag handling

{{ if and ( .TagTypeIDElem ) ( .TagTypeAddBoolElem ) }}
// ListOfMap returns a list of {{ .ServicePackage }} in flattened map.
//
// Compatible with setting Terraform state for strongly typed configuration blocks.
//
// This function strips tag resource identifier and type. Generally, this is
// the desired behavior so the tag schema does not require those attributes.
// Use (tftags.KeyValueTags).ListOfMap() for full tag information.
func ListOfMap(tags tftags.KeyValueTags) []interface{} {
	var result []interface{}

	for _, key := range tags.Keys() {
		m := map[string]interface{}{
			"key":                   key,
			"value":                 aws.StringValue(tags.KeyValue(key)),
			{{ if .TagTypeAddBoolElem }}
			"{{ .TagTypeAddBoolElemSnake }}": aws.BoolValue(tags.KeyAdditionalBoolValue(key, "{{ .TagTypeAddBoolElem }}")),
			{{ end }}
		}

		result = append(result, m)
	}

	return result
}
{{- end }}

{{ if eq .ServicePackage "autoscaling" }}
// ListOfStringMap returns a list of {{ .ServicePackage }} tags in flattened map of only string values.
//
// Compatible with setting Terraform state for legacy []map[string]string schema.
// Deprecated: Will be removed in a future major version without replacement.
func ListOfStringMap(tags tftags.KeyValueTags) []interface{} {
	var result []interface{}

	for _, key := range tags.Keys() {
		m := map[string]string{
			"key":                   key,
			"value":                 aws.StringValue(tags.KeyValue(key)),
			{{ if .TagTypeAddBoolElem }}
			"{{ .TagTypeAddBoolElemSnake }}": strconv.FormatBool(aws.BoolValue(tags.KeyAdditionalBoolValue(key, "{{ .TagTypeAddBoolElem }}"))),
			{{ end }}
		}

		result = append(result, m)
	}

	return result
}
{{- end }}

{{- if .TagKeyType }}
// TagKeys returns {{ .ServicePackage }} service tag keys.
func TagKeys(tags tftags.KeyValueTags) []*{{ .AWSService }}.{{ .TagKeyType }} {
	result := make([]*{{ .AWSService }}.{{ .TagKeyType }}, 0, len(tags))

	for k := range tags.Map() {
		tagKey := &{{ .AWSService }}.{{ .TagKeyType }}{
			{{ .TagTypeKeyElem }}: aws.String(k),
		}

		result = append(result, tagKey)
	}

	return result
}
{{- end }}

// Tags returns {{ .ServicePackage }} service tags.
func Tags(tags tftags.KeyValueTags) []*{{ .TagPackage }}.{{ .TagType }} {
	{{- if or ( .TagTypeIDElem ) ( .TagTypeAddBoolElem) }}
	var result []*{{ .AWSService }}.{{ .TagType }}

	for _, key := range tags.Keys() {
		tag := &{{ .AWSService }}.{{ .TagType }}{
			{{ .TagTypeKeyElem }}:        aws.String(key),
			{{ .TagTypeValElem }}:      tags.KeyValue(key),
			{{- if ( .TagTypeIDElem ) }}
			{{ .TagTypeIDElem }}: tags.KeyAdditionalStringValue(key, "{{ .TagTypeIDElem }}"),
			{{- if ( .TagResTypeElem ) }}
			{{ .TagResTypeElem }}:   tags.KeyAdditionalStringValue(key, "{{ .TagResTypeElem }}"),
			{{- end }}
			{{- end }}
			{{- if .TagTypeAddBoolElem }}
			{{ .TagTypeAddBoolElem }}:                          tags.KeyAdditionalBoolValue(key, "{{ .TagTypeAddBoolElem }}"),
			{{- end }}
		}

		result = append(result, tag)
	}
	{{- else }}
	result := make([]*{{ .TagPackage }}.{{ .TagType }}, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &{{ .TagPackage }}.{{ .TagType }}{
			{{ .TagTypeKeyElem }}:   aws.String(k),
			{{ .TagTypeValElem }}: aws.String(v),
		}

		result = append(result, tag)
	}
	{{- end }}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from {{ .AWSService }} service tags.
{{- if or ( .TagType2 ) ( .TagTypeAddBoolElem ) }}
//
// Accepts the following types:
//   - []*{{ .AWSService }}.{{ .TagType }}
{{- if .TagType2 }}
//   - []*{{ .AWSService }}.{{ .TagType2 }}
{{- end }}
{{- if .TagTypeAddBoolElem }}
//   - []interface{} (Terraform TypeList configuration block compatible)
//   - *schema.Set (Terraform TypeSet configuration block compatible)
{{- end }}
func KeyValueTags(tags interface{}{{ if .TagTypeIDElem }}, identifier string{{ if .TagResTypeElem }}, resourceType string{{ end }}{{ end }}) tftags.KeyValueTags {
	switch tags := tags.(type) {
	case []*{{ .AWSService }}.{{ .TagType }}:
		{{- if or ( .TagTypeIDElem ) ( .TagTypeAddBoolElem) }}
		m := make(map[string]*tftags.TagData, len(tags))

		for _, tag := range tags {
			tagData := &tftags.TagData{
				Value: tag.{{ .TagTypeValElem }},
			}

			tagData.AdditionalBoolFields = make(map[string]*bool)
			{{- if .TagTypeAddBoolElem }}
			tagData.AdditionalBoolFields["{{ .TagTypeAddBoolElemSnake }}"] = tag.{{ .TagTypeAddBoolElem }}
			{{- end }}

			{{- if .TagTypeIDElem }}
			tagData.AdditionalStringFields = make(map[string]*string)
			tagData.AdditionalStringFields["{{ .TagTypeIDElem }}"] = &identifier
			{{- if .TagResTypeElem }}
			tagData.AdditionalStringFields["{{ .TagResTypeElem }}"] = &resourceType
			{{- end }}
			{{- end }}

			m[aws.StringValue(tag.{{ .TagTypeKeyElem }})] = tagData
		}
		{{- else }}
		m := make(map[string]*string, len(tags))

		for _, tag := range tags {
			m[aws.StringValue(tag.{{ .TagTypeKeyElem }})] = tag.{{ .TagTypeValElem }}
		}
		{{- end }}

		return tftags.New(m)
	case []*{{ .AWSService }}.{{ .TagType2 }}:
		{{- if or ( .TagTypeIDElem ) ( .TagTypeAddBoolElem) }}
		m := make(map[string]*tftags.TagData, len(tags))

		for _, tag := range tags {
			tagData := &tftags.TagData{
				Value: tag.{{ .TagTypeValElem }},
			}

			{{- if .TagTypeAddBoolElem }}
			tagData.AdditionalBoolFields = make(map[string]*bool)
			tagData.AdditionalBoolFields["{{ .TagTypeAddBoolElemSnake }}"] = tag.{{ .TagTypeAddBoolElem }}
			{{- end }}

			{{- if .TagTypeIDElem }}
			tagData.AdditionalStringFields = make(map[string]*string)
			tagData.AdditionalStringFields["{{ .TagTypeIDElem }}"] = &identifier
			{{- if .TagResTypeElem }}
			tagData.AdditionalStringFields["{{ .TagResTypeElem }}"] = &resourceType
			{{- end }}
			{{- end }}

			m[aws.StringValue(tag.{{ .TagTypeKeyElem }})] = tagData
		}
		{{- else }}
		m := make(map[string]*string, len(tags))

		for _, tag := range tags {
			m[aws.StringValue(tag.{{ .TagTypeKeyElem }})] = tag.{{ .TagTypeValElem }}
		}
		{{- end }}

		return tftags.New(m)
	{{- if .TagTypeAddBoolElem }}
	case *schema.Set:
		return KeyValueTags(tags.List(){{ if .TagTypeIDElem }}, identifier{{ if .TagResTypeElem }}, resourceType{{ end }}{{ end }})
	case []interface{}:
		result := make(map[string]*tftags.TagData)

		for _, tfMapRaw := range tags {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			key, ok := tfMap["key"].(string)

			if !ok {
				continue
			}

			tagData := &tftags.TagData{}

			if v, ok := tfMap["value"].(string); ok {
				tagData.Value = &v
			}

			{{ if .TagTypeAddBoolElem }}
			tagData.AdditionalBoolFields = make(map[string]*bool)
			{{- if .TagTypeAddBoolElem }}
			if v, ok := tfMap["{{ .TagTypeAddBoolElemSnake }}"].(bool); ok {
				tagData.AdditionalBoolFields["{{ .TagTypeAddBoolElem }}"] = &v
			}
			{{- end }}
			{{ if eq .ServicePackage "autoscaling" }}
			// Deprecated: Legacy map handling
			{{- if .TagTypeAddBoolElem }}
			if v, ok := tfMap["{{ .TagTypeAddBoolElemSnake }}"].(string); ok {
				b, _ := strconv.ParseBool(v)
				tagData.AdditionalBoolFields["{{ .TagTypeAddBoolElem }}"] = &b
			}
			{{- end }}
			{{- end }}
			{{- end }}

			{{ if .TagTypeIDElem }}
			tagData.AdditionalStringFields = make(map[string]*string)
			tagData.AdditionalStringFields["{{ .TagTypeIDElem }}"] = &identifier
			{{- if .TagResTypeElem }}
			tagData.AdditionalStringFields["{{ .TagResTypeElem }}"] = &resourceType
			{{- end }}
			{{- end }}

			result[key] = tagData
		}

		return tftags.New(result)
	{{- end }}
	default:
		return tftags.New(nil)
	}
}
{{- else }}
func KeyValueTags(tags []*{{ .TagPackage }}.{{ .TagType }}) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.{{ .TagTypeKeyElem }})] = tag.{{ .TagTypeValElem }}
	}

	return tftags.New(m)
}
{{- end }}

`

var updatetagsBody = `
// UpdateTags updates {{ .ServicePackage }} service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
{{- if  .TagTypeAddBoolElem }}
func UpdateTags(conn {{ .ClientType }}, identifier string{{ if .TagResTypeElem }}, resourceType string{{ end }}, oldTagsSet interface{}, newTagsSet interface{}) error {
	oldTags := KeyValueTags(oldTagsSet, identifier{{ if .TagResTypeElem }}, resourceType{{ end }})
	newTags := KeyValueTags(newTagsSet, identifier{{ if .TagResTypeElem }}, resourceType{{ end }})
{{- else }}
func UpdateTags(conn {{ .ClientType }}, identifier string{{ if .TagResTypeElem }}, resourceType string{{ end }}, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)
{{- end }}
	{{- if eq (.TagOp) (.UntagOp) }}
	removedTags := oldTags.Removed(newTags)
	updatedTags := oldTags.Updated(newTags)

	// Ensure we do not send empty requests
	if len(removedTags) == 0 && len(updatedTags) == 0 {
		return nil
	}

	input := &{{ .AWSService }}.{{ .TagOp }}Input{
		{{- if not ( .TagTypeIDElem ) }}
		{{- if .TagInIDNeedSlice }}
		{{ .TagInIDElem }}:   aws.StringSlice([]string{identifier}),
		{{- else }}
		{{ .TagInIDElem }}:   aws.String(identifier),
		{{- end }}
		{{- if .TagResTypeElem }}
		{{ .TagResTypeElem }}:      aws.String(resourceType),
		{{- end }}
		{{- end }}
	}

	if len(updatedTags) > 0 {
		input.{{ .TagInTagsElem }} = Tags(updatedTags.IgnoreAws())
	}

	if len(removedTags) > 0 {
		{{- if .UntagInNeedTagType }}
		input.{{ .UntagInTagsElem }} = Tags(removedTags.IgnoreAws())
		{{- else if .UntagInNeedTagKeyType }}
		input.{{ .UntagInTagsElem }} = TagKeys(removedTags.IgnoreAws())
		{{- else if .UntagInCustomVal }}
		input.{{ .UntagInTagsElem }} = {{ .UntagInCustomVal }}
		{{- else }}
		input.{{ .UntagInTagsElem }} = aws.StringSlice(removedTags.Keys())
		{{- end }}
	}

	_, err := conn.{{ .TagOp }}(input)

	if err != nil {
		return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
	}

	{{- else }}

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		{{- if .TagOpBatchSize }}
		for _, removedTags := range removedTags.Chunks({{ .TagOpBatchSize }}) {
		{{- end }}
		input := &{{ .TagPackage }}.{{ .UntagOp }}Input{
			{{- if not ( .TagTypeIDElem ) }}
			{{- if .TagInIDNeedSlice }}
			{{ .TagInIDElem }}:   aws.StringSlice([]string{identifier}),
			{{- else }}
			{{ .TagInIDElem }}:   aws.String(identifier),
			{{- end }}
			{{- if .TagResTypeElem }}
			{{ .TagResTypeElem }}: aws.String(resourceType),
			{{- end }}
			{{- end }}
			{{- if .UntagInNeedTagType }}
			{{ .UntagInTagsElem }}:       Tags(removedTags.IgnoreAws()),
			{{- else if .UntagInNeedTagKeyType }}
			{{ .UntagInTagsElem }}:       TagKeys(removedTags.IgnoreAws()),
			{{- else if .UntagInCustomVal }}
			{{ .UntagInTagsElem }}:       {{ .UntagInCustomVal }},
			{{- else }}
			{{ .UntagInTagsElem }}:       aws.StringSlice(removedTags.IgnoreAws().Keys()),
			{{- end }}
		}

		_, err := conn.{{ .UntagOp }}(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
		{{- if .TagOpBatchSize }}
		}
		{{- end }}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		{{- if .TagOpBatchSize }}
		for _, updatedTags := range updatedTags.Chunks({{ .TagOpBatchSize }}) {
		{{- end }}
		input := &{{ .TagPackage }}.{{ .TagOp }}Input{
			{{- if not ( .TagTypeIDElem ) }}
			{{- if .TagInIDNeedSlice }}
			{{ .TagInIDElem }}: aws.StringSlice([]string{identifier}),
			{{- else }}
			{{ .TagInIDElem }}: aws.String(identifier),
			{{- end }}
			{{- if .TagResTypeElem }}
			{{ .TagResTypeElem }}:    aws.String(resourceType),
			{{- end }}
			{{- end }}
			{{- if .TagInCustomVal }}
			{{ .TagInTagsElem }}:       {{ .TagInCustomVal }},
			{{- else }}
			{{ .TagInTagsElem }}:       Tags(updatedTags.IgnoreAws()),
			{{- end }}
		}

		_, err := conn.{{ .TagOp }}(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
		{{- if .TagOpBatchSize }}
		}
		{{- end }}
	}

	{{- end }}

	return nil
}
`

func awsServiceName(s string) (string, error) {
	s = strings.ToLower(s)

	if _, ok := awsServiceNames[s]; ok {
		return s, nil
	}

	switch s {
	case "cognitoidp":
		return "cognitoidentityprovider", nil
	case "dms":
		return "databasemigrationservice", nil
	case "ds":
		return "directoryservice", nil
	case "resourcegroupstagging":
		return "resourcegroupstaggingapi", nil
	case "serverlessapprepo":
		return "serverlessapplicationrepository", nil
	}

	if _, ok := awsServiceNames[fmt.Sprintf("%sservice", s)]; ok {
		return fmt.Sprintf("%sservice", s), nil
	}

	return "", fmt.Errorf("unable to find AWS service name for %s", s)
}

func awsServiceNameUpper(s string) (string, error) {
	s = strings.ToLower(s)

	if v, ok := awsServiceNames[s]; ok {
		return v, nil
	}

	switch s {
	case "cognitoidp":
		return awsServiceNames["cognitoidentityprovider"], nil
	case "dms":
		return awsServiceNames["databasemigrationservice"], nil
	case "ds":
		return awsServiceNames["directoryservice"], nil
	case "resourcegroupstagging":
		return awsServiceNames["resourcegroupstaggingapi"], nil
	case "serverlessapprepo":
		return awsServiceNames["serverlessapplicationrepository"], nil
	}

	if v, ok := awsServiceNames[fmt.Sprintf("%sservice", s)]; ok {
		return v, nil
	}

	return "", fmt.Errorf("unable to find AWS service name for %s", s)
}

func ToSnakeCase(str string) string {
	result := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	result = regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(result, "${1}_${2}")
	return strings.ToLower(result)
}

//awsServiceNames provides correct names and capitalization as used by AWS in client var
var awsServiceNames map[string]string

func init() {
	awsServiceNames = make(map[string]string)
	awsServiceNames["accessanalyzer"] = "AccessAnalyzer"
	awsServiceNames["acm"] = "ACM"
	awsServiceNames["acmpca"] = "ACMPCA"
	awsServiceNames["amplify"] = "Amplify"
	awsServiceNames["apigateway"] = "APIGateway"
	awsServiceNames["apigatewayv2"] = "ApiGatewayV2"
	awsServiceNames["appconfig"] = "AppConfig"
	awsServiceNames["appmesh"] = "AppMesh"
	awsServiceNames["apprunner"] = "AppRunner"
	awsServiceNames["appstream"] = "AppStream"
	awsServiceNames["appsync"] = "AppSync"
	awsServiceNames["athena"] = "Athena"
	awsServiceNames["autoscaling"] = "AutoScaling"
	awsServiceNames["backup"] = "Backup"
	awsServiceNames["batch"] = "Batch"
	awsServiceNames["cloud9"] = "Cloud9"
	awsServiceNames["cloudformation"] = "CloudFormation"
	awsServiceNames["cloudfront"] = "CloudFront"
	awsServiceNames["cloudhsmv2"] = "CloudHSMV2"
	awsServiceNames["cloudtrail"] = "CloudTrail"
	awsServiceNames["cloudwatch"] = "CloudWatch"
	awsServiceNames["cloudwatchevents"] = "CloudWatchEvents"
	awsServiceNames["cloudwatchlogs"] = "CloudWatchLogs"
	awsServiceNames["codeartifact"] = "CodeArtifact"
	awsServiceNames["codebuild"] = "CodeBuild"
	awsServiceNames["codecommit"] = "CodeCommit"
	awsServiceNames["codedeploy"] = "CodeDeploy"
	awsServiceNames["codepipeline"] = "CodePipeline"
	awsServiceNames["codestarconnections"] = "CodeStarConnections"
	awsServiceNames["codestarnotifications"] = "CodeStarNotifications"
	awsServiceNames["cognitoidentity"] = "CognitoIdentity"
	awsServiceNames["cognitoidentityprovider"] = "CognitoIdentityProvider"
	awsServiceNames["configservice"] = "ConfigService"
	awsServiceNames["databasemigrationservice"] = "DatabaseMigrationService"
	awsServiceNames["dataexchange"] = "DataExchange"
	awsServiceNames["datapipeline"] = "DataPipeline"
	awsServiceNames["datasync"] = "DataSync"
	awsServiceNames["dax"] = "DAX"
	awsServiceNames["devicefarm"] = "DeviceFarm"
	awsServiceNames["directconnect"] = "DirectConnect"
	awsServiceNames["directoryservice"] = "DirectoryService"
	awsServiceNames["dlm"] = "DLM"
	awsServiceNames["docdb"] = "DocDB"
	awsServiceNames["dynamodb"] = "DynamoDB"
	awsServiceNames["ec2"] = "EC2"
	awsServiceNames["ecr"] = "ECR"
	awsServiceNames["ecs"] = "ECS"
	awsServiceNames["efs"] = "EFS"
	awsServiceNames["eks"] = "EKS"
	awsServiceNames["elasticache"] = "ElastiCache"
	awsServiceNames["elasticbeanstalk"] = "ElasticBeanstalk"
	awsServiceNames["elasticsearchservice"] = "ElasticsearchService"
	awsServiceNames["elb"] = "ELB"
	awsServiceNames["elbv2"] = "ELBV2"
	awsServiceNames["emr"] = "EMR"
	awsServiceNames["firehose"] = "Firehose"
	awsServiceNames["fms"] = "FMS"
	awsServiceNames["fsx"] = "FSx"
	awsServiceNames["gamelift"] = "GameLift"
	awsServiceNames["glacier"] = "Glacier"
	awsServiceNames["globalaccelerator"] = "GlobalAccelerator"
	awsServiceNames["glue"] = "Glue"
	awsServiceNames["greengrass"] = "Greengrass"
	awsServiceNames["guardduty"] = "GuardDuty"
	awsServiceNames["iam"] = "IAM"
	awsServiceNames["imagebuilder"] = "Imagebuilder"
	awsServiceNames["inspector"] = "Inspector"
	awsServiceNames["iot"] = "IoT"
	awsServiceNames["iotanalytics"] = "IoTAnalytics"
	awsServiceNames["iotevents"] = "IoTEvents"
	awsServiceNames["kafka"] = "Kafka"
	awsServiceNames["kinesis"] = "Kinesis"
	awsServiceNames["kinesisanalytics"] = "KinesisAnalytics"
	awsServiceNames["kinesisanalyticsv2"] = "KinesisAnalyticsV2"
	awsServiceNames["kinesisvideo"] = "KinesisVideo"
	awsServiceNames["kms"] = "KMS"
	awsServiceNames["lakeformation"] = "LakeFormation"
	awsServiceNames["lambda"] = "Lambda"
	awsServiceNames["licensemanager"] = "LicenseManager"
	awsServiceNames["lightsail"] = "Lightsail"
	awsServiceNames["macie2"] = "Macie2"
	awsServiceNames["mediaconnect"] = "MediaConnect"
	awsServiceNames["mediaconvert"] = "MediaConvert"
	awsServiceNames["medialive"] = "MediaLive"
	awsServiceNames["mediapackage"] = "MediaPackage"
	awsServiceNames["mediastore"] = "MediaStore"
	awsServiceNames["mq"] = "MQ"
	awsServiceNames["mwaa"] = "MWAA"
	awsServiceNames["neptune"] = "Neptune"
	awsServiceNames["networkfirewall"] = "NetworkFirewall"
	awsServiceNames["networkmanager"] = "NetworkManager"
	awsServiceNames["opsworks"] = "OpsWorks"
	awsServiceNames["organizations"] = "Organizations"
	awsServiceNames["pinpoint"] = "Pinpoint"
	awsServiceNames["qldb"] = "QLDB"
	awsServiceNames["quicksight"] = "QuickSight"
	awsServiceNames["ram"] = "RAM"
	awsServiceNames["rds"] = "RDS"
	awsServiceNames["redshift"] = "Redshift"
	awsServiceNames["resourcegroups"] = "ResourceGroups"
	awsServiceNames["resourcegroupstaggingapi"] = "ResourceGroupsTaggingAPI"
	awsServiceNames["route53"] = "Route53"
	awsServiceNames["route53recoveryreadiness"] = "Route53RecoveryReadiness"
	awsServiceNames["route53resolver"] = "Route53Resolver"
	awsServiceNames["s3"] = "S3"
	awsServiceNames["s3control"] = "S3Control"
	awsServiceNames["sagemaker"] = "SageMaker"
	awsServiceNames["schemas"] = "Schemas"
	awsServiceNames["secretsmanager"] = "SecretsManager"
	awsServiceNames["securityhub"] = "SecurityHub"
	awsServiceNames["serverlessapplicationrepository"] = "ServerlessApplicationRepository"
	awsServiceNames["servicecatalog"] = "ServiceCatalog"
	awsServiceNames["servicediscovery"] = "ServiceDiscovery"
	awsServiceNames["sfn"] = "SFN"
	awsServiceNames["shield"] = "Shield"
	awsServiceNames["signer"] = "Signer"
	awsServiceNames["sns"] = "SNS"
	awsServiceNames["sqs"] = "SQS"
	awsServiceNames["ssm"] = "SSM"
	awsServiceNames["ssoadmin"] = "SSOAdmin"
	awsServiceNames["storagegateway"] = "StorageGateway"
	awsServiceNames["swf"] = "SWF"
	awsServiceNames["synthetics"] = "Synthetics"
	awsServiceNames["timestreamwrite"] = "TimestreamWrite"
	awsServiceNames["transfer"] = "Transfer"
	awsServiceNames["waf"] = "WAF"
	awsServiceNames["wafregional"] = "WAFRegional"
	awsServiceNames["wafv2"] = "WAFV2"
	awsServiceNames["wellarchitected"] = "WellArchitected"
	awsServiceNames["workdocs"] = "WorkDocs"
	awsServiceNames["worklink"] = "WorkLink"
	awsServiceNames["workspaces"] = "WorkSpaces"
	awsServiceNames["xray"] = "XRay"
}
