package util

// Implementation based on github.com/grahamedgecombe/ct-submit

import (
	"bytes"
	_ "crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

type signedCertificateTimestamp struct {
	Version    uint8  `json:"sct_version"`
	LogID      string `json:"id"`
	Timestamp  int64  `json:"timestamp"`
	Extensions string `json:"extensions"`
	Signature  string `json:"signature"`
}

func (sct signedCertificateTimestamp) Write(w io.Writer) error {
	// Version
	if err := binary.Write(w, binary.BigEndian, sct.Version); err != nil {
		return err
	}

	// LogID
	bytes, err := base64.StdEncoding.DecodeString(sct.LogID)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}

	// Timestamp
	if err := binary.Write(w, binary.BigEndian, sct.Timestamp); err != nil {
		return err
	}

	// Extensions
	bytes, err = base64.StdEncoding.DecodeString(sct.Extensions)
	if err != nil {
		return err
	}

	length := len(bytes)
	if length > 65535 {
		return errors.New("extensions are too long")
	}

	if err := binary.Write(w, binary.BigEndian, uint16(length)); err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}

	// Signature
	bytes, err = base64.StdEncoding.DecodeString(sct.Signature)
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

// CTLogProvider holds information about a CTLogProvider, currently active in Chrome
type CTLogProvider struct {
	Name string
	Key  string
	URL  string
}

// Submit the given payload to the CTLogProvider and write the response into the given output file
func (ctLog CTLogProvider) Submit(payload []byte, outputFile string, wg *sync.WaitGroup) {
	defer wg.Done()

	addChainURL, err := url.Parse(ctLog.URL)
	if err != nil {
		log.Fatalf("Failed parsing URL %s! Err: %s", ctLog.URL, err)
		os.Exit(1)
	}

	addChainURL, _ = addChainURL.Parse("ct/v1/add-chain")

	// create httpclient with timeout of 10s
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// send add-chain message to the log
	response, err := httpClient.Post(addChainURL.String(), "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("Failed sending post request to log server %s. Error: %s", ctLog.URL, err)
		return
	}

	if response.StatusCode != http.StatusOK {
		log.Printf("Unexpected status %s from server %s: \n\n", response.Status, ctLog.URL)
		io.Copy(os.Stderr, response.Body)
		return
	}

	// decode JSON SCT structure
	payload, err = ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	sct := signedCertificateTimestamp{}
	if err = json.Unmarshal(payload, &sct); err != nil {
		panic(err)
	}

	f, err := os.Create(outputFile)

	defer f.Close()
	if err != nil {
		log.Printf("Failed creating output file %s. Error: %s", outputFile, err)
		return
	}

	// write binary SCT structure to stdout
	if err = sct.Write(f); err != nil {
		panic(err)
	}
}
