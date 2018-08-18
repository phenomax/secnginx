package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// DownloadFile downloads a file and saves it to the given path
func DownloadFile(url string, target string) error {

	// Create the file
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("NginX Download Server returned HTTP Status Code %d", resp.StatusCode)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// CopyFile copies a file from the source to the provided destination
func CopyFile(source, dest string) {
	from, err := os.Open(source)
	if err != nil {
		log.Fatalf("Failed copying file %s to %s Error: %s", source, dest, err)
	}
	defer from.Close()

	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Failed copying file %s to %s Error: %s", source, dest, err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatalf("Failed copying file %s to %s Error: %s", source, dest, err)
	}
}
