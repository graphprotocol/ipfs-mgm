package utils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ValidateEndpoints (src string, dst string) error {
    if src == dst {
        return fmt.Errorf("Error: The specified source <%s> is the same as the destination <%s>", src, dst)
    }
    return nil
}

// GenerateTempFileName generates a temporary file name.
func GenerateTempFileName(prefix, suffix string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%s", prefix, time.Now().UnixNano(), suffix))
}

func GetIPFS(c int, url string, payload io.Reader, r chan HTTPResult) {
    req, err := http.NewRequest(http.MethodGet, url, payload)
    if err != nil {
	r <- HTTPResult{
	    HTTPResponse: nil,
	    Error: fmt.Errorf("Error creating HTTP request: %s", err),
	}
    }

    // Set custom User-Agent for cloudflare WAF policies
    req.Header.Set("User-Agent", "graphprotocol/ipfs-mgm")

    // Create an HTTP client
    client := &http.Client{}

    res, err := client.Do(req)
    if err != nil {
	r <- HTTPResult{
	    HTTPResponse: nil,
	    Error: fmt.Errorf("Error making API request: %s", err),
	}
    }

    if s := res.Status; strings.HasPrefix(s, "5") || strings.HasPrefix(s, "4") {
        // Check if the error is due to the CID being a directory
        var dirIPFS IPFSErrorResponse
        _ = UnmarshalToStruct[IPFSErrorResponse](res.Body, &dirIPFS)
        if dirIPFS.Message == IPFS_DIR_ERROR {
	    r <- HTTPResult{
		HTTPResponse: nil,
		Error: fmt.Errorf("Cannot get this IPFS CID. Error message: %s", dirIPFS.Message),
		Counter: c,
	    }
        } else {
	    r <- HTTPResult{
		HTTPResponse: nil,
		Error: fmt.Errorf("There was an error with the request. Error code: HTTP %s", s),
	    }
        }
    }

    r <- HTTPResult{
	HTTPResponse: res,
	Error: nil,
	Counter: c,
    }
}

func PostIPFS(c int, url string, payload []byte, r chan HTTPResult) {
    // Generate a unique temporary file name
    tempFileName := GenerateTempFileName("ipfs-data-", ".tmp")

    // Create a temporary file to store the IPFS object data
    tempFile, err := os.Create(tempFileName)
    if err != nil {
	r <- HTTPResult{
	    HTTPResponse: nil,
	    Error: fmt.Errorf("Error creating temporary file: %s", err),
	    Counter: c,
	}
    }
    defer tempFile.Close()

    // Write the IPFS object data to the temporary file
    _, err = tempFile.Write(payload)
    if err != nil {
	r <- HTTPResult{
	    HTTPResponse: nil,
	    Error: fmt.Errorf("Error writing data to temporary file: %s", err),
	}
    }

    // Create a new HTTP POST request to add the file to the destination
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    filePart, err := writer.CreateFormFile("file", filepath.Base(tempFileName))
    if err != nil {
	r <- HTTPResult{
	    HTTPResponse: nil,
	    Error: fmt.Errorf("Error creating form file: %s", err),
	    Counter: c,
	}
    }

    // Reset the temporary file pointer to the beginning
    tempFile.Seek(0, 0)

    // Copy the temporary file data into the form file
    _, err = io.Copy(filePart, tempFile)
    if err != nil {
	r <- HTTPResult{
	    HTTPResponse: nil,
	    Error: fmt.Errorf("Error copying file data: %s", err),
	    Counter: c,
	}
    }

    writer.Close() // Close the multipart writer

    req, err := http.NewRequest(http.MethodPost, url, body)
    if err != nil {
	r <- HTTPResult{
	    HTTPResponse: nil,
	    Error: fmt.Errorf("There was an error creating the HTTP request: %s", err),
	    Counter: c,
	}
    }

    // Set custom User-Agent for cloudflare WAF policies
    req.Header.Set("User-Agent", "graphprotocol/ipfs-mgm")
    // req.Header.Set("Content-Type", "text/plain")
    req.Header.Set("Content-Type", writer.FormDataContentType())

    // Create an HTTP client
    client := &http.Client{}

    res, err := client.Do(req)
    if err != nil {
	r <- HTTPResult{
	    HTTPResponse: nil,
	    Error: fmt.Errorf("Error making API request: %s", err),
	    Counter: c,
	}
    }

    if s := res.Status; strings.HasPrefix(s, "5") || strings.HasPrefix(s, "4") {
	r <- HTTPResult{
	    HTTPResponse: nil,
	    Error: fmt.Errorf("The endpoint responded with: HTTP %s", s),
	    Counter: c,
	}
    }

    r <- HTTPResult{
	HTTPResponse: res,
	Error: nil,
	Counter: c,
    }
}

func GetHTTPBody(h *http.Response) ([]byte, error) {
    // Read the body response
    body, err := ioutil.ReadAll(h.Body)
    if err != nil {
        return nil, fmt.Errorf("Error reading response body: %s", err)
    }

    return body, nil
}

func GetCIDVersion(cid string) string {
    if strings.HasPrefix(cid, "Qm") {
        return "0"
    }

    return "1"
}

func TestIPFSHash(s string, d string) (string, error) {
    if s != d {
        return "", fmt.Errorf("The source Hash %s is different from the destination hash %s", s, d)
    }

    return fmt.Sprint("Successfully synced to destination IPFS"), nil
}

func SliceToCIDSStruct(s []string) ([]IPFSCIDResponse, error) {
    var cids []IPFSCIDResponse

    for _, k := range s{
        var cid IPFSCIDResponse
        // create the structure to be unmarshaled from our string
        a := fmt.Sprintf(`{"cid":"%s"}`, k)
        err := json.Unmarshal([]byte(a), &cid)
        if err != nil {
            return nil, fmt.Errorf("Error unmarshaling from slice to IPFS Struct: %s", err)
        }
        cids = append(cids, cid)
    }
    return cids, nil
}

func UnmarshalToStruct[V IPFSResponse | IPFSCIDResponse | IPFSErrorResponse](h io.ReadCloser, m *V) error {
	scanner := bufio.NewScanner(h)
	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &m)
		if err != nil {
			return fmt.Errorf("Error Unmarshaling the structure: %s", err)
		}
	}

	return nil
}

func ReadCIDFromFile(f string) ([]string, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, fmt.Errorf("Error opening the file <%s>", f)
	}
	defer file.Close()

	var s []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s = append(s, scanner.Text())
	}

    return s, nil
}

