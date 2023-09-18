package sync

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/graphprotocol/ipfs-mgm/internal/utils"
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
}

func Sync(cmd *cobra.Command) {
	src, err := cmd.Flags().GetString("source")
	if err != nil {
		fmt.Println(err)
	}

	dst, err := cmd.Flags().GetString("destination")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(dst)
	utils.IPFSPost(src)
}
