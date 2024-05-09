package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type JsonResponse struct {
	Id    int      `json:"id"`
	Time  string   `json:"time"`
	Words []string `json:"words"`
}

func main() {
	// set flags
	useInput := flag.String("input", "", "s3 arn")
	withId := flag.String("with-id", "", "id to search")
	fromTime := flag.String("from-time", "", "from time to search")
	toTime := flag.String("to-time", "", "to time to search")
	withWord := flag.String("with-word", "", "word to search")
	flag.Parse()

	if !isFlagPassed("input") {
		log.Fatal("input flag is required (s3 arn of the bucket and key)")
	}

	filterMap := map[string]func(JsonResponse, string) bool{}
	flagMap := map[string]string{}

	// filter function declaration map
	// with-word
	if isFlagPassed("with-word") {
		flagMap["with-word"] = *withWord
		filterMap["with-word"] = func(jsonResponse JsonResponse, wword string) bool {
			isFound := false
			for _, word := range jsonResponse.Words {
				if word == wword {
					isFound = true
					break
				}
			}
			if isFound {
				return true
			} else {
				return false
			}
		}
	}
	// with-id
	if isFlagPassed("with-id") {
		flagMap["with-id"] = *withId
		filterMap["with-id"] = func(jsonResponse JsonResponse, wid string) bool {
			IdintValue := 0
			IdintValue, err := strconv.Atoi(wid)
			if err != nil {
				exitErrorf("unknown error occurred, %v", err)
			}
			if jsonResponse.Id == IdintValue {
				return true
			} else {
				return false
			}
		}
	}
	// fromTime
	if isFlagPassed("from-time") {
		flagMap["from-time"] = *fromTime
		filterMap["from-time"] = func(jsonResponse JsonResponse, wfromTime string) bool {
			isFrom := false
			jsonRespTime, _ := time.Parse(
				time.RFC3339, jsonResponse.Time)
			fromTime, _ := time.Parse(
				time.RFC3339, wfromTime)
			isFrom = fromTime.Equal(jsonRespTime) || jsonRespTime.After(fromTime)

			return isFrom
		}
	}

	// toTime
	if isFlagPassed("to-time") {
		flagMap["to-time"] = *toTime
		filterMap["to-time"] = func(jsonResponse JsonResponse, wtoTime string) bool {
			isFrom := false
			jsonRespTime, _ := time.Parse(
				time.RFC3339, jsonResponse.Time)
			fromTime, _ := time.Parse(
				time.RFC3339, wtoTime)
			isFrom = fromTime.Equal(jsonRespTime) || jsonRespTime.Before(fromTime)

			return isFrom
		}
	}

	// parse input
	u, _ := url.Parse(*useInput)
	bucket := u.Host
	key := u.Path

	// create session
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	svc := s3.New(sess)

	// get object
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				exitErrorf("bucket %s does not exist", bucket)
			case s3.ErrCodeNoSuchKey:
				exitErrorf("object with key %s does not exist in bucket %s", key, bucket)
			}
		}
		exitErrorf("unknown error occurred, %v", err)
	}
	defer resp.Body.Close()

	// decompress
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		exitErrorf("unknown error occurred, %v", err)
	}
	defer gz.Close()

	// read line by line
	scanner := bufio.NewScanner(gz)
	for scanner.Scan() {
		data := []byte(scanner.Text())
		jsonResponse := JsonResponse{}
		json.Unmarshal(data, &jsonResponse)

		// filtering
		hasAllFlags := true
		for filter := range filterMap {
			if !filterMap[filter](jsonResponse, flagMap[filter]) {
				hasAllFlags = false
			}
		}
		if hasAllFlags {
			fmt.Println(jsonResponse)
		}
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
