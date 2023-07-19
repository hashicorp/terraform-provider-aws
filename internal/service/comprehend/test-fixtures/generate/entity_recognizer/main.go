// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"syreclabs.com/go/faker"
)

func main() {
	var entities = []string{
		"ENGINEER",
		"MANAGER",
	}

	var sentences = []string{
		"%[1]s is a %[2]s in the high tech industry.",
		"%[1]s has been a %[2]s for 14 years.",
		"Our latest new employee, %[1]s, has been a %[2]s in the industry for 4 years.",
		"Announcing %[2]s %[1]s.",
		"Help me welcome our newest %[2]s, %[1]s.",
		"%[1]s is retiring as a %[2]s.",
		"%[1]s has been an %[2]s for over a decade.",
		"%[1]s is a %[2]s with Example Corp.",
		"%[1]s will be the new %[2]s for the team.",
		"%[1]s, an %[2]s, will be presenting the award.",
		"%[1]s joins us as an %[2]s on the Example project.",
		"%[1]s is a %[2]s.",
	}

	log.SetFlags(0)

	seed := int64(1) // Default rand seed
	rand.Seed(seed)
	faker.Seed(seed)

	entitiesFile, err := os.OpenFile("./test-fixtures/entity_recognizer/entitylist.csv", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("error opening file %q: %s", "entitylist.csv", err)
	}
	defer closeFile(entitiesFile, "entitylist.csv")

	if _, err := fmt.Fprintln(entitiesFile, "Text,Type"); err != nil {
		log.Fatalf("error writing to file %q: %s", "entitylist.csv", err)
	}

	documentFile, err := os.OpenFile("./test-fixtures/entity_recognizer/documents.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("error opening file %q: %s", "documents.txt", err)
	}
	defer closeFile(documentFile, "documents.txt")

	annotationsFile, err := os.OpenFile("./test-fixtures/entity_recognizer/annotations.csv", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("error opening file %q: %s", "annotations.csv", err)
	}
	defer closeFile(annotationsFile, "annotations.csv")
	annotationsWriter := csv.NewWriter(annotationsFile)
	if err := annotationsWriter.Write([]string{"File", "Line", "Begin Offset", "End Offset", "Type"}); err != nil {
		log.Fatalf("error writing to file %q: %s", "annotations.csv", err)
	}

	for i := 0; i < 1000; i++ {
		name := faker.Name().Name()
		entity := entities[rand.Intn(len(entities))]

		if _, err := fmt.Fprintf(entitiesFile, "%s,%s\n", name, entity); err != nil {
			log.Fatalf("error writing to file %q: %s", "entitylist.csv", err)
		}

		sentence := sentences[rand.Intn(len(sentences))]
		line := fmt.Sprintf(sentence, name, strings.ToLower(entity))
		if _, err := fmt.Fprintln(documentFile, line); err != nil {
			log.Fatalf("error writing to file %q: %s", "documents.txt", err)
		}

		offset := strings.Index(line, name)
		end := offset + len(name)
		if err := annotationsWriter.Write([]string{"documents.txt", strconv.Itoa(i), strconv.Itoa(offset), strconv.Itoa(end), entity}); err != nil {
			log.Fatalf("error writing to file %q: %s", "annotations.csv", err)
		}
	}

	annotationsWriter.Flush()
}

func closeFile(f *os.File, name string) {
	if err := f.Close(); err != nil {
		log.Fatalf("error closing file %q: %s", name, err)
	}
}
