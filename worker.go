package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	failureCount := 0
	listparam := make(map[string]string)
	// myInt gets a random value.
	avglen := -1
	i := 0
	if len(wordlist) > 0 {
		for {
			//fmt.Println("Fuzzing with wordlist")
			//TIMEOUT HERE
			select {
			case <-timeout.Done():
				fmt.Printf("%d seconds have elapsed. Ending fuzzing process for %s\n", timer, api.path)
				//printMutex.Unlock()
				return nil
			default:
				//POET w/ WORDLIST STARTS HERE
				winner := []byte(list[i])
				apiPath := api.path
				//f.Fuzz(&fstring)
				var fuzztarget string
				//printMutex.Lock()
				pathWinner := pathlist[i]
				//fmt.Printf("Fuzzing with: %s and with corpus of %s, %d\n", fuzztarget, winner, i)
				//printMutex.Unlock()
				fstring = url.PathEscape(pathWinner)
				apiPath = replacePlaceholder(apiPath, fstring)
				requestURL := fmt.Sprintf("%s%s", controllerAddress, apiPath)
				data := url.Values{}
				params := url.Values{}
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
						//fmt.Printf("Fuzzing with: %s\n", winner)
						tstring = replacePlaceholder(string(winner), url.QueryEscape(fuzztarget))
						if p.in == "body" {
							data.Set(p.name, tstring)
						}
						if p.in == "query" {
							params.Set(p.name, tstring)
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
						listparam[p.name] = strconv.FormatBool(fbool)
					case "date":
						fuzz.NewFromGoFuzz(winner).Fuzz(&ftime)
						if p.in == "body" {
							data.Set(p.name, ftime.Format(time.RFC1123))
						}
						if p.in == "query" {
							params.Set(p.name, ftime.Format(time.RFC1123))
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
					//fmt.Println("JSON Data: ", string(jsondata))
					//fmt.Println(string(jsondata))
					if err != nil {
						fmt.Printf("Error marshalling data: %s\n", err)
						return err
					}

					//fmt.Println("Request URL: ", reqUrl)
					//req, err = http.NewRequest(api.call, reqUrl, bytes.NewBuffer(jsondata))
					if len(jsondata) != 2 {
						req, err = http.NewRequest(api.call, reqUrl, bytes.NewBuffer(jsondata))
					} else {
						req, err = http.NewRequest(api.call, reqUrl, nil)
					}
					if err != nil {
						fmt.Printf("Error creating request: %s\n", err)
						return err
					}

					//req, err = http.NewRequest(api.call, requestURL, nil)
					/*reqDump, err := httputil.DumpRequestOut(req, true)
					if err != nil {
						fmt.Printf("Error dumping request: %s\n", err)
						return err
					}
					fmt.Printf("HTTP Request: %s\n", string(reqDump))
					*/
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
					/*reqDump, err := httputil.DumpRequestOut(req, true)
					if err != nil {
						fmt.Printf("Error dumping request: %s\n", err)
						return err
					}
					fmt.Printf("HTTP Request: %s\n", string(reqDump))*/
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
					//fuzz.NewFromGoFuzz(winner).Fuzz(&fint)
					//req.Header.Add("Content-Length", strconv.Itoa(fint))
					//listparam["Content-Length"] = strconv.Itoa(fint)
				}
				/*for _, c := range api.consumes {
					req.Header.Add("Content-Type", c)
				}*/
				if requiresAuth {
					//fmt.Printf("Token: %s\n", token)
					//req.Header.Add("Authorization", "Bearer "+token)
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
				//fmt.Print("Sending request using ", client.Timeout, " timeout\n", startTime, avglen)
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
				//fmt.Println("Response Status: ", resp.Status)
				//ORACLE STARTS HERE
				//time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				printMutex.Lock()
				if avglen == -1 {
					//fmt.Printf("Initial length of response is: %d\n", int(resp.ContentLength)) //200 responses don't have content length?!?!?!?!?
					if avglen == -1 {
						fmt.Println("There is no content length for this response")
					} else {
						avglen = int(resp.ContentLength)
						fmt.Printf("Base length of response for %s is: %d\n", requestURL, avglen)
					}

				}
				if resp.ContentLength != int64(avglen) && avglen != -1 && int(resp.ContentLength) != -1 {
					fmt.Printf("Response length mismatch: Caused by: %s\n", requestURL)
					/*fmt.Println("\nParameters:")
					for key, value := range listparam {
						fmt.Printf("%s -> %s\n", key, value)
					}*/
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
					//fmt.Printf("Returned in %s seconds\n", )
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
			//fmt.Println("Fuzzing without wordlist")
			//TIMEOUT HERE
			select {
			case <-timeout.Done():
				fmt.Printf("%d seconds have elapsed. Ending fuzzing process for %s\n", timer, api.path)
				//printMutex.Unlock()
				return nil
			default:
				//POET w/ WORDLIST STARTS HERE
				apiPath := api.path
				//f.Fuzz(&fstring)
				fuzz := fuzz.New()
				var fuzztarget string
				fuzz.Fuzz(&fuzztarget)
				//printMutex.Lock()
				//fmt.Printf("Fuzzing with: %s and with corpus of %s, %d\n", fuzztarget, winner, i)
				//printMutex.Unlock()
				fstring = url.PathEscape(fuzztarget)
				apiPath = replacePlaceholder(apiPath, fstring)
				requestURL := fmt.Sprintf("%s%s", controllerAddress, apiPath)
				data := url.Values{}
				params := url.Values{}
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
						//fmt.Printf("Fuzzing with: %s\n", winner)
						tstring = url.QueryEscape(fuzztarget)
						if p.in == "body" {
							data.Set(p.name, tstring)
						}
						if p.in == "query" {
							params.Set(p.name, tstring)
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
						listparam[p.name] = strconv.Itoa(fint)
					case "boolean":
						fuzz.Fuzz(&fbool)
						//list = append(list, []byte(strconv.FormatBool(fbool)))
						if p.in == "body" {
							data.Set(p.name, strconv.FormatBool(fbool))
						}
						if p.in == "query" {
							params.Set(p.name, strconv.FormatBool(fbool))
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
					//fmt.Println("JSON Data: ", string(jsondata))
					//fmt.Println(string(jsondata))
					if err != nil {
						fmt.Printf("Error marshalling data: %s\n", err)
						return err
					}

					//fmt.Println("Request URL: ", reqUrl)
					//req, err = http.NewRequest(api.call, reqUrl, bytes.NewBuffer(jsondata))
					if len(jsondata) != 2 {
						req, err = http.NewRequest(api.call, reqUrl, bytes.NewBuffer(jsondata))
					} else {
						req, err = http.NewRequest(api.call, reqUrl, nil)
					}
					if err != nil {
						fmt.Printf("Error creating request: %s\n", err)
						return err
					}

					//req, err = http.NewRequest(api.call, requestURL, nil)
					/*reqDump, err := httputil.DumpRequestOut(req, true)
					if err != nil {
						fmt.Printf("Error dumping request: %s\n", err)
						return err
					}
					fmt.Printf("HTTP Request: %s\n", string(reqDump))
					*/
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
					/*reqDump, err := httputil.DumpRequestOut(req, true)
					if err != nil {
						fmt.Printf("Error dumping request: %s\n", err)
						return err
					}
					fmt.Printf("HTTP Request: %s\n", string(reqDump))*/
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
					//fuzz.Fuzz(&fint)
					//req.Header.Add("Content-Length", strconv.Itoa(fint))
					//listparam["Content-Length"] = strconv.Itoa(fint)
				}
				/*for _, c := range api.consumes {
					req.Header.Add("Content-Type", c)
				}*/
				if requiresAuth {
					//fmt.Printf("Token: %s\n", token)
					//req.Header.Add("Authorization", "Bearer "+token)
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
				//fmt.Print("Sending request using ", client.Timeout, " timeout\n", startTime, avglen)
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
				//fmt.Println("Response Status: ", resp.Status)
				//ORACLE STARTS HERE
				//time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				printMutex.Lock()
				if avglen == -1 {
					//fmt.Printf("Initial length of response is: %d\n", int(resp.ContentLength)) //200 responses don't have content length?!?!?!?!?
					if avglen == -1 {
						fmt.Println("There is no content length for this response")
					} else {
						avglen = int(resp.ContentLength)
						fmt.Printf("Base length of response for %s is: %d\n", requestURL, avglen)
					}

				}
				if resp.ContentLength != int64(avglen) && avglen != -1 && int(resp.ContentLength) != -1 {
					fmt.Printf("Response length mismatch: Caused by: %s\n", requestURL)
					/*fmt.Println("\nParameters:")
					for key, value := range listparam {
						fmt.Printf("%s -> %s\n", key, value)
					}*/
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
					//fmt.Printf("Returned in %s seconds\n", )
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
	}
}
