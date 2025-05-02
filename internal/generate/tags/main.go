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
	"text/template"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/tags/templates"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/hashicorp/terraform-provider-aws/names/data"
)

const (
	defaultCreateTagsFunc         = "createTags"
	defaultGetTagFunc             = "findTag"
	defaultGetTagsInFunc          = "getTagsIn"
	defaultKeyValueTagsFunc       = "keyValueTags"
	defaultListTagsFunc           = "listTags"
	defaultSetTagsOutFunc         = "setTagsOut"
	defaultTagsFunc               = "svcTags"
	defaultUpdateTagsFunc         = "updateTags"
	defaultWaitTagsPropagatedFunc = "waitTagsPropagated"
)

var (
	sdkServicePackage = flag.String("AWSSDKServicePackage", "", "AWS Go SDK package to use. Defaults to the provider service package name.")

	createTags               = flag.Bool("CreateTags", false, "whether to generate CreateTags")
	createTagsFunc           = flag.String("CreateTagsFunc", defaultCreateTagsFunc, "createTagsFunc")
	getTag                   = flag.Bool("GetTag", false, "whether to generate GetTag")
	getTagFunc               = flag.String("GetTagFunc", defaultGetTagFunc, "getTagFunc")
	listTags                 = flag.Bool("ListTags", false, "whether to generate ListTags")
	listTagsFunc             = flag.String("ListTagsFunc", defaultListTagsFunc, "listTagsFunc")
	updateTags               = flag.Bool("UpdateTags", false, "whether to generate UpdateTags")
	updateTagsFunc           = flag.String("UpdateTagsFunc", defaultUpdateTagsFunc, "updateTagsFunc")
	updateTagsNoIgnoreSystem = flag.Bool("UpdateTagsNoIgnoreSystem", false, "whether to not ignore system tags in UpdateTags")

	serviceTagsMap   = flag.Bool("ServiceTagsMap", false, "whether to generate service tags for map")
	kvtValues        = flag.Bool("KVTValues", false, "Whether KVT string map is of string pointers")
	emptyMap         = flag.Bool("EmptyMap", false, "Whether KVT string map should be empty for no tags")
	serviceTagsSlice = flag.Bool("ServiceTagsSlice", false, "whether to generate service tags for slice")

	keyValueTagsFunc = flag.String("KeyValueTagsFunc", defaultKeyValueTagsFunc, "keyValueTagsFunc")
	tagsFunc         = flag.String("TagsFunc", defaultTagsFunc, "tagsFunc")
	getTagsInFunc    = flag.String("GetTagsInFunc", defaultGetTagsInFunc, "getTagsInFunc")
	setTagsOutFunc   = flag.String("SetTagsOutFunc", defaultSetTagsOutFunc, "setTagsOutFunc")

	waitForPropagation      = flag.Bool("Wait", false, "whether to generate WaitTagsPropagated")
	waitTagsPropagatedFunc  = flag.String("WaitFunc", defaultWaitTagsPropagatedFunc, "waitFunc")
	waitContinuousOccurence = flag.Int("WaitContinuousOccurence", 0, "ContinuousTargetOccurence for Wait function")
	waitFuncComparator      = flag.String("WaitFuncComparator", "Equal", "waitFuncComparator")
	waitDelay               = flag.Duration("WaitDelay", 0, "Delay for Wait function")
	waitMinTimeout          = flag.Duration("WaitMinTimeout", 0, `"MinTimeout" (minimum poll interval) for Wait function`)
	waitPollInterval        = flag.Duration("WaitPollInterval", 0, "PollInterval for Wait function")
	waitTimeout             = flag.Duration("WaitTimeout", 0, "Timeout for Wait function")

	listTagsInFiltIDName       = flag.String("ListTagsInFiltIDName", "", "listTagsInFiltIDName")
	listTagsInIDElem           = flag.String("ListTagsInIDElem", "ResourceArn", "listTagsInIDElem")
	listTagsInIDNeedValueSlice = flag.Bool("ListTagsInIDNeedValueSlice", false, "listTagsInIDNeedSlice")
	listTagsOp                 = flag.String("ListTagsOp", "ListTagsForResource", "listTagsOp")
	listTagsOpPaginated        = flag.Bool("ListTagsOpPaginated", false, "whether ListTagsOp is paginated")
	listTagsOutTagsElem        = flag.String("ListTagsOutTagsElem", "Tags", "listTagsOutTagsElem")

	retryErrorCode        = flag.String("RetryErrorCode", "", "error code to retry, must be used with RetryTagOps")
	retryErrorMessage     = flag.String("RetryErrorMessage", "", "error message to retry, must be used with RetryTagOps")
	retryTagOps           = flag.Bool("RetryTagOps", false, "whether to retry tag operations")
	retryTagsListTagsType = flag.String("RetryTagsListTagsType", "", "type of the first ListTagsOp return value such as ListTagsForResourceOutput, must be used with RetryTagOps")
	retryTimeout          = flag.Duration("RetryTimeout", 1*time.Minute, "amount of time tag operations should retry")

	tagInCustomVal        = flag.String("TagInCustomVal", "", "tagInCustomVal")
	tagInIDElem           = flag.String("TagInIDElem", "ResourceArn", "tagInIDElem")
	tagInIDNeedValueSlice = flag.Bool("TagInIDNeedValueSlice", false, "tagInIDNeedValueSlice")
	tagInTagsElem         = flag.String("TagInTagsElem", "Tags", "tagInTagsElem")
	tagKeyType            = flag.String("TagKeyType", "", "tagKeyType")
	tagOp                 = flag.String("TagOp", "TagResource", "tagOp")
	tagOpBatchSize        = flag.Int("TagOpBatchSize", 0, "tagOpBatchSize")
	tagResTypeElem        = flag.String("TagResTypeElem", "", "tagResTypeElem")
	tagResTypeElemType    = flag.String("TagResTypeElemType", "", "tagResTypeElemType")
	tagType               = flag.String("TagType", "Tag", "tagType")
	tagType2              = flag.String("TagType2", "", "tagType")
	tagTypeAddBoolElem    = flag.String("TagTypeAddBoolElem", "", "TagTypeAddBoolElem")
	tagTypeIDElem         = flag.String("TagTypeIDElem", "", "tagTypeIDElem")
	tagTypeKeyElem        = flag.String("TagTypeKeyElem", "Key", "tagTypeKeyElem")
	tagTypeValElem        = flag.String("TagTypeValElem", "Value", "tagTypeValElem")

	untagInCustomVal      = flag.String("UntagInCustomVal", "", "untagInCustomVal")
	untagInNeedTagKeyType = flag.Bool("UntagInNeedTagKeyType", false, "untagInNeedTagKeyType")
	untagInNeedTagType    = flag.Bool("UntagInNeedTagType", false, "whether Untag input needs tag type")
	untagInTagsElem       = flag.String("UntagInTagsElem", "TagKeys", "untagInTagsElem")
	untagOp               = flag.String("UntagOp", "UntagResource", "untagOp")

	parentNotFoundErrCode = flag.String("ParentNotFoundErrCode", "", "Parent 'NotFound' Error Code")
	parentNotFoundErrMsg  = flag.String("ParentNotFoundErrMsg", "", "Parent 'NotFound' Error Message")
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

func newTemplateBody(kvtValues bool) *TemplateBody {
	if kvtValues {
		return &TemplateBody{
			getTag:             "\n" + templates.GetTagBody,
			header:             templates.HeaderBody,
			listTags:           "\n" + templates.ListTagsBody,
			serviceTagsMap:     "\n" + templates.ServiceTagsValueMapBody,
			serviceTagsSlice:   "\n" + templates.ServiceTagsSliceBody,
			updateTags:         "\n" + templates.UpdateTagsBody,
			waitTagsPropagated: "\n" + templates.WaitTagsPropagatedBody,
		}
	}
	return &TemplateBody{
		getTag:             "\n" + templates.GetTagBody,
		header:             templates.HeaderBody,
		listTags:           "\n" + templates.ListTagsBody,
		serviceTagsMap:     "\n" + templates.ServiceTagsMapBody,
		serviceTagsSlice:   "\n" + templates.ServiceTagsSliceBody,
		updateTags:         "\n" + templates.UpdateTagsBody,
		waitTagsPropagated: "\n" + templates.WaitTagsPropagatedBody,
	}
}

type TemplateData struct {
	AWSService        string
	ClientType        string
	ProviderNameUpper string
	ServicePackage    string

	CreateTagsFunc             string
	EmptyMap                   bool
	GetTagFunc                 string
	GetTagsInFunc              string
	KeyValueTagsFunc           string
	ListTagsFunc               string
	ListTagsInFiltIDName       string
	ListTagsInIDElem           string
	ListTagsInIDNeedSlice      string
	ListTagsInIDNeedValueSlice bool
	ListTagsOp                 string
	ListTagsOpPaginated        bool
	ListTagsOutTagsElem        string
	ParentNotFoundErrCode      string
	ParentNotFoundErrMsg       string
	RetryErrorCode             string
	RetryErrorMessage          string
	RetryTagOps                bool
	RetryTagsListTagsType      string
	RetryTimeout               string
	ServiceTagsMap             bool
	SetTagsOutFunc             string
	TagInCustomVal             string
	TagInIDElem                string
	TagInIDNeedValueSlice      bool
	TagInTagsElem              string
	TagKeyType                 string
	TagOp                      string
	TagOpBatchSize             int
	TagPackage                 string
	TagResTypeElem             string
	TagResTypeElemType         string
	TagType                    string
	TagType2                   string
	TagTypeAddBoolElem         string
	TagTypeIDElem              string
	TagTypeKeyElem             string
	TagTypeValElem             string
	TagsFunc                   string
	UntagInCustomVal           string
	UntagInNeedTagKeyType      bool
	UntagInNeedTagType         bool
	UntagInTagsElem            string
	UntagOp                    string
	UpdateTagsFunc             string
	UpdateTagsIgnoreSystem     bool
	WaitForPropagation         bool
	WaitTagsPropagatedFunc     string
	WaitContinuousOccurence    int
	WaitDelay                  string
	WaitFuncComparator         string
	WaitMinTimeout             string
	WaitPollInterval           string
	WaitTimeout                string

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

	servicePackage := os.Getenv("GOPACKAGE")
	if *sdkServicePackage == "" {
		sdkServicePackage = &servicePackage
	}

	g.Infof("Generating internal/service/%s/%s", servicePackage, filename)

	service, err := data.LookupService(*sdkServicePackage)
	if err != nil {
		g.Fatalf("encountered: %s", err)
	}

	awsPkg := service.GoPackageName()

	createTagsFunc := *createTagsFunc
	if *createTags && !*updateTags {
		g.Infof("CreateTags only valid with UpdateTags")
		createTagsFunc = ""
	} else if !*createTags {
		createTagsFunc = ""
	}

	clientType := fmt.Sprintf("*%s.Client", awsPkg)
	tagPackage := awsPkg
	providerNameUpper := service.ProviderNameUpper()
	templateData := TemplateData{
		AWSService:        awsPkg,
		ClientType:        clientType,
		ProviderNameUpper: providerNameUpper,
		ServicePackage:    servicePackage,

		CreateTagsFunc:             createTagsFunc,
		EmptyMap:                   *emptyMap,
		GetTagFunc:                 *getTagFunc,
		GetTagsInFunc:              *getTagsInFunc,
		KeyValueTagsFunc:           *keyValueTagsFunc,
		ListTagsFunc:               *listTagsFunc,
		ListTagsInFiltIDName:       *listTagsInFiltIDName,
		ListTagsInIDElem:           *listTagsInIDElem,
		ListTagsInIDNeedValueSlice: *listTagsInIDNeedValueSlice,
		ListTagsOp:                 *listTagsOp,
		ListTagsOpPaginated:        *listTagsOpPaginated,
		ListTagsOutTagsElem:        *listTagsOutTagsElem,
		ParentNotFoundErrCode:      *parentNotFoundErrCode,
		ParentNotFoundErrMsg:       *parentNotFoundErrMsg,
		RetryErrorCode:             *retryErrorCode,
		RetryErrorMessage:          *retryErrorMessage,
		RetryTagOps:                *retryTagOps,
		RetryTagsListTagsType:      *retryTagsListTagsType,
		RetryTimeout:               formatDuration(*retryTimeout),
		ServiceTagsMap:             *serviceTagsMap,
		SetTagsOutFunc:             *setTagsOutFunc,
		TagInCustomVal:             *tagInCustomVal,
		TagInIDElem:                *tagInIDElem,
		TagInIDNeedValueSlice:      *tagInIDNeedValueSlice,
		TagInTagsElem:              *tagInTagsElem,
		TagKeyType:                 *tagKeyType,
		TagOp:                      *tagOp,
		TagOpBatchSize:             *tagOpBatchSize,
		TagPackage:                 tagPackage,
		TagResTypeElem:             *tagResTypeElem,
		TagResTypeElemType:         *tagResTypeElemType,
		TagType:                    *tagType,
		TagType2:                   *tagType2,
		TagTypeAddBoolElem:         *tagTypeAddBoolElem,
		TagTypeIDElem:              *tagTypeIDElem,
		TagTypeKeyElem:             *tagTypeKeyElem,
		TagTypeValElem:             *tagTypeValElem,
		TagsFunc:                   *tagsFunc,
		UntagInCustomVal:           *untagInCustomVal,
		UntagInNeedTagKeyType:      *untagInNeedTagKeyType,
		UntagInNeedTagType:         *untagInNeedTagType,
		UntagInTagsElem:            *untagInTagsElem,
		UntagOp:                    *untagOp,
		UpdateTagsFunc:             *updateTagsFunc,
		UpdateTagsIgnoreSystem:     !*updateTagsNoIgnoreSystem,
		WaitForPropagation:         *waitForPropagation,
		WaitFuncComparator:         *waitFuncComparator,
		WaitTagsPropagatedFunc:     *waitTagsPropagatedFunc,
		WaitContinuousOccurence:    *waitContinuousOccurence,
		WaitDelay:                  formatDuration(*waitDelay),
		WaitMinTimeout:             formatDuration(*waitMinTimeout),
		WaitPollInterval:           formatDuration(*waitPollInterval),
		WaitTimeout:                formatDuration(*waitTimeout),

		IsDefaultListTags:   *listTagsFunc == defaultListTagsFunc,
		IsDefaultUpdateTags: *updateTagsFunc == defaultUpdateTagsFunc,
	}

	templateBody := newTemplateBody(*kvtValues)
	templateFuncMap := template.FuncMap{
		"Snake": names.ToSnakeCase,
	}
	d := g.NewGoFileDestination(filename)

	if *getTag || *listTags || *serviceTagsMap || *serviceTagsSlice || *updateTags {
		// If you intend to only generate Tags and KeyValueTags helper methods,
		// the corresponding aws-sdk-go	service package does not need to be imported
		if !*getTag && !*listTags && !*serviceTagsSlice && !*updateTags {
			templateData.AWSService = ""
			templateData.TagPackage = ""
		}

		if err := d.BufferTemplate("header", templateBody.header, templateData, templateFuncMap); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *getTag {
		if err := d.BufferTemplate("gettag", templateBody.getTag, templateData, templateFuncMap); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *listTags {
		if err := d.BufferTemplate("listtags", templateBody.listTags, templateData, templateFuncMap); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *serviceTagsMap {
		if err := d.BufferTemplate("servicetagsmap", templateBody.serviceTagsMap, templateData, templateFuncMap); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *serviceTagsSlice {
		if err := d.BufferTemplate("servicetagsslice", templateBody.serviceTagsSlice, templateData, templateFuncMap); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *updateTags {
		if err := d.BufferTemplate("updatetags", templateBody.updateTags, templateData, templateFuncMap); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if *waitForPropagation {
		if err := d.BufferTemplate("waittagspropagated", templateBody.waitTagsPropagated, templateData, templateFuncMap); err != nil {
			g.Fatalf("generating file (%s): %s", filename, err)
		}
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
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
