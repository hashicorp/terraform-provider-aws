//go:build generate
// +build generate

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"

	"syreclabs.com/go/faker"
)

var doctypes = []string{
	"PHISHING",
	"SPAM",
}

var phishingDocs = []string{
	`Dear %[1]s,\n\nYour transaction %[2]s has failed.\n\nCall %[3]s for help.\n`,
	`%[1]s,\n\nYour order number %[2]s has been returned.\n\nCall %[3]s to get help.\n`,
	`Hello %[1]s,\n\nCall %[3]s for help with your order %[2]s. Otherwise it will be returned to the sender.\n`,
}

var spamDocs = []string{
	`Dear %[1]s,\n\nBuy a %[2]s from %[3]s now!\n`,
	`%[1]s,\n\nDon't miss out on buying %[3]s's %[2]s today!\n`,
	`Hello %[1]s,\n\nNow available!\n\nA %[2]s from %[3]s\n`,
}

func main() {
	log.SetFlags(0)

	seed := int64(1) // Default rand seed
	rand.Seed(seed)
	faker.Seed(seed)

	documentFile, err := os.OpenFile("./test-fixtures/document_classifier/documents.csv", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("error opening file %q: %s", "documents.csv", err)
	}
	defer closeFile(documentFile, "documents.csv")
	documentsWriter := csv.NewWriter(documentFile)

	for i := 0; i < 100; i++ {
		name := faker.Name().Name()
		doctype := doctypes[rand.Intn(len(doctypes))]

		var line string
		if doctype == "PHISHING" {
			order := faker.RandomString(10)
			phone := faker.PhoneNumber().PhoneNumber()
			doc := phishingDocs[rand.Intn(len(phishingDocs))]
			line = fmt.Sprintf(doc, name, order, phone)
		} else {
			doc := spamDocs[rand.Intn(len(spamDocs))]
			product := faker.Commerce().ProductName()
			company := faker.Company().Name()
			line = fmt.Sprintf(doc, name, product, company)
		}

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
