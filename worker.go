package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	fuzz "github.com/google/gofuzz"
)

func fullfunc(controllerAddress string, api apiDoc, token string, timer int, requiresAuth bool, headers bool, id int, timeout context.Context, printMutex *sync.Mutex, wordlist []string) error {
	f := fuzz.New()
	var fint int
	var fbool bool
	var fstring string
	var ftime time.Time
	list := [][]byte{}
	for _, word := range wordlist {
		list = append(list, []byte(word))
	}
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
				//printMutex.Unlock()
				return nil
			default:
				//POET w/ WORDLIST STARTS HERE
				winner := []byte(list[i])
				apiPath := api.path
				//f.Fuzz(&fstring)
				var fuzztarget string
				fuzz.NewFromGoFuzz(winner).Fuzz(&fuzztarget)
				//printMutex.Lock()
				//fmt.Printf("Fuzzing with: %s and with corpus of %s, %d\n", fuzztarget, winner, i)
				//printMutex.Unlock()
				fstring = url.PathEscape(fuzztarget)
				apiPath = replacePlaceholder(apiPath, fstring)
				requestURL := fmt.Sprintf("%s%s", controllerAddress, apiPath)
				data := url.Values{}
				var tstring string
				for _, p := range api.parameters {
					types := p.inputType
					switch types {
					case "string":
						fuzz.NewFromGoFuzz(winner).Fuzz(&fuzztarget)
						//fmt.Printf("Fuzzing with: %s\n", winner)
						tstring = replacePlaceholder(string(winner), url.QueryEscape(fuzztarget))
						data.Set(p.name, tstring)
						listparam[p.name] = tstring
					case "integer":
						fuzz.NewFromGoFuzz(winner).Fuzz(&fint)
						data.Set(p.name, tstring)
						listparam[p.name] = strconv.Itoa(fint)
					case "boolean":
						fuzz.NewFromGoFuzz(winner).Fuzz(&fuzztarget)
						data.Set(p.name, tstring)
						list = append(list, []byte(strconv.FormatBool(fbool)))
						listparam[p.name] = strconv.FormatBool(fbool)
					}
				}
				req, err := http.NewRequest(api.call, requestURL, strings.NewReader(data.Encode()))
				if err != nil {
					fmt.Println("Error creating request:", err)
					return err
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
					fuzz.NewFromGoFuzz(winner).Fuzz(&fint)
					req.Header.Add("Content-Length", strconv.Itoa(fint))
					listparam["Content-Length"] = strconv.Itoa(fint)
				}
				for _, c := range api.consumes {
					req.Header.Add("Content-Type", c)
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
				//fmt.Print("Sending request using ", client.Timeout, " timeout\n", startTime, avglen)
				resp, err := client.Do(req)
				timeElapsed := time.Since(startTime).Milliseconds()
				if err != nil {
					if err, ok := err.(net.Error); ok && err.Timeout() {
						fmt.Printf("Timeout: Caused by: %s\n", req.URL)
						continue
					}
					return err
				}
				defer resp.Body.Close()
				//ORACLE STARTS HERE
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				printMutex.Lock()
				//defer printMutex.Unlock()
				if avglen == -1 {
					avglen = int(resp.ContentLength)
					fmt.Printf("Base length of response is: %d\n", avglen)
				}
				if resp.ContentLength != int64(avglen) {
					fmt.Printf("Response length mismatch: Caused by: %s\n", requestURL)
					fmt.Println("\nParameters:")
					for key, value := range listparam {
						fmt.Printf("%s -> %s\n", key, value)
					}
				}
				flag := false
				for _, responseDoc := range api.responses {
					if responseDoc.responseCode == resp.StatusCode {
						flag = true
						break
					}
				}
				if resp.StatusCode == 500 || flag {
					if resp.StatusCode == 500 {
						bodyBytes, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							fmt.Printf("Error reading response body: %s\n", err)
							return err
						}
						fmt.Printf("\nInternal Server Error: Caused by: %s\nReponse Body: %v\n", requestURL, string(bodyBytes))
						//fmt.Printf("Returned in %s seconds\n", )
						fmt.Print("\nParameters:\n")
						for key, value := range listparam {
							fmt.Printf("%s -> %s\n", key, value)
							list = append(list, []byte(value))
						}
						fmt.Printf("\n")
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
			select {
			// TIMEOUT STARTS HERE
			case <-timeout.Done():
				fmt.Printf("%d seconds have elapsed. Ending fuzzing process for %s\n", timer, api.path)
				return nil
			default:
				apiPath := api.path
				f.Fuzz(&fstring)
				fstring = url.QueryEscape(fstring)
				apiPath = replacePlaceholder(apiPath, fstring)
				requestURL := fmt.Sprintf("%s%s", controllerAddress, apiPath)
				data := url.Values{}
				for _, p := range api.parameters {
					types := p.inputType
					switch types {
					case "string":
						f.Fuzz(&fstring)
						data.Set(p.name, fstring)
						fstring = url.QueryEscape(fstring)
						listparam[p.name] = fstring
					case "integer":
						f.Fuzz(&fint)
						data.Set(p.name, strconv.Itoa(fint))
						listparam[p.name] = strconv.Itoa(fint)
					case "boolean":
						f.Fuzz(&fbool)
						data.Set(p.name, strconv.FormatBool(fbool))
						listparam[p.name] = strconv.FormatBool(fbool)
					}
				}
				req, err := http.NewRequest(api.call, requestURL, strings.NewReader(data.Encode()))
				if headers {
					f.Fuzz(&fint)
					req.Header.Add("Age", strconv.Itoa(fint))
					listparam["Age"] = strconv.Itoa(fint)
					f.Fuzz(&ftime)
					req.Header.Add("Date", ftime.Format(time.RFC1123))
					listparam["Date"] = ftime.Format(time.RFC1123)
					f.Fuzz(&ftime)
					req.Header.Add("Expires", ftime.Format(time.RFC1123))
					listparam["Expires"] = ftime.Format(time.RFC1123)
					f.Fuzz(&ftime)
					req.Header.Add("Last-Modified", ftime.Format(time.RFC1123))
					listparam["Last-Modified"] = ftime.Format(time.RFC1123)
					f.Fuzz(&fint)
					req.Header.Add("Content-Length", strconv.Itoa(fint))
					listparam["Content-Length"] = strconv.Itoa(fint)
					if err != nil {
						return err
					}
				}

				f.Fuzz(&fstring)
				for _, c := range api.consumes {
					req.Header.Add("Content-Type", c)
				}
				if requiresAuth {
					req.Header.Add("Authorization", "Bearer "+token)
				} else {
					f.Fuzz(&fstring)
					req.Header.Add("Authorization", fstring)
					listparam["Authorization"] = fstring
				}
				//COURIER STARTS HERE
				client := &http.Client{Timeout: time.Second * 5}
				startTime := time.Now()
				resp, err := client.Do(req)
				timeElapsed := time.Since(startTime).Milliseconds()
				if err != nil {
					if err, ok := err.(net.Error); ok && err.Timeout() {
						fmt.Printf("Timeout: Caused by: %s\n", req.URL)
						continue
					}
					return err
				}
				defer resp.Body.Close()
				//wait random time
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				printMutex.Lock()
				//defer printMutex.Unlock()
				if avglen == -1 && headers {
					fmt.Printf("Initial length of response is: %d\n", avglen)
				}
				if resp.ContentLength != int64(avglen) {
					fmt.Printf("Response length mismatch: Caused by: %s\n", requestURL)
					fmt.Print("\nParameters:")
					//print map
					for key, value := range listparam {
						fmt.Printf("%s -> %s\n", key, value)
					}
				}
				//ORACLE STARTS HERE
				flag := false
				for _, responseDoc := range api.responses {
					if responseDoc.responseCode == resp.StatusCode {
						flag = true
						break
					}
				}
				if resp.StatusCode == 500 || flag {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return err
					}
					fmt.Printf("\nInteresting %d code: Caused by Response Body: %v\n Message: %s\n", resp.StatusCode, requestURL, string(bodyBytes))
					//fmt.Printf("Returned in %s seconds\n", )
					fmt.Print("\nParameters:\n")
					for key, value := range listparam {
						fmt.Printf("%s -> %s\n", key, value)
					}
					fmt.Printf("\n")
					fmt.Printf("\nResponse Status: %s %s\n Responded in: %d milliseconds\n", requestURL, resp.Status, timeElapsed)
				}
				printMutex.Unlock()
			}
		}
	}
}
