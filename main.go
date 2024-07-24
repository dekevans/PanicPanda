package main

// trying to build api fuzzer: part 1
// three sections:
// poet: generate the api fuzzing code
// courier: send the requests
// oracle: check the responses

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to the API Fuzzer")
	fmt.Println("Input domain and paths (everything before the paths on the swaggerdoc):")
	controllerAddress, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading controller address:", err)
		return
	}
	controllerAddress = strings.TrimSpace(controllerAddress)
	fmt.Println("Input auth token (if none, leave blank):")
	token, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading token:", err)
		return
	}
	authflag := true
	if token == "" {
		fmt.Println("No token provided. Will not use token.")
		authflag = false
	}
	token = strings.TrimSpace(token)
	fmt.Println("Input timer:")
	timea, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading timer:", err)
		return
	}
	timea = strings.TrimSpace(timea)
	if timea == "" {
		fmt.Println("No timer provided. Will use default timer of 10 seconds.")
		timea = "10"
	}
	timer, err := strconv.Atoi(timea)
	if err != nil {
		fmt.Println("Error converting timer:", err)
		return
	}
	fmt.Println("Input swagger file path:")
	swagstr, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading swagger file:", err)
		return
	}
	var wordlist []string
	fmt.Println("Input wordlist file path: (if you want pure random data, leave blank)")
	wordlistPath, err := reader.ReadString('\n')
	wordlistPath = strings.TrimSpace(wordlistPath)
	if err != nil {
		fmt.Println("Error reading wordlist file:", err)
		return
	}
	if wordlistPath != "" {
		wordListFile, err := os.Open(wordlistPath)
		if err != nil {
			fmt.Println("Error opening wordlist file:", err)
			return
		}
		scanner := bufio.NewScanner(wordListFile)
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())
			wordlist = append(wordlist, word)
		}
		defer wordListFile.Close()
	}

	swagstr = strings.TrimSpace(swagstr)
	swag := parseSwag(swagstr)
	if swag == nil {
		fmt.Println("Error parsing swagger file")
		return
	}
	headers := false
	fmt.Println("Do you want to fuzz the headers? (Y/N)")
	log, err := reader.ReadString('\n')
	if err != nil && !(log == "y" || log == "n" || log == "Y" || log == "N") {
		fmt.Println("Error reading log option:", err)
		return
	} else if log == "y" || log == "Y" {
		headers = true
	}
	defang := false
	if !defang {
		threadManager(controllerAddress, swag, token, timer, authflag, headers, wordlist)
	}
}
func threadManager(controllerAddress string, apiList []apiDoc, args string, timer int, requiresAuth bool, headers bool, wordlist []string) {
	var wrkgrp sync.WaitGroup
	timeout, cancel := context.WithTimeout(context.Background(), time.Duration(timer)*time.Second)
	fmt.Printf("Starting the fuzzer for %d seconds\n", timer)
	id := 0
	var printMutex sync.Mutex
	for _, api := range apiList {
		wrkgrp.Add(1)
		go func(id int) {
			defer wrkgrp.Done()
			fullfunc(controllerAddress, api, args, timer, requiresAuth, headers, id, timeout, &printMutex, wordlist)
		}(id)
		id++
		//break
	}
	defer cancel()
	wrkgrp.Wait()
}
