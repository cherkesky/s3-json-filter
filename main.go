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
	useInput, withId, fromTime, toTime, withWord := parseFlags()

	if useInput == "" {
		log.Fatal("input flag is required (s3 arn of the bucket and key)")
	}

	filterMap, flagMap := createFilters(withId, fromTime, toTime, withWord)

	bucket, key := parseInput(useInput)

	sess := createAWSSession()

	resp, err := getObjectFromS3(sess, bucket, key)
	if err != nil {
		handleAWSError(err, bucket, key)
	}
	defer resp.Body.Close()

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		exitErrorf("unknown error occurred, %v", err)
	}
	defer gz.Close()

	processLines(gz, filterMap, flagMap)
}

func parseFlags() (string, *string, *string, *string, *string) {
	useInput := flag.String("input", "", "s3 arn")
	withId := flag.String("with-id", "", "id to search")
	fromTime := flag.String("from-time", "", "from time to search")
	toTime := flag.String("to-time", "", "to time to search")
	withWord := flag.String("with-word", "", "word to search")
	flag.Parse()
	return *useInput, withId, fromTime, toTime, withWord
}

func createFilters(withId, fromTime, toTime, withWord *string) (map[string]func(JsonResponse, string) bool, map[string]string) {
	filterMap := make(map[string]func(JsonResponse, string) bool)
	flagMap := make(map[string]string)

	addFilter := func(name string, flagValue string, filter func(JsonResponse, string) bool) {
		flagMap[name] = flagValue
		filterMap[name] = filter
	}

	addStringFilter := func(name string, flagValue *string, filter func(JsonResponse, string) bool) {
		if isFlagPassed(name) {
			addFilter(name, *flagValue, filter)
		}
	}

	// with-word
	addStringFilter("with-word", withWord, func(jsonResponse JsonResponse, wword string) bool {
		for _, word := range jsonResponse.Words {
			if word == wword {
				return true
			}
		}
		return false
	})

	// with-id
	addStringFilter("with-id", withId, func(jsonResponse JsonResponse, wid string) bool {
		id, err := strconv.Atoi(wid)
		if err != nil {
			exitErrorf("unknown error occurred, %v", err)
		}
		return jsonResponse.Id == id
	})

	// fromTime
	addStringFilter("from-time", fromTime, func(jsonResponse JsonResponse, wfromTime string) bool {
		jsonRespTime, _ := time.Parse(time.RFC3339, jsonResponse.Time)
		fromTime, _ := time.Parse(time.RFC3339, wfromTime)
		return jsonRespTime.Equal(fromTime) || jsonRespTime.After(fromTime)
	})

	// toTime
	addStringFilter("to-time", toTime, func(jsonResponse JsonResponse, wtoTime string) bool {
		jsonRespTime, _ := time.Parse(time.RFC3339, jsonResponse.Time)
		toTime, _ := time.Parse(time.RFC3339, wtoTime)
		return jsonRespTime.Equal(toTime) || jsonRespTime.Before(toTime)
	})

	return filterMap, flagMap
}

func parseInput(useInput string) (string, string) {
	u, _ := url.Parse(useInput)
	return u.Host, u.Path
}

func createAWSSession() *session.Session {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	return sess
}

func getObjectFromS3(sess *session.Session, bucket, key string) (*s3.GetObjectOutput, error) {
	svc := s3.New(sess)
	return svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}

func handleAWSError(err error, bucket, key string) {
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

func processLines(gz *gzip.Reader, filterMap map[string]func(JsonResponse, string) bool, flagMap map[string]string) {
	scanner := bufio.NewScanner(gz)
	for scanner.Scan() {
		data := []byte(scanner.Text())
		var jsonResponse JsonResponse
		if err := json.Unmarshal(data, &jsonResponse); err != nil {
			exitErrorf("error unmarshalling JSON: %v", err)
		}

		// filtering
		hasAllFlags := true
		for filter := range filterMap {
			if !filterMap[filter](jsonResponse, flagMap[filter]) {
				hasAllFlags = false
				break
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
