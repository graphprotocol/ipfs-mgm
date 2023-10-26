package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func createTempDirWithFile(f []string) (*os.File, error) {
	dir := f[:len(f)-1]
	// Create the directories for the CID structure
	err := os.MkdirAll(strings.Join(dir, "/"), 0755)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(fmt.Sprintf("%s", strings.Join(f, "/")))
	if err != nil {
		return nil, err
	}

	return file, nil
}

func GetCID(url string, payload io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, fmt.Errorf("Error creating HTTP request: %s", err)
	}

	// Set custom User-Agent for cloudflare WAF policies
	req.Header.Set("User-Agent", "graphprotocol/ipfs-mgm")

	// Create an HTTP client
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making API request: %s", err)
	}

	if s := res.Status; strings.HasPrefix(s, "5") || strings.HasPrefix(s, "4") {
		// Check if the error is due to the CID being a directory
		var dirIPFS IPFSErrorResponse
		_ = UnmarshalToStruct[IPFSErrorResponse](res.Body, &dirIPFS)
		if dirIPFS.Message == DIR_ERROR {
			return nil, fmt.Errorf("Cannot get this IPFS CID. Error message: %s", dirIPFS.Message)
		} else {
			return nil, fmt.Errorf("There was an error with the request. Error code: HTTP %s", s)
		}
	}

	return res, nil
}

func PostCID(dst string, payload []byte, fPath string) (*http.Response, error) {
	var tempFileName []string
	var base string
	if len(fPath) != 0 {
		tempFileName = strings.Split(fPath, "/")
		base = tempFileName[0]
		// Fix the nested directories
		if len(tempFileName) > 2 {
			tempFileName = tempFileName[1:]
			base = tempFileName[0]
		}
	} else {
		// Generate a unique temporary file name
		base = fmt.Sprintf("%d", time.Now().UnixNano())
		tempFileName = []string{base, "ipfs-data.tmp"}
	}

	// Create a temporary file to store the IPFS object data
	tempFile, err := createTempDirWithFile(tempFileName)
	if err != nil {
		return nil, fmt.Errorf("Error creating temporary file: %s", err)
	}
	defer tempFile.Close()

	// Write the IPFS object data to the temporary file
	_, err = tempFile.Write(payload)
	if err != nil {
		return nil, fmt.Errorf("Error writing data to temporary file: %s", err)
	}

	// Create a new HTTP POST request to add the file to the destination
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	filePart, err := writer.CreateFormFile("file", strings.Join(tempFileName, "/"))
	if err != nil {
		return nil, fmt.Errorf("Error creating form file: %s", err)
	}

	// Reset the temporary file pointer to the beginning
	tempFile.Seek(0, 0)

	// Copy the temporary file data into the form file
	_, err = io.Copy(filePart, tempFile)
	if err != nil {
		return nil, fmt.Errorf("Error copying file data: %s", err)
	}

	writer.Close() // Close the multipart writer

	req, err := http.NewRequest(http.MethodPost, dst, body)
	if err != nil {
		return nil, fmt.Errorf("There was an error creating the HTTP request: %s", err)
	}

	// Set custom User-Agent for cloudflare WAF policies
	req.Header.Set("User-Agent", "graphprotocol/ipfs-mgm")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set Directory Headers
	if len(fPath) != 0 {
		f := strings.Split(fPath, "/")
		fileNameParts := strings.Split(f[len(f)-1], ".")
		fileName := fileNameParts[0]
		_, a, _ := strings.Cut(fPath, "/")
		req.Header.Set("Content-Disposition", fmt.Sprintf("form-data; name=\"%s\"; filename=%s", fileName, url.PathEscape(a)))
		req.Header.Set("Abspath", fmt.Sprintf("%s", tempFileName))
	}
	defer os.RemoveAll(base)

	// Create an HTTP client
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making API request: %s", err)
	}

	if s := res.Status; strings.HasPrefix(s, "5") || strings.HasPrefix(s, "4") {
		return nil, fmt.Errorf("The endpoint responded with: HTTP %s", s)
	}

	return res, nil
}

func ParseHTTPBody(h *http.Response) ([]byte, error) {
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

func TestIPFSHash(s string, d string) error {
	if s != d {
		return fmt.Errorf("The source IPFS Hash is different from the destination Hash%s", "")
	}

	return nil
}

func PrintLogMessage(c int, l int, cid string, message string) {
	log.Printf("%d/%d (%s): %s", c, l, cid, message)
}

func SliceToCIDSStruct(s []string) ([]IPFSCIDResponse, error) {
	var cids []IPFSCIDResponse

	for _, k := range s {
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

func UnmarshalToStruct[V Data | IPFSResponse | IPFSCIDResponse | IPFSErrorResponse](h io.ReadCloser, m *V) error {
	scanner := bufio.NewScanner(h)

	for scanner.Scan() {
		err := json.Unmarshal(scanner.Bytes(), &m)
		if err != nil {
			return fmt.Errorf("Error Unmarshaling the structure: %s", err)
		}
	}

	return nil
}

func UnmarshalIPFSResponse(h io.ReadCloser, m *[]IPFSResponse) error {
	scanner := bufio.NewScanner(h)

	for scanner.Scan() {
		var rm IPFSResponse
		err := json.Unmarshal(scanner.Bytes(), &rm)
		if err != nil {
			return err
		}

		*m = append(*m, rm)
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
