//go:build generate
// +build generate

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

	v1 "github.com/hashicorp/terraform-provider-aws/internal/generate/tags/templates/v1"
	v2 "github.com/hashicorp/terraform-provider-aws/internal/generate/tags/templates/v2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	filename = `tags_gen.go`

	AwsSdkV1 = 1
	AwsSdkV2 = 2
)

var (
	getTag             = flag.Bool("GetTag", false, "whether to generate GetTag")
	listTags           = flag.Bool("ListTags", false, "whether to generate ListTags")
	serviceTagsMap     = flag.Bool("ServiceTagsMap", false, "whether to generate service tags for map")
	serviceTagsSlice   = flag.Bool("ServiceTagsSlice", false, "whether to generate service tags for slice")
	untagInNeedTagType = flag.Bool("UntagInNeedTagType", false, "whether Untag input needs tag type")
	updateTags         = flag.Bool("UpdateTags", false, "whether to generate UpdateTags")

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
	untagInTagsElem       = flag.String("UntagInTagsElem", "TagKeys", "untagInTagsElem")
	untagOp               = flag.String("UntagOp", "UntagResource", "untagOp")

	parentNotFoundErrCode = flag.String("ParentNotFoundErrCode", "", "Parent 'NotFound' Error Code")
	parentNotFoundErrMsg  = flag.String("ParentNotFoundErrMsg", "", "Parent 'NotFound' Error Message")

	awsSdkVersion = flag.Int("AwsSdkVersion", AwsSdkV1, "Version of the AWS SDK Go to use i.e. 1 or 2")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

type TemplateBody struct {
	getTag           string
	header           string
	listTags         string
	serviceTagsMap   string
	serviceTagsSlice string
	updateTags       string
}

func NewTemplateBody(version int) *TemplateBody {
	switch version {
	case AwsSdkV1:
		return &TemplateBody{
			v1.GetTagBody,
			v1.HeaderBody,
			v1.ListTagsBody,
			v1.ServiceTagsMapBody,
			v1.ServiceTagsSliceBody,
			v1.UpdateTagsBody,
		}
	case AwsSdkV2:
		return &TemplateBody{
			v2.GetTagBody,
			v2.HeaderBody,
			v2.ListTagsBody,
			v2.ServiceTagsMapBody,
			v2.ServiceTagsSliceBody,
			v2.UpdateTagsBody,
		}
	default:
		return nil
	}
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
	ParentNotFoundErrCode   string
	ParentNotFoundErrMsg    string
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
	UntagInNeedTagType      bool
	UntagInTagsElem         string
	UntagOp                 string

	// The following are specific to writing import paths in the `headerBody`;
	// to include the package, set the corresponding field's value to true
	FmtPkg          bool
	HelperSchemaPkg bool
	StrConvPkg      bool
	TfResourcePkg   bool

	AwsSdkVersion int
}

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *awsSdkVersion != AwsSdkV1 && *awsSdkVersion != AwsSdkV2 {
		log.Fatalf("AWS SDK Go Version %d not supported", *awsSdkVersion)
	}

	servicePackage := os.Getenv("GOPACKAGE")
	awsPkg, err := names.AWSGoPackage(servicePackage, *awsSdkVersion)

	if err != nil {
		log.Fatalf("encountered: %s", err)
	}

	clientName, err := names.AWSGoClientName(servicePackage, *awsSdkVersion)

	if err != nil {
		log.Fatalf("encountered: %s", err)
	}

	clientType := fmt.Sprintf("*%s.%s", awsPkg, clientName)

	tagPackage := awsPkg

	if tagPackage == "wafregional" {
		tagPackage = "waf"
	}

	templateData := TemplateData{
		AWSService:     awsPkg,
		ClientType:     clientType,
		ServicePackage: servicePackage,

		FmtPkg:          *updateTags,
		HelperSchemaPkg: awsPkg == "autoscaling",
		StrConvPkg:      awsPkg == "autoscaling",
		TfResourcePkg:   *getTag,

		ListTagsInFiltIDName:    *listTagsInFiltIDName,
		ListTagsInIDElem:        *listTagsInIDElem,
		ListTagsInIDNeedSlice:   *listTagsInIDNeedSlice,
		ListTagsOp:              *listTagsOp,
		ListTagsOutTagsElem:     *listTagsOutTagsElem,
		ParentNotFoundErrCode:   *parentNotFoundErrCode,
		ParentNotFoundErrMsg:    *parentNotFoundErrMsg,
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

	templateBody := NewTemplateBody(*awsSdkVersion)

	if *getTag || *listTags || *serviceTagsMap || *serviceTagsSlice || *updateTags {
		// If you intend to only generate Tags and KeyValueTags helper methods,
		// the corresponding aws-sdk-go	 service package does not need to be imported
		if !*getTag && !*listTags && !*serviceTagsSlice && !*updateTags {
			templateData.AWSService = ""
		}
		writeTemplate(templateBody.header, "header", templateData)
	}

	if *getTag {
		writeTemplate(templateBody.getTag, "gettag", templateData)
	}

	if *listTags {
		writeTemplate(templateBody.listTags, "listtags", templateData)
	}

	if *serviceTagsMap {
		writeTemplate(templateBody.serviceTagsMap, "servicetagsmap", templateData)
	}

	if *serviceTagsSlice {
		writeTemplate(templateBody.serviceTagsSlice, "servicetagsslice", templateData)
	}

	if *updateTags {
		writeTemplate(templateBody.updateTags, "updatetags", templateData)
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

func ToSnakeCase(str string) string {
	result := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	result = regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(result, "${1}_${2}")
	return strings.ToLower(result)
}
