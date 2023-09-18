package utils

import (
    "fmt"
    "io"
    "io/ioutil"
    "strings"
    "time"

    "os"

    "net/http"
)

var IPFS_LIST_ENDPOINT string = "/ipfs/api/v0/pin/ls?stream-channels=true"
var HEADER_APP_JSON string = "application/json"

func ValidateEndpoints (src string, dst string) error {
    if src == dst {
        return fmt.Errorf("Error: The specified source <%s> is the same as the destination <%s>\n", src, dst)
    }
    return nil
}

func postListUrl(url string, payload io.Reader) *http.Response {
    // Create an HTTP client
    client := &http.Client{
        Timeout: 10 * time.Minute,
    }

    // Create the URL for the IPFS LIST
    apiURL := url + IPFS_LIST_ENDPOINT

    fmt.Printf("Making request to: %s\n", apiURL)

    res, err := client.Post(apiURL, HEADER_APP_JSON, payload)
    if err != nil {
        fmt.Println("Error making API request:", err)
        os.Exit(1)
    }

    if s := res.Status; strings.HasPrefix(s, "5") || strings.HasPrefix(s, "4") {
        fmt.Printf("There was an error with the request. Error code: HTTP %s\n", s)
        os.Exit(1)
    }

    fmt.Println(res.Status)

    fmt.Println(res)

    defer res.Body.Close()

    return res
}

func IPFSPost(src string) {
    // Make POST to IPFS list endpoint
    res := postListUrl(src, nil)

    // Read the body response
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        os.Exit(1)
    }

    // Print all IPFS content
    fmt.Println("Files and directories in remote IPFS root:")
    fmt.Println(string(body))
}
