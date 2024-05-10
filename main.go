package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

var (
	useInput string
	withId   *string
	fromTime *string
	toTime   *string
	withWord *string
)

func main() {
	parseFlags()

	filterMap := prepareFilters()

	bucket, key := parseInput()

	svc := createSession()

	resp := getObject(bucket, key, svc)

	defer resp.Body.Close()

	gz := decompress(resp.Body)

	defer gz.Close()

	scanner := bufio.NewScanner(gz)
	for scanner.Scan() {
		processLine(scanner.Text(), filterMap)
	}
}

func parseFlags() {
	flag.StringVar(&useInput, "input", "", "s3 arn")
	withId = flag.String("with-id", "", "id to search")
	fromTime = flag.String("from-time", "", "from time to search")
	toTime = flag.String("to-time", "", "to time to search")
	withWord = flag.String("with-word", "", "word to search")
	flag.Parse()

	if useInput == "" {
		log.Fatal("input flag is required (s3 arn of the bucket and key)")
	}
}

func prepareFilters() map[string]func(JsonResponse, string) bool {
	filterMap := make(map[string]func(JsonResponse, string) bool)

	if *withWord != "" {
		filterMap["with-word"] = prepareWithWordFilter()
	}

	if *withId != "" {
		filterMap["with-id"] = prepareWithIdFilter()
	}

	if *fromTime != "" {
		filterMap["from-time"] = prepareFromTimeFilter()
	}

	if *toTime != "" {
		filterMap["to-time"] = prepareToTimeFilter()
	}

	return filterMap
}

func prepareWithWordFilter() func(JsonResponse, string) bool {
	return func(jsonResponse JsonResponse, wword string) bool {
		for _, word := range jsonResponse.Words {
			if word == wword {
				return true
			}
		}
		return false
	}
}

func prepareWithIdFilter() func(JsonResponse, string) bool {
	return func(jsonResponse JsonResponse, wid string) bool {
		IdintValue, err := strconv.Atoi(wid)
		if err != nil {
			exitErrorf("unknown error occurred, %v", err)
		}
		return jsonResponse.Id == IdintValue
	}
}

func prepareFromTimeFilter() func(JsonResponse, string) bool {
	return func(jsonResponse JsonResponse, wfromTime string) bool {
		jsonRespTime, _ := time.Parse(time.RFC3339, jsonResponse.Time)
		fromTime, _ := time.Parse(time.RFC3339, wfromTime)
		return jsonRespTime.Equal(fromTime) || jsonRespTime.After(fromTime)
	}
}

func prepareToTimeFilter() func(JsonResponse, string) bool {
	return func(jsonResponse JsonResponse, wtoTime string) bool {
		jsonRespTime, _ := time.Parse(time.RFC3339, jsonResponse.Time)
		toTime, _ := time.Parse(time.RFC3339, wtoTime)
		return jsonRespTime.Equal(toTime) || jsonRespTime.Before(toTime)
	}
}

func parseInput() (string, string) {
	u, _ := url.Parse(useInput)
	return u.Host, u.Path
}

func createSession() *s3.S3 {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	return s3.New(sess)
}

func getObject(bucket, key string, svc *s3.S3) *s3.GetObjectOutput {
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

	return resp
}

func decompress(body io.Reader) *gzip.Reader {
	gz, err := gzip.NewReader(body)
	if err != nil {
		exitErrorf("unknown error occurred, %v", err)
	}
	return gz
}

func processLine(line string, filterMap map[string]func(JsonResponse, string) bool) {
	data := []byte(line)
	jsonResponse := JsonResponse{}
	json.Unmarshal(data, &jsonResponse)

	hasAllFlags := true
	for filter := range filterMap {
		if !filterMap[filter](jsonResponse, getFlagValue(filter)) {
			hasAllFlags = false
		}
	}
	if hasAllFlags {
		fmt.Println(jsonResponse)
	}
}

func getFlagValue(name string) string {
	switch name {
	case "input":
		return useInput
	case "with-id":
		return *withId
	case "from-time":
		return *fromTime
	case "to-time":
		return *toTime
	case "with-word":
		return *withWord
	default:
		return ""
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
