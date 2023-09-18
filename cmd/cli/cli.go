package main

import (
   "fmt"
   "os"

   "github.com/spf13/cobra"
   "github.com/graphprotocol/ipfs-mgm/internal/sync"
)

var version = "2.0.0"

var rootCmd = &cobra.Command{
   Use: "ipfs-mgm",
   Version: version,
   Short: "ipfs-mgm - CLI for manage IPFS objects",
   Long: `ipfs-mgm is a simple CLI to manage the IPFS objects

One can use ipfs-mgm to migrate the objects from one IPFS endpoint to another or to sync both IPFS endpoints or to get the status of a file`,
   Run: func(cmd *cobra.Command, args []string) {
      cmd.Help()
   },
}

func init() {
   rootCmd.AddCommand(sync.SyncCmd)
}

func main() {
   if err := rootCmd.Execute(); err != nil {
      fmt.Fprintf(os.Stderr, "There was an error while executing your CLI '%s'", err)
      os.Exit(1)
   }
}
