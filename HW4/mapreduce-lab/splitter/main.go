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
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type SplitResponse struct {
	Chunks []string `json:"chunks"`
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
	outPrefix := getenv("OUT_PREFIX", "mr/chunks") // where to write chunks inside bucket

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

	http.HandleFunc("/split", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		q := r.URL.Query()
		in := q.Get("s3")
		if in == "" {
			http.Error(w, "missing ?s3=s3://bucket/key", 400)
			return
		}
		n := 3
		if q.Get("chunks") != "" {
			v, err := strconv.Atoi(q.Get("chunks"))
			if err != nil || v < 1 || v > 50 {
				http.Error(w, "invalid chunks (1..50)", 400)
				return
			}
			n = v
		}

		inBucket, inKey, err := parseS3URL(in)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		obj, err := s3c.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &inBucket,
			Key:    &inKey,
		})
		if err != nil {
			http.Error(w, "s3 get error: "+err.Error(), 500)
			return
		}
		defer obj.Body.Close()

		// Read all (fine for class-sized inputs; for huge files you’d stream)
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(obj.Body)
		data := buf.Bytes()
		if len(data) == 0 {
			http.Error(w, "input file empty", 400)
			return
		}

		// Split by lines so chunks are readable + stable
		lines := strings.Split(string(data), "\n")
		chunkSize := (len(lines) + n - 1) / n

		ts := time.Now().UTC().Format("20060102T150405Z")
		baseName := sanitizeBaseName(inKey)
		outBucket := inBucket

		var outURLs []string
		for i := 0; i < n; i++ {
			from := i * chunkSize
			if from >= len(lines) {
				break
			}
			to := (i + 1) * chunkSize
			if to > len(lines) {
				to = len(lines)
			}
			chunkText := strings.Join(lines[from:to], "\n")

			outKey := fmt.Sprintf("%s/%s_%s_chunk%02d.txt", outPrefix, baseName, ts, i)
			_, err := s3c.PutObject(ctx, &s3.PutObjectInput{
				Bucket: &outBucket,
				Key:    &outKey,
				Body:   strings.NewReader(chunkText),
			})
			if err != nil {
				http.Error(w, "s3 put error: "+err.Error(), 500)
				return
			}
			outURLs = append(outURLs, s3URL(outBucket, outKey))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SplitResponse{Chunks: outURLs})

		log.Printf("split ok input=%s chunks=%d out=%d dur=%s", in, n, len(outURLs), time.Since(start))
	})

	log.Printf("splitter listening on %s region=%s outPrefix=%s", addr, region, outPrefix)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func sanitizeBaseName(key string) string {
	// turn "input/myfile.txt" -> "myfile"
	name := key
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.TrimSuffix(name, ".txt")
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	if name == "" {
		return "input"
	}
	return name
}
