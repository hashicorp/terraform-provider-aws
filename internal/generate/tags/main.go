// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	v1 "github.com/hashicorp/terraform-provider-aws/internal/generate/tags/templates/v1"
	v2 "github.com/hashicorp/terraform-provider-aws/internal/generate/tags/templates/v2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	sdkV1 = 1
	sdkV2 = 2
)

const (
	defaultListTagsFunc           = "listTags"
	defaultUpdateTagsFunc         = "updateTags"
	defaultWaitTagsPropagatedFunc = "waitTagsPropagated"
)

var (
	createTags               = flag.Bool("CreateTags", false, "whether to generate CreateTags")
	getTag                   = flag.Bool("GetTag", false, "whether to generate GetTag")
	listTags                 = flag.Bool("ListTags", false, "whether to generate ListTags")
	serviceTagsMap           = flag.Bool("ServiceTagsMap", false, "whether to generate service tags for map")
	serviceTagsSlice         = flag.Bool("ServiceTagsSlice", false, "whether to generate service tags for slice")
	untagInNeedTagType       = flag.Bool("UntagInNeedTagType", false, "whether Untag input needs tag type")
	updateTags               = flag.Bool("UpdateTags", false, "whether to generate UpdateTags")
	updateTagsNoIgnoreSystem = flag.Bool("UpdateTagsNoIgnoreSystem", false, "whether to not ignore system tags in UpdateTags")
	waitForPropagation       = flag.Bool("Wait", false, "whether to generate WaitTagsPropagated")

	createTagsFunc          = flag.String("CreateTagsFunc", "createTags", "createTagsFunc")
	getTagFunc              = flag.String("GetTagFunc", "GetTag", "getTagFunc")
	getTagsInFunc           = flag.String("GetTagsInFunc", "getTagsIn", "getTagsInFunc")
	keyValueTagsFunc        = flag.String("KeyValueTagsFunc", "KeyValueTags", "keyValueTagsFunc")
	listTagsFunc            = flag.String("ListTagsFunc", defaultListTagsFunc, "listTagsFunc")
	listTagsInFiltIDName    = flag.String("ListTagsInFiltIDName", "", "listTagsInFiltIDName")
	listTagsInIDElem        = flag.String("ListTagsInIDElem", "ResourceArn", "listTagsInIDElem")
	listTagsInIDNeedSlice   = flag.String("ListTagsInIDNeedSlice", "", "listTagsInIDNeedSlice")
	listTagsOp              = flag.String("ListTagsOp", "ListTagsForResource", "listTagsOp")
	listTagsOpPaginated     = flag.Bool("ListTagsOpPaginated", false, "whether ListTagsOp is paginated")
	listTagsOutTagsElem     = flag.String("ListTagsOutTagsElem", "Tags", "listTagsOutTagsElem")
	setTagsOutFunc          = flag.String("SetTagsOutFunc", "setTagsOut", "setTagsOutFunc")
	tagInCustomVal          = flag.String("TagInCustomVal", "", "tagInCustomVal")
	tagInIDElem             = flag.String("TagInIDElem", "ResourceArn", "tagInIDElem")
	tagInIDNeedSlice        = flag.String("TagInIDNeedSlice", "", "tagInIDNeedSlice")
	tagInIDNeedValueSlice   = flag.String("TagInIDNeedValueSlice", "", "tagInIDNeedValueSlice")
	tagInTagsElem           = flag.String("TagInTagsElem", "Tags", "tagInTagsElem")
	tagKeyType              = flag.String("TagKeyType", "", "tagKeyType")
	tagOp                   = flag.String("TagOp", "TagResource", "tagOp")
	tagOpBatchSize          = flag.String("TagOpBatchSize", "", "tagOpBatchSize")
	tagResTypeElem          = flag.String("TagResTypeElem", "", "tagResTypeElem")
	tagType                 = flag.String("TagType", "Tag", "tagType")
	tagType2                = flag.String("TagType2", "", "tagType")
	tagTypeAddBoolElem      = flag.String("TagTypeAddBoolElem", "", "TagTypeAddBoolElem")
	tagTypeIDElem           = flag.String("TagTypeIDElem", "", "tagTypeIDElem")
	tagTypeKeyElem          = flag.String("TagTypeKeyElem", "Key", "tagTypeKeyElem")
	tagTypeValElem          = flag.String("TagTypeValElem", "Value", "tagTypeValElem")
	tagsFunc                = flag.String("TagsFunc", "Tags", "tagsFunc")
	untagInCustomVal        = flag.String("UntagInCustomVal", "", "untagInCustomVal")
	untagInNeedTagKeyType   = flag.String("UntagInNeedTagKeyType", "", "untagInNeedTagKeyType")
	untagInTagsElem         = flag.String("UntagInTagsElem", "TagKeys", "untagInTagsElem")
	untagOp                 = flag.String("UntagOp", "UntagResource", "untagOp")
	updateTagsFunc          = flag.String("UpdateTagsFunc", defaultUpdateTagsFunc, "updateTagsFunc")
	waitTagsPropagatedFunc  = flag.String("WaitFunc", defaultWaitTagsPropagatedFunc, "waitFunc")
	waitContinuousOccurence = flag.Int("WaitContinuousOccurence", 0, "ContinuousTargetOccurence for Wait function")
	waitDelay               = flag.Duration("WaitDelay", 0, "Delay for Wait function")
	waitMinTimeout          = flag.Duration("WaitMinTimeout", 0, `"MinTimeout" (minimum poll interval) for Wait function`)
	waitPollInterval        = flag.Duration("WaitPollInterval", 0, "PollInterval for Wait function")
	waitTimeout             = flag.Duration("WaitTimeout", 0, "Timeout for Wait function")

	parentNotFoundErrCode = flag.String("ParentNotFoundErrCode", "", "Parent 'NotFound' Error Code")
	parentNotFoundErrMsg  = flag.String("ParentNotFoundErrMsg", "", "Parent 'NotFound' Error Message")

	sdkServicePackage = flag.String("AWSSDKServicePackage", "", "AWS Go SDK package to use. Defaults to the provider service package name.")
	sdkVersion        = flag.Int("AWSSDKVersion", sdkV1, "Version of the AWS Go SDK to use i.e. 1 or 2")
	kvtValues         = flag.Bool("KVTValues", false, "Whether KVT string map is of string pointers")
	skipAWSImp        = flag.Bool("SkipAWSImp", false, "Whether to skip importing the AWS Go SDK aws package") // nosemgrep:ci.aws-in-var-name
	skipNamesImp      = flag.Bool("SkipNamesImp", false, "Whether to skip importing names")
	skipServiceImp    = flag.Bool("SkipAWSServiceImp", false, "Whether to skip importing the AWS service package")
	skipTypesImp      = flag.Bool("SkipTypesImp", false, "Whether to skip importing types")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tmain.go [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

type TemplateBody struct {
	getTag             string
	header             string
	listTags           string
	serviceTagsMap     string
	serviceTagsSlice   string
	updateTags         string
	waitTagsPropagated string
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
			"\n" + v1.WaitTagsPropagatedBody,
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
				"\n" + v2.WaitTagsPropagatedBody,
			}
		}
		return &TemplateBody{
			"\n" + v2.GetTagBody,
			v2.HeaderBody,
			"\n" + v2.ListTagsBody,
			"\n" + v2.ServiceTagsMapBody,
			"\n" + v2.ServiceTagsSliceBody,
			"\n" + v2.UpdateTagsBody,
			"\n" + v2.WaitTagsPropagatedBody,
		}
	default:
		return nil
	}
}

