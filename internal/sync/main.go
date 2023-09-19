package sync

import (
	"fmt"
	"encoding/json"
	// "strings"
	"time"

	"github.com/graphprotocol/ipfs-mgm/internal/utils"
	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync IPFS objects",
	Long:  `Sync objects between two different IPFS endpoints`,
	Run: func(cmd *cobra.Command, args []string) {
		Sync(cmd)
	},
}

func init() {
	SyncCmd.Flags().StringP("source", "s", "", "IPFS source endpoint")
	SyncCmd.MarkFlagRequired("source")
	SyncCmd.Flags().StringP("destination", "d", "", "IPFS destination endpoint")
	SyncCmd.MarkFlagRequired("destination")
	SyncCmd.Flags().StringP("from-file", "f", "", "Sync CID's from file. Has priority over <source> and <destination> flags")
}

func Sync(cmd *cobra.Command) {
	timeStart := time.Now()


	src, err := cmd.Flags().GetString("source")
	if err != nil {
		fmt.Println(err)
	}

	// fromFile, err := cmd.Flags().GetString("from-file")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// if len(fromFile) != 0 {
	// 	syncFromFile(src, fromFile)
	// }

	dst, err := cmd.Flags().GetString("destination")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Syncing from %s to %s\n", src, dst)

	// Create the API URL for the IPFS pin/ls operation
	apiURL := src + utils.IPFS_LIST_ENDPOINT

	// Get the HTTP response
	res := utils.PostIPFS(apiURL, nil)

	// Get the HTTP body
	body, _ := utils.GetHTTPBody(res)

	m := make(map[string]interface{})
	err = json.Unmarshal(body, &m)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(m)

	// newBody := strings.Split(string(body), "\n")
	//
	// for _, cid := range newBody {
	// 	fmt.Printf("CID: %v\n", cid)
	// 	fmt.Println(reflect.TypeOf(cid))
	// }

	// Print Final statistics
	fmt.Printf("Total number of objects: %d\n", len(body))
	fmt.Printf("Total time: %s\n", time.Since(timeStart))
}

// func syncFromFile(url string, f string) {
// 	fmt.Println(f)
//
// 	// Create the API URL for the IPFS GET request
// 	apiUrl := url + utils.IPFS_CAT_ENDPOINT
//
// 	f, err := os.Open(f)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
//
// 	// Make GET HTTP request
// 	res := utils.GetIPFS(apiUrl)
//
// 	//Get the HTTP body
// 	body, _ := utils.GetHTTPBody(res)
//
// 	fmt.Println(string(body))
// 	os.Exit(0)
// }
