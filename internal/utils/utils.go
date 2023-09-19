package utils

import (
    "fmt"
    "io"
    "io/ioutil"
    "strings"
    "os"

    "net/http"
)

var IPFS_LIST_ENDPOINT string = "/ipfs/api/v0/pin/ls?stream=true"
var IPFS_CAT_ENDPOINT string = "/ipfs/api/v0/cat?arg="
var HEADER_APP_JSON string = "application/json"

type IPFSStruct struct {
    Cid string `json:"cid"`
    Type string `json:"type"`
}

func ValidateEndpoints (src string, dst string) error {
    if src == dst {
        return fmt.Errorf("Error: The specified source <%s> is the same as the destination <%s>\n", src, dst)
    }
    return nil
}

func GetIPFS(url string) *http.Response {
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    // Set custom User-Agent for cloudflare WAF policies
    req.Header.Set("User-Agent", "graphprotocol/ipfs-mgm")

    // Create an HTTP client
    client := &http.Client{}

    res, err := client.Do(req)
    if err != nil {
        fmt.Println("Error making API request:", err)
        os.Exit(1)
    }

    if s := res.Status; strings.HasPrefix(s, "5") || strings.HasPrefix(s, "4") {
        fmt.Printf("There was an error with the request. Error code: HTTP %s\n", s)
        os.Exit(1)
    }

    return res
}

func PostIPFS(url string, payload io.Reader) *http.Response {

    req, err := http.NewRequest(http.MethodPost, url, payload)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    // Set custom User-Agent for cloudflare WAF policies
    req.Header.Set("User-Agent", "graphprotocol/ipfs-mgm")

    // Create an HTTP client
    client := &http.Client{}

    res, err := client.Do(req)
    if err != nil {
        fmt.Println("Error making API request:", err)
        os.Exit(1)
    }

    if s := res.Status; strings.HasPrefix(s, "5") || strings.HasPrefix(s, "4") {
        fmt.Printf("There was an error with the request. Error code: HTTP %s\n", s)
        os.Exit(1)
    }

    return res
}

func GetHTTPBody(h *http.Response) ([]byte, error) {
    // Read the body response
    body, err := ioutil.ReadAll(h.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        os.Exit(1)
    }
    defer h.Body.Close()

    return body, err
}
