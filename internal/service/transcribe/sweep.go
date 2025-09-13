// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transcribe

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_transcribe_language_model", &resource.Sweeper{
		Name: "aws_transcribe_language_model",
		F:    sweepLanguageModels,
		Dependencies: []string{
			"aws_s3_bucket",
			"aws_iam_role",
		},
	})

	resource.AddTestSweepers("aws_transcribe_medical_vocabulary", &resource.Sweeper{
		Name: "aws_transcribe_medical_vocabulary",
		F:    sweepMedicalVocabularies,
		Dependencies: []string{
			"aws_s3_bucket",
		},
	})

	resource.AddTestSweepers("aws_transcribe_vocabulary", &resource.Sweeper{
		Name: "aws_transcribe_vocabulary",
		F:    sweepVocabularies,
		Dependencies: []string{
			"aws_s3_bucket",
		},
	})

	resource.AddTestSweepers("aws_transcribe_vocabulary_filter", &resource.Sweeper{
		Name: "aws_transcribe_vocabulary_filter",
		F:    sweepVocabularyFilters,
		Dependencies: []string{
			"aws_s3_bucket",
		},
	})
}

func sweepLanguageModels(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.TranscribeClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &transcribe.ListLanguageModelsInput{}

	pages := transcribe.NewListLanguageModelsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Transcribe Language Models sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving Transcribe Language Models: %w", err)
		}

		for _, model := range page.Models {
			name := aws.ToString(model.ModelName)
			log.Printf("[INFO] Deleting Transcribe Language Model: %s", name)

			r := ResourceLanguageModel()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Transcribe Language Models for %s: %w", region, err)
	}

	return nil
}

func sweepMedicalVocabularies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.TranscribeClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &transcribe.ListMedicalVocabulariesInput{}

	for {
		out, err := conn.ListMedicalVocabularies(ctx, in)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Transcribe Medical Vocabularies sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Transcribe Medical Vocabularies: %w", err)
		}

		for _, vocab := range out.Vocabularies {
			name := aws.ToString(vocab.VocabularyName)
			log.Printf("[INFO] Deleting Transcribe Medical Vocabularies: %s", name)

			r := ResourceMedicalVocabulary()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Transcribe Medical Vocabularies for %s: %w", region, err)
	}

	return nil
}

func sweepVocabularies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.TranscribeClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &transcribe.ListVocabulariesInput{}

	for {
		out, err := conn.ListVocabularies(ctx, in)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Transcribe Vocabularies sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Transcribe Vocabularies: %w", err)
		}

		for _, vocab := range out.Vocabularies {
			name := aws.ToString(vocab.VocabularyName)
			log.Printf("[INFO] Deleting Transcribe Vocabularies: %s", name)

			r := ResourceVocabulary()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Transcribe Vocabularies for %s: %w", region, err)
	}

	return nil
}

func sweepVocabularyFilters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.TranscribeClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &transcribe.ListVocabularyFiltersInput{}

	for {
		out, err := conn.ListVocabularyFilters(ctx, in)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Transcribe Vocabulary Filter sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Transcribe Vocabulary Filters: %w", err)
		}

		for _, filter := range out.VocabularyFilters {
			name := aws.ToString(filter.VocabularyFilterName)
			log.Printf("[INFO] Deleting Transcribe Vocabulary Filter: %s", name)

			r := ResourceVocabularyFilter()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Transcribe Vocabulary Filters for %s: %w", region, err)
	}

	return nil
}
