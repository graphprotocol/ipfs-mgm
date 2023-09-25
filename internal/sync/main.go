package sync

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	SyncCmd.Flags().StringP("from-file", "f", "", "Sync CID's from file")
}

func Sync(cmd *cobra.Command) {
	timeStart := time.Now()
	failed := 0
	synced := 0

	var cids []utils.IPFSCIDResponse

	src, err := cmd.Flags().GetString("source")
	if err != nil {
		log.Println(err)
	}

	dst, err := cmd.Flags().GetString("destination")
	if err != nil {
		log.Println(err)
	}

	fromFile, err := cmd.Flags().GetString("from-file")
	if err != nil {
		fmt.Println(err)
	}

	// Will use the file only if specified
	if len(fromFile) > 0 {
		log.Printf("Syncing from %s to %s using the file <%s> as input\n", src, dst, fromFile)
		c, err := utils.ReadCIDFromFile(fromFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Create our structure with the CIDS's
		cids, err = utils.SliceToCIDSStruct(c)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		log.Printf("Syncing from %s to %s\n", src, dst)

		// Create the API URL for the IPFS pin/ls operation
		apiURL := src + utils.IPFS_LIST_ENDPOINT

		// Get the HTTP response
		res, err := utils.PostIPFS(apiURL, nil)
		if err != nil {
			fmt.Println(err)
		}

		// // Get the HTTP body
		// body, _ := utils.GetHTTPBody(res)

		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			var j utils.IPFSCIDResponse
			err := json.Unmarshal(scanner.Bytes(), &j)
			if err != nil {
				fmt.Printf("Error unmarshaling the response: %s", err)
			}
			cids = append(cids, j)
		}
	}

	// Create the API URL for the IPFS GET
	apiGet := src + utils.IPFS_CAT_ENDPOINT

	counter := 1
	length := len(cids)
	for _, k := range cids {
		// Get IPFS CID from source
		apiCID := apiGet + k.Cid
		log.Printf("%d/%d: Syncing the CID: %s\n",counter, length, k.Cid)

		// Get CID from source
		cid, err := utils.GetIPFS(apiCID, nil)
		if err != nil {
			log.Printf("%d/%d: %s",counter, length, err)
			failed += 1
			continue
		}
		defer cid.Body.Close()

		cidV := utils.GetCIDVersion(k.Cid)
		// Create the API URL fo the POST on destination
		apiADD := fmt.Sprintf("%s%s?cid-version=%s", dst, utils.IPFS_PIN_ENDPOINT, cidV)

		newBody, err := utils.GetHTTPBody(cid)
		if err != nil {
			log.Printf("%d/%d: %s",counter, length, err)
		}

		// Sync IPFS CID into destination
		var r utils.IPFSResponse
		add, err := utils.PostIPFS(apiADD, newBody)
		if err != nil {
			log.Printf("%d/%d: %s",counter, length, err)
			failed += 1
		} else {
			defer add.Body.Close()

			// Generic function to parse the response and create a struct
			err := utils.UnmarshalToStruct[utils.IPFSResponse](add.Body, &r)
			if err != nil {
				log.Printf("%d/%d: %s",counter, length, err)
			}
		}

		// Check if the IPFS Hash is the same as the source one
		// If not the syncing didn't work
		_, err = utils.TestIPFSHash(k.Cid, r.Hash)
		if err != nil {
			log.Printf("%d/%d: %s",counter, length, err)
			failed += 1
		} else {
			// Print success message
			synced += 1
		}
		counter += 1
	}


	// Print Final statistics
	log.Printf("Total number of objects: %d; Synced: %d; Failed: %d\n", len(cids), synced, failed)
	log.Printf("Total time: %s\n", time.Since(timeStart))
}

// func getFromFile(url string, f string) {
// 	log.Println("Syncing from file")
//
// 	cids, err := utils.ReadCDIFromFile(f)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	log.Println(cids)
//
// 	for _, v := range cids {
// 		log.Println(v)
// 		// Create the API URL for the IPFS GET operation without the CID ID
// 		apiGet := url + utils.IPFS_CAT_ENDPOINT + v
// 	}
// 	// // Create the API URL for the IPFS GET request
// 	// apiUrl := url + utils.IPFS_CAT_ENDPOINT
// 	//
// 	// f, err := os.Open(f)
// 	// if err != nil {
// 	// 	log.Println(err)
// 	// }
// 	//
// 	// // Make GET HTTP request
// 	// res, err := utils.GetIPFS(apiUrl)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// }
// 	//
// 	// //Get the HTTP body
// 	// body, _ := utils.GetHTTPBody(res)
// 	//
// 	// fmt.Println(string(body))
// 	// os.Exit(0)
//
// 	// Print Final statistics
// 	log.Printf("Total number of objects: %d; Synced: %d; Failed: %d\n", len(sliceIPFS), synced, failed)
// 	log.Printf("Total time: %s\n", time.Since(timeStart))
// }
