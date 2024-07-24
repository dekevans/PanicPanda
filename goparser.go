package main

import (
	"fmt"
	"regexp"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

type responseDoc struct {
	responseCode int    `json:"response"`
	description  string `json:"description"`
	schema       string `json:"schema"` // TODO: ADD MORE SCHEMA TYPES
	ref          string `json:"ref"`
}

type ref struct {
	ref string `json:"ref"`
} //TODO: ADD FUNCTIONALITY

type param struct {
	inputType   string `json:"type"`
	description string `json:"description"`
	name        string `json:"name"`
	in          string `json:"in"`
}

type components struct {
	parameters []param `json:"parameters"`
}

type apiDoc struct {
	path       string        `json:"path"`
	call       string        `json:"call"`
	consumes   []string      `json:"consumes"`
	produces   []string      `json:"produces"`
	tags       []string      `json:"tags"`
	summary    string        `json:"summary"`
	parameters []param       `json:"parameters"`
	responses  []responseDoc `json:"responses"`
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

func parseSwag(file string) []apiDoc {
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
	/*for _, schema := range swagger.Definitions {
		fmt.Printf("Ref: %s\n", schema.Ref.String())
	}*/

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
				//fmt.Printf("Path: %s\n", path)
				apiDocPH.call = method
				//fmt.Printf("  Call: %s\n", method)
				apiDocPH.summary = operation.Summary
				//fmt.Printf("  Summary: %s\n", operation.Summary)
				apiDocPH.consumes = operation.Consumes
				//fmt.Printf("  Consumes: %s\n", operation.Consumes)
				apiDocPH.produces = operation.Produces
				//fmt.Printf("  Produces: %s\n", operation.Produces)
				apiDocPH.tags = operation.Tags
				//fmt.Printf("  Tags: %s\n", operation.Tags)
				for _, params := range operation.Parameters {
					if params.Ref.String() != "" {
						//fmt.Printf("	Ref: %s\n", params.Ref.String())
					} else {
						apiDocPH.parameters = append(apiDocPH.parameters, param{params.Type, params.Description, params.Name, params.In})
						//printParams(param{params.Type, params.Description, params.Name, params.In})
					}
				}
				responseDocPH := responseDoc{}
				for code, response := range operation.Responses.StatusCodeResponses {
					responseDocPH.responseCode = code
					//fmt.Printf("    Response: %d\n", code)
					responseDocPH.description = response.Description
					//fmt.Printf("	Description: %s\n", response.Description)
					//responseDocPH.ref = response.Schema.Ref.String()
					//fmt.Printf("	Ref: %s\n", response.Schema.Ref.String())
					apiDocPH.responses = append(apiDocPH.responses, responseDocPH)
				}
				apiDocList = append(apiDocList, apiDocPH)
				//printAPI(apiDocPH)
				//printResponse(responseDocPH)
			}
		}
	}
	return apiDocList
}

func replacePlaceholder(urlTemplate, id string) string {
	re := regexp.MustCompile(`\{.*?\}`)
	return re.ReplaceAllString(urlTemplate, id)
}

func refResolver(ref string) string {
	re := regexp.MustCompile(`\{.*?\}`)
	return re.FindString(ref) //TODO: ADD FUNCTIONALITY
}
