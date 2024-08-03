package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

type responseDoc struct {
	responseCode int
	description  string
	schema       string
	ref          string
}

type param struct {
	inputType   string
	description string
	name        string
	in          string
}

type apiDoc struct {
	path       string
	call       string
	consumes   []string
	produces   []string
	tags       []string
	summary    string
	parameters []param
	responses  []responseDoc
}

func printResponse(r responseDoc) {
	fmt.Printf("    Response: %d\n", r.responseCode)
	fmt.Printf("	Description: %s\n", r.description)
	fmt.Printf("	Ref: %s\n", r.ref)
}

func printParams(p param) {
	fmt.Printf("    Parameter: %s\n", p.name)
	fmt.Printf("	Type: %s\n", p.inputType)
	fmt.Printf("	Description: %s\n", p.description)
	fmt.Printf("	In: %s\n", p.in)
}

func printAPI(api apiDoc) {
	fmt.Printf("Path: %s\n", api.path)
	fmt.Printf("  Call: %s\n", api.call)
	fmt.Printf("  Summary: %s\n", api.summary)
	fmt.Printf("  Consumes: %s\n", api.consumes)
	fmt.Printf("  Produces: %s\n", api.produces)
	fmt.Printf("  Tags: %s\n", api.tags)
	for _, p := range api.parameters {
		printParams(p)
	}
	for _, r := range api.responses {
		printResponse(r)
	}
}

func swag2(file string) []apiDoc {
	doc, err := loads.Spec(file)
	if err != nil {
		panic(err)
	}
	swagger := doc.Spec()
	apiDocList := []apiDoc{}
	if swagger.Swagger != "2.0" {
		fmt.Println("Only Swagger 2.0 is supported")
		return apiDocList
	}
	for path, pathItem := range swagger.Paths.Paths {
		apiDocPH := apiDoc{}
		for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"} {
			var operation *spec.Operation
			switch method {
			case "GET":
				operation = pathItem.Get
			case "POST":
				operation = pathItem.Post
			case "PUT":
				operation = pathItem.Put
			case "DELETE":
				operation = pathItem.Delete
			case "OPTIONS":
				operation = pathItem.Options
			case "PATCH":
				operation = pathItem.Patch
			}
			if operation != nil {
				apiDocPH.path = path
				apiDocPH.call = method
				apiDocPH.summary = operation.Summary
				apiDocPH.consumes = operation.Consumes
				apiDocPH.produces = operation.Produces
				apiDocPH.tags = operation.Tags
				for _, params := range operation.Parameters {
					if strings.Contains(params.Description, "RFC") {
						params.Type = "date"
					}
					apiDocPH.parameters = append(apiDocPH.parameters, param{params.Type, params.Description, params.Name, params.In})
				}
				responseDocPH := responseDoc{}
				for code, response := range operation.Responses.StatusCodeResponses {
					responseDocPH.responseCode = code
					responseDocPH.description = response.Description
					apiDocPH.responses = append(apiDocPH.responses, responseDocPH)
				}
				apiDocList = append(apiDocList, apiDocPH)
			}
		}
	}
	return apiDocList
}

func replacePlaceholder(urlTemplate, id string) string {
	re := regexp.MustCompile(`\{.*?\}`)
	return re.ReplaceAllString(urlTemplate, id)
}
