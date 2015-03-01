package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"log"
)

func CompressToBase64(logger *log.Logger, input string) (string, error) {
	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		logger.Printf("failed to create gzip writer: %s", err)
		return "", err
	}
	if _, err := gz.Write([]byte(input)); err != nil {
		logger.Printf("failed to compress string: %s", err)
		return "", err
	}
	if err := gz.Flush(); err != nil {
		logger.Printf("failed to flush gzip: %s", err)
		return "", err
	}
	if err := gz.Close(); err != nil {
		logger.Printf("failed to close gzip: %s", err)
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(b.Bytes())
	return encoded, nil
}

func DecompressFromBase64(logger *log.Logger, input string) (string, error) {
	// DecodedLen returns the maximum length in bytes of the decoded
	// data. But this is a maximum. You must use the 'n' return value
	// from the Decode call to know exactly how many bytes to use. If
	// you don't you'll feed the gzip reader garbage nulls at the end.
	// (Recall that base64 must be padded to the nearest 4 bytes).
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(input)))
	n, err := base64.StdEncoding.Decode(decoded, []byte(input))
	if err != nil {
		logger.Printf("failed to base64 decode: %s", err)
		return "", err
	}
	gz, err := gzip.NewReader(bytes.NewBuffer(decoded[:n]))
	if err != nil {
		logger.Printf("failed to create gzip reader: %s", err)
		return "", err
	}
	uncompressed, _ := ioutil.ReadAll(gz)
	if err != nil {
		logger.Printf("failed to decompress string: %s", err)
		return "", err
	}
	if err := gz.Close(); err != nil {
		logger.Printf("failed to close gzip: %s", err)
		return "", err
	}
	return string(uncompressed), nil
}