type TemplateData struct {
	AWSService             string
	AWSServiceIfacePackage string
	ClientType             string
	ProviderNameUpper      string
	ServicePackage         string

	CreateTagsFunc          string
	GetTagFunc              string
	GetTagsInFunc           string
	KeyValueTagsFunc        string
	ListTagsFunc            string
	ListTagsInFiltIDName    string
	ListTagsInIDElem        string
	ListTagsInIDNeedSlice   string
	ListTagsOp              string
	ListTagsOpPaginated     bool
	ListTagsOutTagsElem     string
	ParentNotFoundErrCode   string
	ParentNotFoundErrMsg    string
	RetryCreateOnNotFound   string
	SetTagsOutFunc          string
	TagInCustomVal          string
	TagInIDElem             string
	TagInIDNeedSlice        string
	TagInIDNeedValueSlice   string
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
	TagsFunc                string
	UntagInCustomVal        string
	UntagInNeedTagKeyType   string
	UntagInNeedTagType      bool
	UntagInTagsElem         string
	UntagOp                 string
	UpdateTagsFunc          string
	UpdateTagsIgnoreSystem  bool
	WaitForPropagation      bool
	WaitTagsPropagatedFunc  string
	WaitContinuousOccurence int
	WaitDelay               string
	WaitMinTimeout          string
	WaitPollInterval        string
	WaitTimeout             string

	// The following are specific to writing import paths in the `headerBody`;
	// to include the package, set the corresponding field's value to true
	ConnsPkg         bool
	FmtPkg           bool
	HelperSchemaPkg  bool
	InternalTypesPkg bool
	LoggingPkg       bool
	NamesPkg         bool
	SkipAWSImp       bool
	SkipServiceImp   bool
	SkipTypesImp     bool
	TfLogPkg         bool
	TfResourcePkg    bool
	TimePkg          bool

	IsDefaultListTags   bool
	IsDefaultUpdateTags bool
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
	if *sdkServicePackage == "" {
		sdkServicePackage = &servicePackage
	}

	g.Infof("Generating internal/service/%s/%s", servicePackage, filename)

	awsPkg, err := names.AWSGoPackage(*sdkServicePackage, *sdkVersion)
	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	var awsIntfPkg string
	if *sdkVersion == sdkV1 && (*getTag || *listTags || *updateTags) {
		awsIntfPkg = fmt.Sprintf("%[1]s/%[1]siface", awsPkg)
	}

	clientTypeName, err := names.AWSGoClientTypeName(*sdkServicePackage, *sdkVersion)

	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	providerNameUpper, err := names.ProviderNameUpper(*sdkServicePackage)

	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	createTagsFunc := *createTagsFunc
	if *createTags && !*updateTags {
		g.Infof("CreateTags only valid with UpdateTags")
		createTagsFunc = ""
	} else if !*createTags {
		createTagsFunc = ""
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
		ProviderNameUpper:      providerNameUpper,
		ServicePackage:         servicePackage,

		ConnsPkg:         (*listTags && *listTagsFunc == defaultListTagsFunc) || (*updateTags && *updateTagsFunc == defaultUpdateTagsFunc),
		FmtPkg:           *updateTags,
		HelperSchemaPkg:  awsPkg == "autoscaling",
		InternalTypesPkg: (*listTags && *listTagsFunc == defaultListTagsFunc) || *serviceTagsMap || *serviceTagsSlice,
		LoggingPkg:       *updateTags,
		NamesPkg:         *updateTags && !*skipNamesImp,
		SkipAWSImp:       *skipAWSImp,
		SkipServiceImp:   *skipServiceImp,
		SkipTypesImp:     *skipTypesImp,
		TfLogPkg:         *updateTags,
		TfResourcePkg:    (*getTag || *waitForPropagation),
		TimePkg:          *waitForPropagation,

		CreateTagsFunc:          createTagsFunc,
		GetTagFunc:              *getTagFunc,
		GetTagsInFunc:           *getTagsInFunc,
		KeyValueTagsFunc:        *keyValueTagsFunc,
		ListTagsFunc:            *listTagsFunc,
		ListTagsInFiltIDName:    *listTagsInFiltIDName,
		ListTagsInIDElem:        *listTagsInIDElem,
		ListTagsInIDNeedSlice:   *listTagsInIDNeedSlice,
		ListTagsOp:              *listTagsOp,
		ListTagsOpPaginated:     *listTagsOpPaginated,
		ListTagsOutTagsElem:     *listTagsOutTagsElem,
		ParentNotFoundErrCode:   *parentNotFoundErrCode,
		ParentNotFoundErrMsg:    *parentNotFoundErrMsg,
		SetTagsOutFunc:          *setTagsOutFunc,
		TagInCustomVal:          *tagInCustomVal,
		TagInIDElem:             *tagInIDElem,
		TagInIDNeedSlice:        *tagInIDNeedSlice,
		TagInIDNeedValueSlice:   *tagInIDNeedValueSlice,
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
		TagsFunc:                *tagsFunc,
		UntagInCustomVal:        *untagInCustomVal,
		UntagInNeedTagKeyType:   *untagInNeedTagKeyType,
		UntagInNeedTagType:      *untagInNeedTagType,
		UntagInTagsElem:         *untagInTagsElem,
		UntagOp:                 *untagOp,
		UpdateTagsFunc:          *updateTagsFunc,
		UpdateTagsIgnoreSystem:  !*updateTagsNoIgnoreSystem,
		WaitForPropagation:      *waitForPropagation,
		WaitTagsPropagatedFunc:  *waitTagsPropagatedFunc,
		WaitContinuousOccurence: *waitContinuousOccurence,
		WaitDelay:               formatDuration(*waitDelay),
		WaitMinTimeout:          formatDuration(*waitMinTimeout),
		WaitPollInterval:        formatDuration(*waitPollInterval),
		WaitTimeout:             formatDuration(*waitTimeout),

		IsDefaultListTags:   *listTagsFunc == defaultListTagsFunc,
		IsDefaultUpdateTags: *updateTagsFunc == defaultUpdateTagsFunc,
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

	if *waitForPropagation {
		if err := d.WriteTemplate("waittagspropagated", templateBody.waitTagsPropagated, templateData); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

func toSnakeCase(str string) string {
	result := regexache.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	result = regexache.MustCompile("([0-9a-z])([A-Z])").ReplaceAllString(result, "${1}_${2}")
	return strings.ToLower(result)
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}

	var buf []string
	if h := d.Hours(); h >= 1 {
		buf = append(buf, fmt.Sprintf("%d * time.Hour", int64(h)))
		d = d - time.Duration(int64(h)*int64(time.Hour))
	}
	if m := d.Minutes(); m >= 1 {
		buf = append(buf, fmt.Sprintf("%d * time.Minute", int64(m)))
		d = d - time.Duration(int64(m)*int64(time.Minute))
	}
	if s := d.Seconds(); s >= 1 {
		buf = append(buf, fmt.Sprintf("%d * time.Second", int64(s)))
		d = d - time.Duration(int64(s)*int64(time.Second))
	}
	if ms := d.Milliseconds(); ms >= 1 {
		buf = append(buf, fmt.Sprintf("%d * time.Millisecond", int64(ms)))
	}
	// Ignoring anything below milliseconds

	return strings.Join(buf, " + ")
}
