package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: convert <input.json> [output.json]")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := inputFile
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", inputFile, err)
	}

	var doc2 openapi2.T
	if err := json.Unmarshal(data, &doc2); err != nil {
		log.Fatalf("Failed to parse Swagger 2.0: %v", err)
	}

	doc3, err := openapi2conv.ToV3(&doc2)
	if err != nil {
		log.Fatalf("Failed to convert to OpenAPI 3.0: %v", err)
	}

	output, err := json.MarshalIndent(doc3, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		log.Fatalf("Failed to write %s: %v", outputFile, err)
	}

	log.Printf("Converted %s to OpenAPI 3.0: %s", inputFile, outputFile)
}
