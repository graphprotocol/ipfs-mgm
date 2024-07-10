package diff

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/graphprotocol/ipfs-mgm/internal/utils"
	"github.com/spf13/cobra"
)

var DiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff two IPFS node objects",
	Long:  `Diff two IPFS node objects and report the missing objects in destination`,
	Run: func(cmd *cobra.Command, args []string) {
		Diff(cmd)
	},
}

func init() {
	DiffCmd.Flags().StringP("source", "s", "", "IPFS source endpoint")
	DiffCmd.MarkFlagRequired("source")
	DiffCmd.Flags().StringP("destination", "d", "", "IPFS destination endpoint")
	DiffCmd.MarkFlagRequired("destination")
	DiffCmd.Flags().StringP("from-file", "f", "", "Diff CID's from file")
}

func Diff(cmd *cobra.Command) {
	// timeStart := time.Now()
	// failed := 0
	// Diffed := 0

	var srcCIDs []utils.IPFSCIDResponse
	var dstCIDs []utils.IPFSCIDResponse

	// check if Diffing only the CIDS specified in the file
	fromFile, err := cmd.Flags().GetString("from-file")
	if err != nil {
		log.Println(err)
	}

	// get source to Diff from
	src, err := cmd.Flags().GetString("source")
	if err != nil {
		log.Println(err)
	}

	dst, err := cmd.Flags().GetString("destination")
	if err != nil {
		log.Println(err)
	}

	// Will use the file only if specified
	if len(fromFile) > 0 {
		log.Printf("Diffing from <%s> to <%s> using as input the file <%s>\n", src, dst, fromFile)
		c, err := utils.ReadCIDFromFile(fromFile)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		// Create our structure with the CID's
		srcCIDs, err = utils.SliceToCIDSStruct(c)
		if err != nil {
			log.Println(err)
		}
		// // Create the slice with the CIDS's
		// err = utils.UnmarshalIPFSResponse[utils.IPFSCIDResponse](c, srcCIDs)
		// if err != nil {
		// 	log.Println(err)
		// }
	} else {
		log.Printf("Comparing source: %s with destination: %s\n", src, dst)

		dstUrl := fmt.Sprintf("%s%s", dst, utils.PIN_LIST_ENDPOINT)
		srcUrl := fmt.Sprintf("%s%s", src, utils.PIN_LIST_ENDPOINT)

		str := [2]utils.DiffStruct{
			{
				Url: dstUrl, Cids: dstCIDs,
			},
			{
				Url: srcUrl, Cids: srcCIDs,
			},
		}

		// get the list of IPFS objects in an async mode
		var wg sync.WaitGroup

		for i := 0; i < len(str); i++ {
			wg.Add(1)

			go func(url string, r []utils.IPFSCIDResponse) {
				defer wg.Done()
				log.Printf("getting the CID's from <%s>\n", url)

				res, err := utils.PostCall(url, nil, "")
				if err != nil {
					log.Println(err)
				}

				// getting source CID's
				err = utils.UnmarshalIPFSResponse[utils.IPFSCIDResponse](string(res), r)
				if err != nil {
					log.Printf("%s\n", err)
				}
				fmt.Println(r)
			}(str[i].Url, str[i].Cids)
		}

		wg.Wait()

		fmt.Println(len(srcCIDs))

		// Create the API URL for the IPFS pin/ls operation

		for _, k := range srcCIDs {
			fmt.Println(k)
		}
		// // Create the slice with the CID's
		// scanner := bufio.NewScanner(resL.Body)
		// for scanner.Scan() {
		// 	var j utils.IPFSCIDResponse
		// 	err := json.Unmarshal(scanner.Bytes(), &j)
		// 	if err != nil {
		// 		log.Printf("Error unmarshaling the response: %s", err)
		// 	}
		// 	cids = append(cids, j)
		// }
	}

	// counter := 1

	// length := len(cids)
	// log.Printf("There are %d CIDs to be Diffed", length)

	// for i := 0; i < length; {
	// 	// Create a channel with buffer of workerItemCount size
	// 	workChan := make(chan utils.HTTPResult, batch)
	// 	var wg Diff.WaitGroup

	// 	for j := 0; j < batch; j++ {
	// 		wg.Add(1)
	// 		go func(c int, cidID string) {
	// 			defer wg.Done()
	// 			ADiffCall(src, dst, cidID, &c, length, &failed, &Diffed)
	// 		}(counter, cids[i].Cid)
	// 		counter += 1
	// 		i++
	// 	}

	// 	if cooldown > 0 {
	// 		time.Sleep(time.Duration(cooldown) * time.Second)
	// 	}

	// 	close(workChan)
	// 	wg.Wait()
	// }

	// // Print Final statistics
	// log.Printf("Total number of objects: %d; Diffed: %d; Failed: %d\n", len(cids), Diffed, failed)
	// log.Printf("Total time: %s\n", time.Since(timeStart))
}

