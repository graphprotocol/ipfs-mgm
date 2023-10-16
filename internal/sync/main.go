package sync

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
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

var workerItemCount int = 50

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

	// Get all command flags
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
		listURL := fmt.Sprintf("%s%s", src, utils.PIN_LIST_ENDPOINT)

		// Get the list of all CID's from the source IPFS
		// TODO: implement retry backoff with pester
		resL, err := utils.PostCID(listURL, nil, "")
		if err != nil {
			fmt.Println(err)
		}
		defer resL.Body.Close()

		// Create the slice with the CIDS's
		scanner := bufio.NewScanner(resL.Body)
		for scanner.Scan() {
			var j utils.IPFSCIDResponse
			err := json.Unmarshal(scanner.Bytes(), &j)
			if err != nil {
				fmt.Printf("Error unmarshaling the response: %s", err)
			}
			cids = append(cids, j)
		}
	}

	counter := 1
	length := len(cids)

	// Adjust for the number of CID's
	if length < workerItemCount {
		workerItemCount = length
	}

	for i := 0; i < length; {
		// Create a channel with buffer of workerItemCount size
		workChan := make(chan utils.HTTPResult, workerItemCount)
		var wg sync.WaitGroup

		for j := 0; j < workerItemCount; j++ {
			wg.Add(1)
			go func(c int, cidID string) {
				defer wg.Done()
				AsyncCall(src, dst, cidID, &c, length, &failed, &synced)

			}(counter, cids[i].Cid)
			counter += 1
			i++
		}

		close(workChan)
		wg.Wait()
	}

	// Print Final statistics
	log.Printf("Total number of objects: %d; Synced: %d; Failed: %d\n", len(cids), synced, failed)
	log.Printf("Total time: %s\n", time.Since(timeStart))
}

func AsyncCall(src string, dst string, cidID string, counter *int, length int, failed *int, synced *int) {
	// Create the API URL for the IPFS GET
	srcGet := fmt.Sprintf("%s%s%s", src, utils.CAT_ENDPOINT, cidID)

	utils.PrintLogMessage(*counter, length, cidID, "Syncing")

	// Get CID from source
	resG, err := utils.GetCID(srcGet, nil)
	if err != nil {
		// Check if it's a directory
		if strings.Contains(fmt.Sprintf("%s", err), utils.DIR_ERROR) {
			err := syncDir(src, dst, cidID, cidID)
			if err != nil {
				utils.PrintLogMessage(*counter, length, cidID, fmt.Sprintf("%s", err))
				*failed += 1
				*counter += 1
			} else {
				utils.PrintLogMessage(*counter, length, cidID, "Successfully synced directory")
			}
		} else {
			utils.PrintLogMessage(*counter, length, cidID, fmt.Sprintf("%s", err))
			*failed += 1
			*counter += 1
		}
		return
	}
	defer resG.Body.Close()

	payload, err := utils.ParseHTTPBody(resG)
	if err != nil {
		utils.PrintLogMessage(*counter, length, cidID, fmt.Sprintf("%s", err))
	}

	err = syncCall(src, dst, cidID, "", "", payload)
	if err != nil {
		utils.PrintLogMessage(*counter, length, cidID, fmt.Sprintf("%s", err))
		*failed += 1
	}

	// Print success message
	utils.PrintLogMessage(*counter, length, cidID, "Successfully synced")
	*synced += 1
}

func syncCall(src, dst, cid, parentCid, filePath string, payload []byte) error {
	// We need to get the body if this was a fresh call
	if len(payload) == 0 {
		// Create the API URL for the IPFS GET
		srcGet := fmt.Sprintf("%s%s%s", src, utils.CAT_ENDPOINT, cid)

		// Get CID from source
		resG, err := utils.GetCID(srcGet, nil)
		if err != nil {
			return err
		}
		defer resG.Body.Close()

		payload, err = utils.ParseHTTPBody(resG)
		if err != nil {
			return err
		}
	}
	cidV := utils.GetCIDVersion(cid)

	var apiADD string
	if len(filePath) != 0 {
		// Create the API URL for the directory POST on destination
		apiADD = fmt.Sprintf("%s%s?cid-version=%s&wrap-with-directory=1&to-files=1", dst, utils.IPFS_PIN_ENDPOINT, cidV)
	} else {
		// Create the API URL for the POST on destination
		apiADD = fmt.Sprintf("%s%s?cid-version=%s", dst, utils.IPFS_PIN_ENDPOINT, cidV)
	}

	// Sync IPFS CID into destination
	// TODO: implement retry backoff with pester
	// log.Printf(filePath)
	resP, err := utils.PostCID(apiADD, payload, filePath)
	if err != nil {
		return err
	}
	defer resP.Body.Close()

	// Generic function to parse the response and create a struct
	var m []utils.IPFSResponse
	err = utils.UnmarshalIPFSResponse(resP.Body, &m)
	if err != nil {
		return err
	}

	// Check if the IPFS Hash is the same as the source one
	// If not the syncing didn't work
	ok := false
	for _, v := range m {
		if len(parentCid) != 0 {
			if v.Hash == parentCid {
				ok = true
				break
			}
		} else {
			if v.Hash == cid {
				ok = true
			}
		}

	}

	if !ok {
		return fmt.Errorf("Can't be synced. The source and destination IPFS Hash differ")
	}

	return nil
}

func syncDir(src, dst, file, parentCid string) error {
	listURL := fmt.Sprintf("%s%s%s", src, utils.DIR_LIST_ENDPOINT, file)

	// List directory
	lsD, err := utils.GetCID(listURL, nil)
	if err != nil {
		return err
	}
	defer lsD.Body.Close()

	// Create the structure with the CID directory
	var data utils.Data
	err = utils.UnmarshalToStruct[utils.Data](lsD.Body, &data)
	if err != nil {
		return err
	}

	// Recursive function to sync all directory content
	for _, v := range data.Objects {
		err = syncDirContent(src, dst, parentCid, v, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncDirContent(src, dst, parentCID string, data utils.Object, s bool) error {
	for _, v := range data.Links {
		// Syntax: https://ipfs.com/ipfs/api/v0/cat?arg=QmcoBTSpxyBx2AuUqhuy5X1UrasbLoz76QFGLgqUqhXLK6/foo.txt
		filePath := fmt.Sprintf("%s/%s", data.Hash, v.Name)
		url := fmt.Sprintf("%s%s%s", src, utils.CAT_ENDPOINT, filePath)

		_, err := utils.GetCID(url, nil)
		if err != nil {
			// Check if it's a directory
			// If true, the new source will be like: https://ipfs.com/ipfs/api/v0/cat?arg=QmcoBTSpxyBx2AuUqhuy5X1UrasbLoz76QFGLgqUqhXLK6/FOO
			if strings.Contains(fmt.Sprintf("%s", err), utils.DIR_ERROR) {
				// The new CID for directory will be like: QmcoBTSpxyBx2AuUqhuy5X1UrasbLoz76QFGLgqUqhXLK6/FOO
				filePath := fmt.Sprintf("%s/%s", data.Hash, v.Name)
				err := syncDir(src, dst, filePath, v.Hash)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			err = syncCall(src, dst, v.Hash, parentCID, filePath, []byte{})
			if err != nil {
				return err
			}
		}

	}

	return nil
}
