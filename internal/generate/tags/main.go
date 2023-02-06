//go:build generate
// +build generate

package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	v1 "github.com/hashicorp/terraform-provider-aws/internal/generate/tags/templates/v1"
	v2 "github.com/hashicorp/terraform-provider-aws/internal/generate/tags/templates/v2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	sdkV1 = 1
	sdkV2 = 2
)

var (
	getTag             = flag.Bool("GetTag", false, "whether to generate GetTag")
	listTags           = flag.Bool("ListTags", false, "whether to generate ListTags")
	serviceTagsMap     = flag.Bool("ServiceTagsMap", false, "whether to generate service tags for map")
	serviceTagsSlice   = flag.Bool("ServiceTagsSlice", false, "whether to generate service tags for slice")
	untagInNeedTagType = flag.Bool("UntagInNeedTagType", false, "whether Untag input needs tag type")
	updateTags         = flag.Bool("UpdateTags", false, "whether to generate UpdateTags")
	contextOnly        = flag.Bool("ContextOnly", false, "whether to only generate Context-aware functions")

	getTagFunc            = flag.String("GetTagFunc", "GetTag", "getTagFunc")
	listTagsFunc          = flag.String("ListTagsFunc", "ListTags", "listTagsFunc")
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
	tagTypeAddBoolElem    = flag.String("TagTypeAddBoolElem", "", "TagTypeAddBoolElem")
	tagTypeIDElem         = flag.String("TagTypeIDElem", "", "tagTypeIDElem")
	tagTypeKeyElem        = flag.String("TagTypeKeyElem", "Key", "tagTypeKeyElem")
	tagTypeValElem        = flag.String("TagTypeValElem", "Value", "tagTypeValElem")
	untagInCustomVal      = flag.String("UntagInCustomVal", "", "untagInCustomVal")
	untagInNeedTagKeyType = flag.String("UntagInNeedTagKeyType", "", "untagInNeedTagKeyType")
	untagInTagsElem       = flag.String("UntagInTagsElem", "TagKeys", "untagInTagsElem")
	untagOp               = flag.String("UntagOp", "UntagResource", "untagOp")
	updateTagsFunc        = flag.String("UpdateTagsFunc", "UpdateTags", "updateTagsFunc")

	parentNotFoundErrCode = flag.String("ParentNotFoundErrCode", "", "Parent 'NotFound' Error Code")
	parentNotFoundErrMsg  = flag.String("ParentNotFoundErrMsg", "", "Parent 'NotFound' Error Message")

	sdkVersion   = flag.Int("AWSSDKVersion", sdkV1, "Version of the AWS SDK Go to use i.e. 1 or 2")
	kvtValues    = flag.Bool("KVTValues", false, "Whether KVT string map is of string pointers")
	skipTypesImp = flag.Bool("SkipTypesImp", false, "Whether to skip importing types")
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

func newTemplateBody(version int, kvtValues bool) *TemplateBody {
	switch version {
	case sdkV1:
		return &TemplateBody{
			"\n" + v1.GetTagBody,
			v1.HeaderBody,
			"\n" + v1.ListTagsBody,
			"\n" + v1.ServiceTagsMapBody,
			"\n" + v1.ServiceTagsSliceBody,
			"\n" + v1.UpdateTagsBody,
		}
	case sdkV2:
		if kvtValues {
			return &TemplateBody{
				"\n" + v2.GetTagBody,
				v2.HeaderBody,
				"\n" + v2.ListTagsBody,
				"\n" + v2.ServiceTagsValueMapBody,
				"\n" + v2.ServiceTagsSliceBody,
				"\n" + v2.UpdateTagsBody,
			}
		}
		return &TemplateBody{
			"\n" + v2.GetTagBody,
			v2.HeaderBody,
			"\n" + v2.ListTagsBody,
			"\n" + v2.ServiceTagsMapBody,
			"\n" + v2.ServiceTagsSliceBody,
			"\n" + v2.UpdateTagsBody,
		}
	default:
		return nil
	}
}

type TemplateData struct {
	AWSService             string
	AWSServiceIfacePackage string
	ClientType             string
	ServicePackage         string

	GetTagFunc              string
	ListTagsFunc            string
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
	UpdateTagsFunc          string
	ContextOnly             bool

	// The following are specific to writing import paths in the `headerBody`;
	// to include the package, set the corresponding field's value to true
	ContextPkg      bool
	FmtPkg          bool
	HelperSchemaPkg bool
	SkipTypesImp    bool
	StrConvPkg      bool
	TfResourcePkg   bool
}

func main() {
	flag.Usage = usage
	flag.Parse()

	filename := `tags_gen.go`
	if args := flag.Args(); len(args) > 0 {
		filename = args[0]
	}

	g := common.NewGenerator()

	if *sdkVersion != sdkV1 && *sdkVersion != sdkV2 {
		g.Fatalf("AWS SDK Go Version %d not supported", *sdkVersion)
	}

	servicePackage := os.Getenv("GOPACKAGE")
	awsPkg, err := names.AWSGoPackage(servicePackage, *sdkVersion)

	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	var awsIntfPkg string
	if *sdkVersion == sdkV1 && (*getTag || *listTags || *updateTags) {
		awsIntfPkg = fmt.Sprintf("%[1]s/%[1]siface", awsPkg)
	}

	clientTypeName, err := names.AWSGoClientTypeName(servicePackage, *sdkVersion)

	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	var clientType string
	if *sdkVersion == sdkV1 {
		clientType = fmt.Sprintf("%siface.%sAPI", awsPkg, clientTypeName)
	} else {
		clientType = fmt.Sprintf("*%s.%s", awsPkg, clientTypeName)
	}

	tagPackage := awsPkg

	if tagPackage == "wafregional" {
		tagPackage = "waf"
		if *sdkVersion == sdkV1 {
			awsPkg = ""
		}
	}

	templateData := TemplateData{
		AWSService:             awsPkg,
		AWSServiceIfacePackage: awsIntfPkg,
		ClientType:             clientType,
		ServicePackage:         servicePackage,

		ContextPkg:      *sdkVersion == sdkV2 || (*getTag || *listTags || *updateTags),
		FmtPkg:          *updateTags,
		HelperSchemaPkg: awsPkg == "autoscaling",
		SkipTypesImp:    *skipTypesImp,
		StrConvPkg:      awsPkg == "autoscaling",
		TfResourcePkg:   *getTag,

		GetTagFunc:              *getTagFunc,
		ListTagsFunc:            *listTagsFunc,
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
		TagTypeAddBoolElem:      *tagTypeAddBoolElem,
		TagTypeAddBoolElemSnake: toSnakeCase(*tagTypeAddBoolElem),
		TagTypeIDElem:           *tagTypeIDElem,
		TagTypeKeyElem:          *tagTypeKeyElem,
		TagTypeValElem:          *tagTypeValElem,
		UntagInCustomVal:        *untagInCustomVal,
		UntagInNeedTagKeyType:   *untagInNeedTagKeyType,
		UntagInNeedTagType:      *untagInNeedTagType,
		UntagInTagsElem:         *untagInTagsElem,
		UntagOp:                 *untagOp,
		UpdateTagsFunc:          *updateTagsFunc,
		ContextOnly:             *contextOnly,
	}

	templateBody := newTemplateBody(*sdkVersion, *kvtValues)
	d := g.NewGoFileDestination(filename)

	if *getTag || *listTags || *serviceTagsMap || *serviceTagsSlice || *updateTags {
		// If you intend to only generate Tags and KeyValueTags helper methods,
		// the corresponding aws-sdk-go	service package does not need to be imported
		if !*getTag && !*listTags && !*serviceTagsSlice && !*updateTags {
			templateData.AWSService = ""
			templateData.TagPackage = ""
		}

		if err := d.WriteTemplate("header", templateBody.header, templateData); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *getTag {
		if err := d.WriteTemplate("gettag", templateBody.getTag, templateData); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *listTags {
		if err := d.WriteTemplate("listtags", templateBody.listTags, templateData); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *serviceTagsMap {
		if err := d.WriteTemplate("servicetagsmap", templateBody.serviceTagsMap, templateData); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *serviceTagsSlice {
		if err := d.WriteTemplate("servicetagsslice", templateBody.serviceTagsSlice, templateData); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *updateTags {
		if err := d.WriteTemplate("updatetags", templateBody.updateTags, templateData); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

func toSnakeCase(str string) string {
	result := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	result = regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(result, "${1}_${2}")
	return strings.ToLower(result)
}
