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
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MapResponse struct {
	Out string `json:"out"`
}

var wordRe = regexp.MustCompile(`[A-Za-z0-9']+`) // keeps contractions like don't

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
	outPrefix := getenv("OUT_PREFIX", "mr/maps")

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

	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		in := r.URL.Query().Get("s3")
		if in == "" {
			http.Error(w, "missing ?s3=s3://bucket/key", 400)
			return
		}

		bucket, key, err := parseS3URL(in)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		obj, err := s3c.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
		if err != nil {
			http.Error(w, "s3 get error: "+err.Error(), 500)
			return
		}
		defer obj.Body.Close()

		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(obj.Body)
		text := strings.ToLower(buf.String())

		counts := map[string]int{}
		for _, tok := range wordRe.FindAllString(text, -1) {
			counts[tok]++
		}

		// Write JSON to S3
		outKey := fmt.Sprintf("%s/%s_%s.json", outPrefix, sanitizeKey(key), time.Now().UTC().Format("20060102T150405Z"))
		body, _ := json.Marshal(counts)

		_, err = s3c.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucket,
			Key:    &outKey,
			Body:   bytes.NewReader(body),
		})
		if err != nil {
			http.Error(w, "s3 put error: "+err.Error(), 500)
			return
		}

		out := s3URL(bucket, outKey)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MapResponse{Out: out})

		log.Printf("map ok input=%s unique=%d out=%s dur=%s", in, len(counts), out, time.Since(start))
	})

	log.Printf("mapper listening on %s region=%s outPrefix=%s", addr, region, outPrefix)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func sanitizeKey(key string) string {
	// make chunk key safe-ish to embed in output filename
	key = strings.ReplaceAll(key, "/", "_")
	key = strings.ReplaceAll(key, ".", "_")
	if len(key) > 60 {
		key = key[len(key)-60:]
	}
	return key
}