func getIPFSObjects(dst string, r []utils.IPFSCIDResponse) {
	url := fmt.Sprintf("%s%s", dst, utils.PIN_LIST_ENDPOINT)
	log.Printf("getting the CID's from <%s>\n", url)

	res, err := utils.PostCall(url, nil, "")
	if err != nil {
		log.Println(err)
	}

	// getting source CID's
	err = utils.UnmarshalIPFSResponse[utils.IPFSCIDResponse](string(res), r)
	if err != nil {
		log.Printf("%s\n", err)
	}
}

// func ADiffCall(src string, dst string, cidID string, counter *int, length int, failed *int, Diffed *int) {
// 	// Create the API URL for the IPFS GET
// 	srcGet := fmt.Sprintf("%s%s%s", src, utils.CAT_ENDPOINT, cidID)

// 	utils.PrintLogMessage(*counter, length, cidID, "Diffing")

// 	// Get CID from source
// 	resG, err := utils.GetCID(srcGet, nil)
// 	if err != nil {
// 		// Check if it's a directory
// 		if strings.Contains(fmt.Sprintf("%s", err), utils.DIR_ERROR) {
// 			err := DiffDir(src, dst, cidID, cidID)
// 			if err != nil {
// 				utils.PrintLogMessage(*counter, length, cidID, fmt.Sprintf("%s", err))
// 				*failed += 1
// 				*counter += 1
// 			} else {
// 				utils.PrintLogMessage(*counter, length, cidID, "Successfully Diffed directory")
// 			}
// 		} else {
// 			utils.PrintLogMessage(*counter, length, cidID, fmt.Sprintf("%s", err))
// 			*failed += 1
// 			*counter += 1
// 		}
// 		return
// 	}
// 	defer resG.Body.Close()

// 	payload, err := utils.ParseHTTPBody(resG)
// 	if err != nil {
// 		utils.PrintLogMessage(*counter, length, cidID, fmt.Sprintf("%s", err))
// 	}

// 	err = DiffCall(src, dst, cidID, "", "", payload)
// 	if err != nil {
// 		utils.PrintLogMessage(*counter, length, cidID, fmt.Sprintf("%s", err))
// 		*failed += 1
// 	}

// 	// Print success message
// 	utils.PrintLogMessage(*counter, length, cidID, "Successfully Diffed")
// 	*Diffed += 1
// }

// func DiffCall(src, dst, cid, parentCid, filePath string, payload []byte) error {
// 	// We need to get the body if this was a fresh call
// 	if len(payload) == 0 {
// 		// Create the API URL for the IPFS GET
// 		srcGet := fmt.Sprintf("%s%s%s", src, utils.CAT_ENDPOINT, cid)

// 		// Get CID from source
// 		resG, err := utils.GetCID(srcGet, nil)
// 		if err != nil {
// 			return err
// 		}
// 		defer resG.Body.Close()

// 		payload, err = utils.ParseHTTPBody(resG)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	cidV := utils.GetCIDVersion(cid)

