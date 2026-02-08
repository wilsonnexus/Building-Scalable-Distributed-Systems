package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ReduceResponse struct {
	Out   string `json:"out"`
	Files int    `json:"files"`
}

func parseS3URL(s string) (bucket, key string, err error) {
	if !strings.HasPrefix(s, "s3://") {
		return "", "", errors.New("must start with s3://")
	}
	rest := strings.TrimPrefix(s, "s3://")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("invalid s3 url, expected s3://bucket/key")
	}
	return parts[0], parts[1], nil
}

func s3URL(bucket, key string) string {
	return "s3://" + bucket + "/" + key
}

func main() {
	addr := getenv("ADDR", ":8080")
	region := getenv("AWS_REGION", "us-east-1")
	outPrefix := getenv("OUT_PREFIX", "mr/reduce")

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("aws config: %v", err)
	}
	s3c := s3.NewFromConfig(cfg)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/reduce", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ins := r.URL.Query()["in"]
		if len(ins) < 1 {
			http.Error(w, "provide at least one ?in=s3://bucket/key (repeat ?in=...)", 400)
			return
		}

		// Require same bucket for simplicity (you can relax later)
		firstBucket, _, err := parseS3URL(ins[0])
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		final := map[string]int{}
		for _, in := range ins {
			bucket, key, err := parseS3URL(in)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if bucket != firstBucket {
				http.Error(w, "all inputs must be in same bucket for this simple reducer", 400)
				return
			}

			obj, err := s3c.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
			if err != nil {
				http.Error(w, "s3 get error: "+err.Error(), 500)
				return
			}
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(obj.Body)
			obj.Body.Close()

			part := map[string]int{}
			if err := json.Unmarshal(buf.Bytes(), &part); err != nil {
				http.Error(w, "bad json in "+in+": "+err.Error(), 400)
				return
			}
			for k, v := range part {
				final[k] += v
			}
		}

		// Write final json
		outKey := fmt.Sprintf("%s/final_%s.json", outPrefix, time.Now().UTC().Format("20060102T150405Z"))
		body, _ := json.MarshalIndent(orderKeys(final), "", "  ")

		_, err = s3c.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &firstBucket,
			Key:    &outKey,
			Body:   bytes.NewReader(body),
		})
		if err != nil {
			http.Error(w, "s3 put error: "+err.Error(), 500)
			return
		}

		out := s3URL(firstBucket, outKey)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ReduceResponse{Out: out, Files: len(ins)})

		log.Printf("reduce ok files=%d out=%s dur=%s", len(ins), out, time.Since(start))
	})

	log.Printf("reducer listening on %s region=%s outPrefix=%s", addr, region, outPrefix)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

// Makes output stable for easier diffing / demos
func orderKeys(m map[string]int) map[string]int {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string]int, len(m))
	for _, k := range keys {
		ordered[k] = m[k]
	}
	return ordered
}
