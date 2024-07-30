package main

import (
	"github.com/getkin/kin-openapi/openapi3"
)

func swag3(file string) []apiDoc {
	loader := openapi3.NewLoader()
	//open the file
	//fs, err := os.Open(file)
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(file)
	if err != nil {
		panic(err)
	}
	if err = doc.Validate(loader.Context); err != nil {
		panic(err)
	}
	/*paths := doc.Paths


	for path, pathItem := range paths {
		fmt.Println("Path: ", path)
		fmt.Println("PathItem: ", pathItem)
	}
	//complist := []openapi3.Components{}

	for _, path := range doc.Paths {
		for _, method := range path.Operations() {
			fmt.Println("Path: ", path.Path)
			fmt.Println("Method: ", method)
			fmt.Println("Summary: ", method.OperationID)
			fmt.Println("Consumes: ", method.RequestBody)
		}
	}*/

	/*for _, schema := range swagger.Definitions {
		fmt.Printf("Ref: %s\n", schema.Ref.String())
	}*/
	apiDocList := []apiDoc{}
	return apiDocList
}
