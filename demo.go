package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	mrand "math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	fuzz "github.com/google/gofuzz"
)

func demo(controllerAddress string, api apiDoc, token string, timer int, requiresAuth bool, headers bool, id int, timeout context.Context, printMutex *sync.Mutex, wordlist []string, pathlist []string, backoff int) error {
	f := fuzz.New()
	var fint int
	var fbool bool
	var fstring string
	var ftime time.Time
	var evoflag bool
	evolist := [][]byte{}
	evolist = append(evolist, []byte("test"))
	evolist = append(evolist, []byte("test1"))
	evolist = append(evolist, []byte("test2"))
	evolist = append(evolist, []byte("test3"))
	var evoph string
	list := [][]byte{}
	for _, word := range wordlist {
		list = append(list, []byte(word))
	}
	listparam := make(map[string]string)
	// myInt gets a random value.
	i := 0 // wordlist iterator
	j := 0 // pathlist iterator
	if len(wordlist) > 0 {
		for {
			//TIMEOUT HERE
			select {
			case <-timeout.Done():
				fmt.Printf("%d seconds have elapsed. Ending fuzzing process for %s\n", timer, api.path)
				return nil
			default:
				//POET w/ WORDLIST STARTS HERE
				winner := list[i]
				i++
				if i >= len(list) {
					i = 0
				}
				evoflag = false
				var evotarget []byte
				if len(evolist) > 0 {
					if mrand.Intn(2) == 1 {
						evotarget = evolist[mrand.Intn(len(evolist))]
						evoflag = true
					}
				}
				j++
				if j >= len(list) {
					j = 0
				}
				apiPath := api.path
				var fuzztarget string
				f.Fuzz(&fstring)
				if len(pathlist) != 0 {
					fstring = url.PathEscape(pathlist[j])
				} else {
					fstring = url.PathEscape(fstring)
				}
				apiPath = replacePlaceholder(apiPath, fstring)
				requestURL := fmt.Sprintf("%s%s", controllerAddress, apiPath)
				data := url.Values{}
				params := url.Values{}
				headerlist := url.Values{}
				var tstring string
				for _, p := range api.parameters {
					types := p.inputType
					switch types {
					case "string":
						if evoflag {
							evoph = string(mutate(evotarget))
							tstring = replacePlaceholder(string(winner), url.QueryEscape(evoph))
						} else {
							fuzz.NewFromGoFuzz(winner).Fuzz(&fuzztarget)
							tstring = replacePlaceholder(string(winner), url.QueryEscape(fuzztarget))
						}
						data.Set(p.name, tstring)
						if p.in == "query" {
							params.Set(p.name, tstring)
						}
						if p.in == "header" {
							headerlist.Set(p.name, tstring)
						}
						listparam[p.name] = tstring
					case "integer":
						fuzz.NewFromGoFuzz(winner).Fuzz(&fint)
						if p.in == "body" {
							data.Set(p.name, strconv.Itoa(fint))
						}
						if p.in == "query" {
							params.Set(p.name, strconv.Itoa(fint))
						}
						if p.in == "header" {
							headerlist.Set(p.name, strconv.Itoa(fint))
						}
						listparam[p.name] = strconv.Itoa(fint)
					case "boolean":
						fbool = mrand.Intn(2) == 1
						//list = append(list, []byte(strconv.FormatBool(fbool)))
						if p.in == "body" {
							data.Set(p.name, strconv.FormatBool(fbool))
						}
						if p.in == "query" {
							params.Set(p.name, strconv.FormatBool(fbool))
						}
						if p.in == "header" {
							headerlist.Set(p.name, strconv.FormatBool(fbool))
						}
						listparam[p.name] = strconv.FormatBool(fbool)
					case "date":
						fuzz.NewFromGoFuzz(winner).Fuzz(&ftime)
						if p.in == "body" {
							data.Set(p.name, ftime.Format(time.RFC1123))
						}
						if p.in == "query" {
							params.Set(p.name, ftime.Format(time.RFC1123))
						}
						if p.in == "header" {
							headerlist.Set(p.name, tstring)
						}
						listparam[p.name] = ftime.Format(time.RFC1123)
					}
				}
				printMutex.Lock()
				if evoflag {
					fmt.Println("\nEvolving: ", string(evotarget), " -> ", evoph, " -> ", tstring)
				}
				fmt.Printf("Path generated %s\n", requestURL)
				for key, value := range listparam {
					fmt.Printf("%s -> %s\n", key, value)
				}
				if len([]byte(evoph)) != 0 {
					evolist = append(evolist, []byte(evoph))
				}
				fmt.Print("\n")
				printMutex.Unlock()
			}
		}
	} else {
		select {
		case <-timeout.Done():
			fmt.Printf("%d seconds have elapsed. Ending fuzzing process for %s\n", timer, api.path)
			return nil
		default:
			//POET w/ WORDLIST STARTS HERE
			winner := list[i]
			i++
			if i >= len(list) {
				i = 0
			}
			evoflag = false
			var evotarget []byte
			if len(evolist) > 0 {
				if mrand.Intn(2) == 1 {
					evotarget = evolist[mrand.Intn(len(evolist))]
					evoflag = true
				}
			}
			apiPath := api.path
			printMutex.Lock()
			fmt.Println("Method: ", api.call)
			printMutex.Unlock()
			var fuzztarget string
			f.Fuzz(&fstring)
			if len(pathlist) != 0 {
				fstring = url.PathEscape(pathlist[j])
			} else {
				fstring = url.PathEscape(fstring)
			}
			apiPath = replacePlaceholder(apiPath, fstring)
			requestURL := fmt.Sprintf("%s%s", controllerAddress, apiPath)
			data := url.Values{}
			params := url.Values{}
			headerlist := url.Values{}
			var tstring string
			datatype := "none"
			for _, consume := range api.consumes {
				if consume == "application/json" {
					datatype = "json"
					break
				}
				if consume == "application/x-www-form-urlencoded" {
					datatype = "form"
					break
				}
				if consume == "application/vnd.api+json" {
					datatype = "jsonvnd"
					break
				}
			}
			for _, p := range api.parameters {
				types := p.inputType
				switch types {
				case "string":
					if evoflag {
						evoph := string(mutate(evotarget))
						tstring = replacePlaceholder(evoph, url.QueryEscape(evoph))
						//fmt.Println("Evolving: ", evoph, " -> ", tstring)
					} else {
						fuzz.NewFromGoFuzz(winner).Fuzz(&fuzztarget)
						tstring = replacePlaceholder(string(winner), url.QueryEscape(fuzztarget))
					}
					data.Set(p.name, tstring)
					if p.in == "query" {
						params.Set(p.name, tstring)
					}
					if p.in == "header" {
						headerlist.Set(p.name, tstring)
					}
					listparam[p.name] = tstring
				case "integer":
					if evoflag {
						fint, _ = strconv.Atoi(string(mutate(evotarget)))
					} else {
						fuzz.NewFromGoFuzz(winner).Fuzz(&fint)
					}
					if p.in == "body" {
						data.Set(p.name, strconv.Itoa(fint))
					}
					if p.in == "query" {
						params.Set(p.name, strconv.Itoa(fint))
					}
					if p.in == "header" {
						headerlist.Set(p.name, strconv.Itoa(fint))
					}
					listparam[p.name] = strconv.Itoa(fint)
				case "boolean":
					fbool = mrand.Intn(2) == 1
					//list = append(list, []byte(strconv.FormatBool(fbool)))
					if p.in == "body" {
						data.Set(p.name, strconv.FormatBool(fbool))
					}
					if p.in == "query" {
						params.Set(p.name, strconv.FormatBool(fbool))
					}
					if p.in == "header" {
						headerlist.Set(p.name, strconv.FormatBool(fbool))
					}
					listparam[p.name] = strconv.FormatBool(fbool)
				case "date":
					if evoflag {
						ftime, _ = time.Parse(time.RFC1123, string(mutate(evotarget)))
					} else {
						fuzz.NewFromGoFuzz(winner).Fuzz(&ftime)
					}
					if p.in == "body" {
						data.Set(p.name, ftime.Format(time.RFC1123))
					}
					if p.in == "query" {
						params.Set(p.name, ftime.Format(time.RFC1123))
					}
					if p.in == "header" {
						headerlist.Set(p.name, tstring)
					}
					listparam[p.name] = ftime.Format(time.RFC1123)
				}
			}
			req := &http.Request{}
			reqUrl := fmt.Sprintf("%s?%s", requestURL, params.Encode())
			if datatype == "json" {
				dataMap := make(map[string]interface{})
				for key, value := range data {
					if len(value) > 0 {
						dataMap[key] = value[0]
					}
				}
				jsondata, err := json.Marshal(dataMap)
				if err != nil {
					fmt.Printf("Error marshalling data: %s\n", err)
					return err
				}
				if len(jsondata) != 2 {
					req, err = http.NewRequest(api.call, reqUrl, bytes.NewBuffer(jsondata))
				} else {
					req, err = http.NewRequest(api.call, reqUrl, nil)
				}
				if err != nil {
					fmt.Printf("Error creating request: %s\n", err)
					return err
				}
				req.Header.Set("Content-Type", "application/json")
			} else if datatype == "jsonvnd" {
				dataMap := make(map[string]interface{})
				for key, value := range data {
					if len(value) > 1 {
						dataMap[key] = value[0]
					}
				}
				jsondata, err := json.Marshal(data)
				if err != nil {
					fmt.Printf("Error marshalling data: %s\n", err)
					return err
				}
				//print jsondata
				req, err = http.NewRequest(api.call, requestURL, bytes.NewBuffer(jsondata))
				if err != nil {
					fmt.Printf("Error creating request: %s\n", err)
					return err
				}
				req.Header.Set("Content-Type", "application/vnd.api+json")
			} else if datatype == "form" {
				req, _ = http.NewRequest(api.call, requestURL, strings.NewReader(data.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			for key, value := range headerlist {
				req.Header.Add(key, value[0]) //forced header fuzzing
			}
			if headers {
				fuzz.NewFromGoFuzz(winner).Fuzz(&fint)
				req.Header.Add("Age", strconv.Itoa(fint))
				listparam["Age"] = strconv.Itoa(fint)
				fuzz.NewFromGoFuzz(winner).Fuzz(&ftime)
				req.Header.Add("Date", ftime.Format(time.RFC1123))
				listparam["Date"] = ftime.Format(time.RFC1123)
				fuzz.NewFromGoFuzz(winner).Fuzz(&ftime)
				req.Header.Add("Expires", ftime.Format(time.RFC1123))
				listparam["Expires"] = ftime.Format(time.RFC1123)
				fuzz.NewFromGoFuzz(winner).Fuzz(&ftime)
				req.Header.Add("Last-Modified", ftime.Format(time.RFC1123))
				listparam["Last-Modified"] = ftime.Format(time.RFC1123)
			}
			if requiresAuth {
				req.Header.Add("Authorization", "Bearer "+token)
			} else {
				f.Fuzz(&fstring)
				req.Header.Add("Authorization", fstring)
				fuzz.NewFromGoFuzz(winner).Fuzz(&fuzztarget)
				listparam["Authorization"] = fuzztarget
			}
			printMutex.Lock()
			if evoflag {
				fmt.Println("\nEvolving: ", string(evotarget), " -> ", evoph, " -> ", tstring)
			}
			fmt.Printf("Path generated %s\n", requestURL)
			for key, value := range listparam {
				fmt.Printf("%s -> %s\n", key, value)
			}
			evolist = append(evolist, []byte(evoph))
			fmt.Print("\n")
			printMutex.Unlock()
		}
	}
	return nil
}
