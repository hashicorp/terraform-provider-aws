//go:build generate
// +build generate

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"syreclabs.com/go/faker"
)

const (
	defaultSeparator = "|"
)

var doctypes = []string{
	"DRAMA",
	"COMEDY",
}

var dramaWords = []string{
	"dramatic",
	"gripping",
	"poingnant",
}

var comedyWords = []string{
	"funny",
	"comedic",
	"hilarious",
}

func main() {
	log.SetFlags(0)

	seed := int64(1) // Default rand seed
	rand.Seed(seed)
	faker.Seed(seed)

	// documentFile, err := os.OpenFile("./test-fixtures/document_classifier_multilabel/documents.csv", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	documentFile, err := os.OpenFile("../../../test-fixtures/document_classifier_multilabel/documents.csv", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("error opening file %q: %s", "documents.csv", err)
	}
	defer closeFile(documentFile, "documents.csv")
	documentsWriter := csv.NewWriter(documentFile)

	for i := 0; i < 100; i++ {
		f := rand.Intn(2)
		var doctype string
		if f == 0 {
			doctype = doctypes[rand.Intn(len(doctypes))]
		} else {
			doctype = strings.Join(doctypes, defaultSeparator)
		}

		title := faker.Lorem().Word()

		var desc string
		if doctype == "DRAMA" {
			desc = dramaWords[rand.Intn(len(dramaWords))]
		} else if doctype == "COMEDY" {
			desc = comedyWords[rand.Intn(len(comedyWords))]
		} else {
			desc = fmt.Sprintf("%s and %s",
				dramaWords[rand.Intn(len(dramaWords))],
				comedyWords[rand.Intn(len(comedyWords))],
			)
		}

		line := fmt.Sprintf("%s is %s", title, desc)

		if err := documentsWriter.Write([]string{doctype, line}); err != nil {
			log.Fatalf("error writing to file %q: %s", "documents.csv", err)
		}
	}

	documentsWriter.Flush()
}

func closeFile(f *os.File, name string) {
	if err := f.Close(); err != nil {
		log.Fatalf("error closing file %q: %s", name, err)
	}
}