// 	var apiADD string
// 	if len(filePath) != 0 {
// 		// Create the API URL for the directory POST on destination
// 		apiADD = fmt.Sprintf("%s%s?cid-version=%s&wrap-with-directory=1&to-files=1", dst, utils.IPFS_PIN_ENDPOINT, cidV)
// 	} else {
// 		// Create the API URL for the POST on destination
// 		apiADD = fmt.Sprintf("%s%s?cid-version=%s", dst, utils.IPFS_PIN_ENDPOINT, cidV)
// 	}

// 	// Diff IPFS CID into destination
// 	// TODO: implement retry backoff with pester
// 	// log.Printf(filePath)
// 	resP, err := utils.PostCID(apiADD, payload, filePath)
// 	if err != nil {
// 		return err
// 	}
// 	defer resP.Body.Close()

// 	// Generic function to parse the response and create a struct
// 	var m []utils.IPFSResponse
// 	err = utils.UnmarshalIPFSResponse(resP.Body, &m)
// 	if err != nil {
// 		return err
// 	}

// 	// Check if the IPFS Hash is the same as the source one
// 	// If not the Diffing didn't work
// 	ok := false
// 	for _, v := range m {
// 		if len(parentCid) != 0 {
// 			if v.Hash == parentCid {
// 				ok = true
// 				break
// 			}
// 		} else {
// 			if v.Hash == cid {
// 				ok = true
// 			}
// 		}

// 	}

// 	if !ok {
// 		return fmt.Errorf("Can't be Diffed. The source and destination IPFS Hash differ")
// 	}

// 	return nil
// }

// func DiffDir(src, dst, file, parentCid string) error {
// 	listURL := fmt.Sprintf("%s%s%s", src, utils.DIR_LIST_ENDPOINT, file)

// 	// List directory
// 	lsD, err := utils.GetCID(listURL, nil)
// 	if err != nil {
// 		return err
// 	}
// 	defer lsD.Body.Close()

// 	// Create the structure with the CID directory
// 	var data utils.Data
// 	err = utils.UnmarshalToStruct[utils.Data](lsD.Body, &data)
// 	if err != nil {
// 		return err
// 	}

// 	// Recursive function to Diff all directory content
// 	for _, v := range data.Objects {
// 		err = DiffDirContent(src, dst, parentCid, v, true)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func DiffDirContent(src, dst, parentCID string, data utils.Object, s bool) error {
// 	for _, v := range data.Links {
// 		// Syntax: https://ipfs.com/ipfs/api/v0/cat?arg=QmcoBTSpxyBx2AuUqhuy5X1UrasbLoz76QFGLgqUqhXLK6/foo.txt
// 		filePath := fmt.Sprintf("%s/%s", data.Hash, v.Name)
// 		url := fmt.Sprintf("%s%s%s", src, utils.CAT_ENDPOINT, filePath)

// 		_, err := utils.GetCID(url, nil)
// 		if err != nil {
// 			// Check if it's a directory
// 			// If true, the new source will be like: https://ipfs.com/ipfs/api/v0/cat?arg=QmcoBTSpxyBx2AuUqhuy5X1UrasbLoz76QFGLgqUqhXLK6/FOO
// 			if strings.Contains(fmt.Sprintf("%s", err), utils.DIR_ERROR) {
// 				// The new CID for directory will be like: QmcoBTSpxyBx2AuUqhuy5X1UrasbLoz76QFGLgqUqhXLK6/FOO
// 				filePath := fmt.Sprintf("%s/%s", data.Hash, v.Name)
// 				err := DiffDir(src, dst, filePath, v.Hash)
// 				if err != nil {
// 					return err
// 				}
// 			} else {
// 				return err
// 			}
// 		} else {
// 			err = DiffCall(src, dst, v.Hash, parentCID, filePath, []byte{})
// 			if err != nil {
// 				return err
// 			}
// 		}

// 	}

// 	return nil
// }
