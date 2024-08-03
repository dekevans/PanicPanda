package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	fuzz "github.com/google/gofuzz"
)

func fullfunc(controllerAddress string, api apiDoc, token string, timer int, requiresAuth bool, headers bool, id int, timeout context.Context, printMutex *sync.Mutex, wordlist []string, pathlist []string, backoff int) error {
	dbg := false
	f := fuzz.New()
	var fint int
	var fbool bool
	var fstring string
	var ftime time.Time
	list := [][]byte{}
	for _, word := range wordlist {
		list = append(list, []byte(word))
	}
	failureCount := 0
	listparam := make(map[string]string)
	// myInt gets a random value.
	avglen := -1
	i := 0
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
				apiPath := api.path
				var fuzztarget string
				f.Fuzz(&fstring)
				if len(pathlist) != 0 {
					fstring = url.PathEscape(pathlist[rand.Intn(len(pathlist))])
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
				if datatype == "none" {
					fmt.Printf("No valid content type found for %s\n", requestURL)
					return nil
				}
				for _, p := range api.parameters {
					types := p.inputType
					switch types {
					case "string":
						fuzz.NewFromGoFuzz(winner).Fuzz(&fuzztarget)
						tstring = replacePlaceholder(string(winner), url.QueryEscape(fuzztarget))
						if p.in == "body" {
							data.Set(p.name, tstring)
						}
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
							headerlist.Set(p.name, tstring)
						}
						listparam[p.name] = strconv.Itoa(fint)
					case "boolean":
						fuzz.NewFromGoFuzz(winner).Fuzz(&fbool)
						//list = append(list, []byte(strconv.FormatBool(fbool)))
						if p.in == "body" {
							data.Set(p.name, strconv.FormatBool(fbool))
						}
						if p.in == "query" {
							params.Set(p.name, strconv.FormatBool(fbool))
						}
						if p.in == "header" {
							headerlist.Set(p.name, tstring)
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
				//COURIER STARTS HERE
				client := &http.Client{Timeout: time.Second * 5}
				startTime := time.Now()
				resp, err := client.Do(req)
				timeElapsed := time.Since(startTime).Milliseconds()
				if err != nil {
					if err, ok := err.(net.Error); ok && err.Timeout() {
						fmt.Printf("Timeout: Caused by: %s\n", req.URL)
						failureCount++
						if failureCount > 10 {
							fmt.Printf("Too many failures. Waiting %d seconds before continuing\n", backoff)
							time.Sleep(time.Duration(backoff) * time.Second)
							failureCount = 0
						}
						continue
					}
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode == 429 {
					fmt.Printf("Rate limited: Backing off for %d seconds\n", backoff)
					time.Sleep(time.Duration(backoff) * time.Second)
					continue
				}
				//ORACLE STARTS HERE
				printMutex.Lock()
				if avglen == -1 {
					if avglen == -1 {
						fmt.Println("There is no content length for this response")
					} else {
						avglen = int(resp.ContentLength)
						fmt.Printf("Base length of response for %s is: %d\n", requestURL, avglen)
					}

				}
				if resp.ContentLength != int64(avglen) && avglen != -1 && int(resp.ContentLength) != -1 {
					fmt.Printf("Response length mismatch: Caused by: %s\n", requestURL)
					reqDump, err := httputil.DumpRequestOut(req, true)
					if err != nil {
						fmt.Printf("Error dumping request: %s\n", err)
						return err
					}
					fmt.Printf("HTTP Request: %s\n", string(reqDump))

				}
				flag := true
				for _, responseDoc := range api.responses {
					if responseDoc.responseCode == resp.StatusCode {
						flag = false
						break
					}
				}
				fmt.Println("Response Status: ", resp.StatusCode)
				if dbg {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Printf("Error reading response body: %s\n", err)
						return err
					}
					fmt.Printf("Response Status: %s %s\n Responded in: %d milliseconds\n", requestURL, resp.Status, timeElapsed)
					fmt.Printf("Response Body: %s\n", string(bodyBytes))
				}
				if resp.StatusCode == 500 || flag {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Printf("Error reading response body: %s\n", err)
						return err
					}
					fmt.Printf("Interesting code %d: Caused by: %s\nReponse Body: %v\n", resp.StatusCode, requestURL, string(bodyBytes))
					fmt.Print("\nParameters:\n")
					for key, value := range listparam {
						fmt.Printf("%s -> %s\n", key, value)
						list = append(list, []byte(value))
					}
					//added failure backoff
					fmt.Printf("\n")
					failureCount++
					if failureCount > 10 {
						fmt.Printf("Too many failures. Waiting %d seconds before continuing\n", backoff)
						time.Sleep(time.Duration(backoff) * time.Second)
						failureCount = 0
					}
					if headers {
						fmt.Printf("\nResponse Status: %s %s\n Responded in: %d milliseconds\n", requestURL, resp.Status, timeElapsed)
					}
					printMutex.Unlock()
				}
				i++
				if i >= len(list) {
					i = 0
				}
			}
		}
	} else {
		for {
			//TIMEOUT HERE
			select {
			case <-timeout.Done():
				fmt.Printf("%d seconds have elapsed. Ending fuzzing process for %s\n", timer, api.path)
				//printMutex.Unlock()
				return nil
			default:
				//POET w/ WORDLIST STARTS HERE
				apiPath := api.path
				fuzz := fuzz.New()
				var fuzztarget string
				fuzz.Fuzz(&fuzztarget)
				fstring = url.PathEscape(fuzztarget)
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
				if datatype == "none" {
					fmt.Printf("No valid content type found for %s\n", requestURL)
					return nil
				}
				for _, p := range api.parameters {
					types := p.inputType
					switch types {
					case "string":
						fuzz.Fuzz(&fuzztarget)
						tstring = url.QueryEscape(fuzztarget)
						if p.in == "body" {
							data.Set(p.name, tstring)
						}
						if p.in == "query" {
							params.Set(p.name, tstring)
						}
						if p.in == "header" {
							headerlist.Set(p.name, tstring)
						}
						listparam[p.name] = tstring
					case "integer":
						fuzz.Fuzz(&fint)
						if p.in == "body" {
							data.Set(p.name, strconv.Itoa(fint))
						}
						if p.in == "query" {
							params.Set(p.name, strconv.Itoa(fint))
						}
						if p.in == "header" {
							headerlist.Set(p.name, tstring)
						}
						listparam[p.name] = strconv.Itoa(fint)
					case "boolean":
						fuzz.Fuzz(&fbool)
						if p.in == "body" {
							data.Set(p.name, strconv.FormatBool(fbool))
						}
						if p.in == "query" {
							params.Set(p.name, strconv.FormatBool(fbool))
						}
						if p.in == "header" {
							headerlist.Set(p.name, tstring)
						}
						listparam[p.name] = strconv.FormatBool(fbool)
					case "date":
						fuzz.Fuzz(&ftime)
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
				if headers {
					fuzz.Fuzz(&fint)
					req.Header.Add("Age", strconv.Itoa(fint))
					listparam["Age"] = strconv.Itoa(fint)
					fuzz.Fuzz(&ftime)
					req.Header.Add("Date", ftime.Format(time.RFC1123))
					listparam["Date"] = ftime.Format(time.RFC1123)
					fuzz.Fuzz(&ftime)
					req.Header.Add("Expires", ftime.Format(time.RFC1123))
					listparam["Expires"] = ftime.Format(time.RFC1123)
					fuzz.Fuzz(&ftime)
					req.Header.Add("Last-Modified", ftime.Format(time.RFC1123))
					listparam["Last-Modified"] = ftime.Format(time.RFC1123)
				}
				for key, value := range headerlist {
					req.Header.Add(key, value[0]) //forced header fuzzing
				}

				if requiresAuth {
					req.Header.Add("Authorization", "Bearer "+token)
				} else {
					f.Fuzz(&fstring)
					req.Header.Add("Authorization", fstring)
					fuzz.Fuzz(&fuzztarget)
					listparam["Authorization"] = fuzztarget
				}
				//COURIER STARTS HERE
				client := &http.Client{Timeout: time.Second * 5}
				startTime := time.Now()
				resp, err := client.Do(req)
				timeElapsed := time.Since(startTime).Milliseconds()
				if err != nil {
					if err, ok := err.(net.Error); ok && err.Timeout() {
						fmt.Printf("Timeout: Caused by: %s\n", req.URL)
						failureCount++
						if failureCount > 10 {
							fmt.Printf("Too many failures. Waiting %d seconds before continuing\n", backoff)
							time.Sleep(time.Duration(backoff) * time.Second)
							failureCount = 0
						}
						continue
					}
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode == 429 {
					fmt.Printf("Rate limited: Backing off for %d seconds\n", backoff)
					time.Sleep(time.Duration(backoff) * time.Second)
					continue
				}
				//ORACLE STARTS HERE
				printMutex.Lock()
				if avglen == -1 {
					if avglen == -1 {
						fmt.Println("There is no content length for this response")
					} else {
						avglen = int(resp.ContentLength)
						fmt.Printf("Base length of response for %s is: %d\n", requestURL, avglen)
					}

				}
				if resp.ContentLength != int64(avglen) && avglen != -1 && int(resp.ContentLength) != -1 {
					fmt.Printf("Response length mismatch: Caused by: %s\n", requestURL)
					reqDump, err := httputil.DumpRequestOut(req, true)
					if err != nil {
						fmt.Printf("Error dumping request: %s\n", err)
						return err
					}
					fmt.Printf("HTTP Request: %s\n", string(reqDump))

				}
				flag := true
				for _, responseDoc := range api.responses {
					if responseDoc.responseCode == resp.StatusCode {
						flag = false
						break
					}
				}
				fmt.Println("Response Status: ", resp.StatusCode)
				if dbg {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Printf("Error reading response body: %s\n", err)
						return err
					}
					fmt.Printf("Response Status: %s %s\n Responded in: %d milliseconds\n", requestURL, resp.Status, timeElapsed)
					fmt.Printf("Response Body: %s\n", string(bodyBytes))
				}
				if resp.StatusCode == 500 || flag {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Printf("Error reading response body: %s\n", err)
						return err
					}
					fmt.Printf("Interesting code %d: Caused by: %s\nReponse Body: %v\n", resp.StatusCode, requestURL, string(bodyBytes))
					fmt.Print("\nParameters:\n")
					for key, value := range listparam {
						fmt.Printf("%s -> %s\n", key, value)
						list = append(list, []byte(value))
					}
					fmt.Printf("\n")
					failureCount++
					if failureCount > 10 {
						fmt.Printf("Too many failures. Waiting %d seconds before continuing\n", backoff)
						time.Sleep(time.Duration(backoff) * time.Second)
						failureCount = 0
					}
					if headers {
						fmt.Printf("\nResponse Status: %s %s\n Responded in: %d milliseconds\n", requestURL, resp.Status, timeElapsed)
					}
					printMutex.Unlock()
				}
				i++
				if i >= len(list) {
					i = 0
				}
			}
		}
	}
}
