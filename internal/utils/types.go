package utils

var IPFS_LIST_ENDPOINT string = "/ipfs/api/v0/pin/ls?stream=true"
var IPFS_CAT_ENDPOINT string = "/ipfs/api/v0/cat?arg="
// var IPFS_PIN_ENDPOINT string = "/ipfs/api/v0/add?stream-channels=true"
var IPFS_PIN_ENDPOINT string = "/ipfs/api/v0/add"
var HEADER_APP_JSON string = "application/json"

var IPFS_DIR_ERROR = "this dag node is a directory"

type IPFSCIDResponse struct {
    Cid     string `json:"cid"`
    Type    string `json:"type"`
}

type IPFSResponse struct {
    Name    string `json:"name"`
    Hash    string `json:"hash"`
    Size    string `json:"size"`
}

type IPFSErrorResponse struct {
    Message string  `json:"message"`
    Code    int	    `json:"code"`
    Type    string  `json:"type"`
}
