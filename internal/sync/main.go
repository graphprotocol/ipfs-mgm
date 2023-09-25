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

	r := make(chan utils.HTTPResult)

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
		srcURL := src + utils.IPFS_LIST_ENDPOINT

		// Get the list of all CID's from the source IPFS
		go utils.PostIPFS(0, srcURL, nil, r)
		res := <- r
		if res.Error != nil {
			fmt.Println(res.Error)
		}

		scanner := bufio.NewScanner(res.HTTPResponse.Body)
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
	srcGet := src + utils.IPFS_CAT_ENDPOINT

	counter := 1
	length := len(cids)
	for _, k := range cids {
		// Get IPFS CID from source
		srcCID := srcGet + k.Cid
		log.Printf("%d/%d: Syncing the CID: %s\n",counter, length, k.Cid)

		// Get CID from source
		go utils.GetIPFS(counter, srcCID, nil, r)
		resC := <-r
		if resC.Error != nil {
			log.Printf("%d/%d: %s",resC.Counter, length, resC.Error)
			failed += 1
			counter += 1
			continue
		}
		defer resC.HTTPResponse.Body.Close()

		cidV := utils.GetCIDVersion(k.Cid)
		// Create the API URL fo the POST on destination
		apiADD := fmt.Sprintf("%s%s?cid-version=%s", dst, utils.IPFS_PIN_ENDPOINT, cidV)

		newBody, err := utils.GetHTTPBody(resC.HTTPResponse)
		if err != nil {
			log.Printf("%d/%d: %s",resC.Counter, length, err)
		}

		// Sync IPFS CID into destination
		// TODO: implement retry backoff with pester
		var m utils.IPFSResponse
		go utils.PostIPFS(resC.Counter, apiADD, newBody, r)
		resP := <- r
		if resP.Error != nil {
			log.Printf("%d/%d: %s", resP.Counter, length, resP.Error)
			failed += 1
		} else {
			defer resP.HTTPResponse.Body.Close()

			// Generic function to parse the response and create a struct
			err := utils.UnmarshalToStruct[utils.IPFSResponse](resC.HTTPResponse.Body, &m)
			if err != nil {
				log.Printf("%d/%d: %s", resP.Counter, length, err)
			}
		}

		// Check if the IPFS Hash is the same as the source one
		// If not the syncing didn't work
		ok, err := utils.TestIPFSHash(k.Cid, m.Hash)
		if err != nil {
			log.Printf("%d/%d: %s",resP.Counter, length, err)
			failed += 1
		} else {
			// Print success message
			log.Printf("%d/%d: %s",resP.Counter, length, ok)
			synced += 1
		}
		counter += 1
	}


	// Print Final statistics
	log.Printf("Total number of objects: %d; Synced: %d; Failed: %d\n", len(cids), synced, failed)
	log.Printf("Total time: %s\n", time.Since(timeStart))
}

