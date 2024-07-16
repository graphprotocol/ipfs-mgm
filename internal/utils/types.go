package utils

import "net/http"

var GATEWAY_TIMEOUT_HEADERS int8 = 120

var DIR_LIST_ENDPOINT string = "/ipfs/api/v0/ls?arg="
var PIN_LIST_ENDPOINT string = "/ipfs/api/v0/pin/ls?stream=true"
var CAT_ENDPOINT string = "/ipfs/api/v0/cat?arg="
var IPFS_PIN_ENDPOINT string = "/ipfs/api/v0/add"
var HEADER_APP_JSON string = "application/json"

var DIR_ERROR = "this dag node is a directory"

type Link struct {
	Name   string `json:"Name"`
	Hash   string `json:"Hash"`
	Size   int    `json:"Size"`
	Type   int    `json:"Type"`
	Target string `json:"Target"`
}

type Object struct {
	Hash  string `json:"Hash"`
	Links []Link `json:"Links"`
}

type Data struct {
	Objects []Object `json:"Objects"`
}

type HTTPResult struct {
	HTTPResponse *http.Response `json:"http_response"`
	Error        error          `json:"error"`
	Counter      int            `json:"counter" default:"0"`
}

type IPFSCIDResponse struct {
	Cid string `json:"cid"`
}

type IPFSResponse struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Size string `json:"size"`
}

type IPFSErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Type    string `json:"type"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
