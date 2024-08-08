package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func swag3(file string) []apiDoc {
	loader := openapi3.NewLoader()
	//open the file
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(file)
	if err != nil {
		panic(err)
	}
	if err = doc.Validate(loader.Context); err != nil {
		panic(err)
	}

	apiDocList := []apiDoc{}
	for pathstr, pathit := range doc.Paths.Map() {

		for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"} {
			apiDocph := apiDoc{}
			apiDocph.path = pathstr
			switch method {
			case "GET":
				if pathit.Get != nil {
					operation := pathit.Get
					apiDocph.call = method
					for _, params := range operation.Parameters {
						params := doc.Components.Parameters[strings.Split(params.Ref, "/")[3]]
						param := param{(*params.Value.Schema.Value.Type)[0], params.Value.Description, params.Value.Name, params.Value.In}
						apiDocph.parameters = append(apiDocph.parameters, param)
					}
					if operation.RequestBody != nil {
						for mediastr, media := range operation.RequestBody.Value.Content {
							apiDocph.consumes = append(apiDocph.consumes, mediastr)
							parmph := param{}
							parmph.in = "body"
							for schemastr, schema := range media.Schema.Value.Properties {
								parmph.name = schemastr
								if schema.Ref != "" {
									params := doc.Components.Schemas[strings.Split(schema.Ref, "/")[3]]
									if (params.Value.Type) != nil {
										parmph.inputType = (*params.Value.Type)[0]
									}
									if params.Value.Description != "" {
										parmph.description = params.Value.Description
									}
								} else {
									for _, prop := range media.Schema.Value.Properties {
										if (prop.Value.Type) != nil {
											parmph.inputType = (*prop.Value.Type)[0]
										}
										if prop.Value.Format != "" {
											parmph.description = prop.Value.Format
										}
									}
								}
								apiDocph.parameters = append(apiDocph.parameters, parmph)
							}
						}
					}
					for respstr, resp := range operation.Responses.Map() {
						if respstr != "default" {
							respInt, err := strconv.Atoi(respstr)
							if err != nil {
								fmt.Println("Error converting response code to int:", err)
							} else {
								resp1 := responseDoc{respInt, *resp.Value.Description, "", ""}
								apiDocph.responses = append(apiDocph.responses, resp1)
							}
						}
					}
				}
			case "POST":
				if pathit.Post != nil {
					operation := pathit.Post
					apiDocph.call = method
					for _, params := range operation.Parameters {
						params := doc.Components.Parameters[strings.Split(params.Ref, "/")[3]]
						param := param{(*params.Value.Schema.Value.Type)[0], params.Value.Description, params.Value.Name, params.Value.In}
						apiDocph.parameters = append(apiDocph.parameters, param)
					}
					if operation.RequestBody != nil {
						for mediastr, media := range operation.RequestBody.Value.Content {
							apiDocph.consumes = append(apiDocph.consumes, mediastr)
							parmph := param{}
							parmph.in = "body"
							for schemastr, schema := range media.Schema.Value.Properties {
								parmph.name = schemastr
								if schema.Ref != "" {
									params := doc.Components.Schemas[strings.Split(schema.Ref, "/")[3]]
									if (params.Value.Type) != nil {
										parmph.inputType = (*params.Value.Type)[0]
									}
									if params.Value.Description != "" {
										parmph.description = params.Value.Description
									}
								} else {
									for _, prop := range media.Schema.Value.Properties {
										if (prop.Value.Type) != nil {
											parmph.inputType = (*prop.Value.Type)[0]
										}
										if prop.Value.Format != "" {
											parmph.description = prop.Value.Format
										}
									}
								}
								apiDocph.parameters = append(apiDocph.parameters, parmph)
							}
						}
					}
					for respstr, resp := range operation.Responses.Map() {
						if respstr != "default" {
							respInt, err := strconv.Atoi(respstr)
							if err != nil {
								fmt.Println("Error converting response code to int:", err)
							} else {
								resp1 := responseDoc{respInt, *resp.Value.Description, "", ""}
								apiDocph.responses = append(apiDocph.responses, resp1)
							}
						}
					}
				}
			case "PUT":
				if pathit.Put != nil {
					operation := pathit.Put
					apiDocph.call = method
					for _, params := range operation.Parameters {
						if params.Ref != "" {
							params := doc.Components.Parameters[strings.Split(params.Ref, "/")[3]]
							param := param{(*params.Value.Schema.Value.Type)[0], params.Value.Description, params.Value.Name, params.Value.In}
							apiDocph.parameters = append(apiDocph.parameters, param)
						} else {
							param := param{(*params.Value.Schema.Value.Type)[0], params.Value.Description, params.Value.Name, params.Value.In}
							apiDocph.parameters = append(apiDocph.parameters, param)
						}
					}
					if operation.RequestBody != nil {
						for mediastr, media := range operation.RequestBody.Value.Content {
							apiDocph.consumes = append(apiDocph.consumes, mediastr)
							parmph := param{}
							parmph.in = "body"
							if media.Schema.Ref != "" {
								parms := doc.Components.Schemas[strings.Split(media.Schema.Ref, "/")[3]]
								for _, params := range parms.Value.Properties {
									if (params.Value.Type) != nil {
										parmph.inputType = (*params.Value.Type)[0]
									}
									if params.Value.Description != "" {
										parmph.description = params.Value.Description
									}
									apiDocph.parameters = append(apiDocph.parameters, parmph)
								}
							} else {
								for _, prop := range media.Schema.Value.Properties {
									if (prop.Value.Type) != nil {
										parmph.inputType = (*prop.Value.Type)[0]
									}
									if prop.Value.Format != "" {
										parmph.description = prop.Value.Format
									}
								}
								apiDocph.parameters = append(apiDocph.parameters, parmph)
							}
						}
					}
					for respstr, resp := range operation.Responses.Map() {
						if respstr != "default" {
							respInt, err := strconv.Atoi(respstr)
							if err != nil {
								fmt.Println("Error converting response code to int:", err)
							} else {
								resp1 := responseDoc{respInt, *resp.Value.Description, "", ""}
								apiDocph.responses = append(apiDocph.responses, resp1)
							}
						}
					}
				}
			case "DELETE":
				if pathit.Delete != nil {
					operation := pathit.Delete
					apiDocph.call = method
					for _, params := range operation.Parameters {
						params := doc.Components.Parameters[strings.Split(params.Ref, "/")[3]]
						param := param{(*params.Value.Schema.Value.Type)[0], params.Value.Description, params.Value.Name, params.Value.In}
						apiDocph.parameters = append(apiDocph.parameters, param)
					}
					if operation.RequestBody != nil {
						for mediastr, media := range operation.RequestBody.Value.Content {
							apiDocph.consumes = append(apiDocph.consumes, mediastr)
							parmph := param{}
							parmph.in = "body"
							for schemastr, schema := range media.Schema.Value.Properties {
								parmph.name = schemastr
								if schema.Ref != "" {
									params := doc.Components.Schemas[strings.Split(schema.Ref, "/")[3]]
									if (params.Value.Type) != nil {
										parmph.inputType = (*params.Value.Type)[0]
									}
									if params.Value.Description != "" {
										parmph.description = params.Value.Description
									}
								} else {
									for _, prop := range media.Schema.Value.Properties {
										if (prop.Value.Type) != nil {
											parmph.inputType = (*prop.Value.Type)[0]
										}
										if prop.Value.Format != "" {
											parmph.description = prop.Value.Format
										}
									}
								}
								apiDocph.parameters = append(apiDocph.parameters, parmph)
							}
						}
					}
					for respstr, resp := range operation.Responses.Map() {
						if respstr != "default" {
							respInt, err := strconv.Atoi(respstr)
							if err != nil {
								fmt.Println("Error converting response code to int:", err)
							} else {
								resp1 := responseDoc{respInt, *resp.Value.Description, "", ""}
								apiDocph.responses = append(apiDocph.responses, resp1)
							}
						}
					}
				}
			case "OPTIONS":
				if pathit.Options != nil {
					operation := pathit.Options
					apiDocph.call = method
					for _, params := range operation.Parameters {
						params := doc.Components.Parameters[strings.Split(params.Ref, "/")[3]]
						param := param{(*params.Value.Schema.Value.Type)[0], params.Value.Description, params.Value.Name, params.Value.In}
						apiDocph.parameters = append(apiDocph.parameters, param)
					}
					if operation.RequestBody != nil {
						for mediastr, media := range operation.RequestBody.Value.Content {
							apiDocph.consumes = append(apiDocph.consumes, mediastr)
							parmph := param{}
							parmph.in = "body"
							for schemastr, schema := range media.Schema.Value.Properties {
								parmph.name = schemastr
								if schema.Ref != "" {
									params := doc.Components.Schemas[strings.Split(schema.Ref, "/")[3]]
									if (params.Value.Type) != nil {
										parmph.inputType = (*params.Value.Type)[0]
									}
									if params.Value.Description != "" {
										parmph.description = params.Value.Description
									}
								} else {
									for _, prop := range media.Schema.Value.Properties {
										if (prop.Value.Type) != nil {
											parmph.inputType = (*prop.Value.Type)[0]
										}
										if prop.Value.Format != "" {
											parmph.description = prop.Value.Format
										}
									}
								}
								apiDocph.parameters = append(apiDocph.parameters, parmph)
							}
						}
					}
					for respstr, resp := range operation.Responses.Map() {
						respInt, err := strconv.Atoi(respstr)
						if err != nil {
							fmt.Println("Error converting response code to int:", err)
						} else {
							resp1 := responseDoc{respInt, *resp.Value.Description, "", ""}
							apiDocph.responses = append(apiDocph.responses, resp1)
						}
					}
				}
			case "PATCH":
				if pathit.Patch != nil {
					operation := pathit.Patch
					apiDocph.call = method
					for _, params := range operation.Parameters {
						params := doc.Components.Parameters[strings.Split(params.Ref, "/")[3]]
						param := param{(*params.Value.Schema.Value.Type)[0], params.Value.Description, params.Value.Name, params.Value.In}
						apiDocph.parameters = append(apiDocph.parameters, param)
					}
					if operation.RequestBody != nil {
						for mediastr, media := range operation.RequestBody.Value.Content {
							apiDocph.consumes = append(apiDocph.consumes, mediastr)
							parmph := param{}
							parmph.in = "body"
							for schemastr, schema := range media.Schema.Value.Properties {
								parmph.name = schemastr
								if schema.Ref != "" {
									params := doc.Components.Schemas[strings.Split(schema.Ref, "/")[3]]
									if (params.Value.Type) != nil {
										parmph.inputType = (*params.Value.Type)[0]
									}
									if params.Value.Description != "" {
										parmph.description = params.Value.Description
									}
								} else {
									for _, prop := range media.Schema.Value.Properties {
										if (prop.Value.Type) != nil {
											parmph.inputType = (*prop.Value.Type)[0]
										}
										if prop.Value.Format != "" {
											parmph.description = prop.Value.Format
										}
									}
								}
								apiDocph.parameters = append(apiDocph.parameters, parmph)
							}
						}
					}
					for respstr, resp := range operation.Responses.Map() {
						if respstr != "default" {
							respInt, err := strconv.Atoi(respstr)
							if err != nil {
								fmt.Println("Error converting response code to int:", err)
							} else {
								resp1 := responseDoc{respInt, *resp.Value.Description, "", ""}
								apiDocph.responses = append(apiDocph.responses, resp1)
							}
						}
					}
				}
			}
			if len(apiDocph.parameters) != 0 && len(apiDocph.responses) != 0 {
				apiDocList = append(apiDocList, apiDocph)
			}
		}
	}

	return apiDocList
}
