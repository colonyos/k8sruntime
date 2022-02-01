package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/kolony/pkg/colony"
	"github.com/spf13/cobra"
)

func init() {
	serverCmd.AddCommand(serverStartCmd)
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Colonies Server host")
	serverCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Colonies Server port")
	serverCmd.PersistentFlags().StringVarP(&KubeColonyID, "kubeid", "", "", "The Id of the Colony where Kolony will register itself")
	serverCmd.PersistentFlags().StringVarP(&KubeColonyPrvKey, "kubeprvkey", "", "", "The PrvKey of the Colony where Kolony will register itself")
	serverCmd.PersistentFlags().StringVarP(&TargetColonyID, "targetid", "", "", "The Id of the Colony where the Kolony will spawn new runtimes")
	serverCmd.PersistentFlags().StringVarP(&TargetColonyPrvKey, "targetprvkey", "", "", "The PrvKey of the Colony where the Kolony will spawn new runtimes")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage a Kolony server",
	Long:  "Manage a Kolony server",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Kolony server",
	Long:  "Start a Kolony server",
	Run: func(cmd *cobra.Command, args []string) {
		if KubeColonyID == "" {
			KubeColonyID = os.Getenv("KUBECOLONYID")
		}
		if KubeColonyID == "" {
			CheckError(errors.New("Unknown Kube Colony Id"))
		}

		if TargetColonyID == "" {
			TargetColonyID = os.Getenv("TARGETCOLONYID")
		}
		if TargetColonyID == "" {
			CheckError(errors.New("Unknown Target Colony Id"))
		}

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if KubeColonyPrvKey == "" {
			KubeColonyPrvKey, err = keychain.GetPrvKey(KubeColonyID)
			CheckError(err)
		}

		if TargetColonyPrvKey == "" {
			TargetColonyPrvKey, err = keychain.GetPrvKey(TargetColonyID)
			CheckError(err)
		}

		kubeCRT, err := colony.CreateKubeColonyRT("kolony", ServerHost, ServerPort, KubeColonyID, KubeColonyPrvKey, TargetColonyID, TargetColonyPrvKey, "test")
		CheckError(err)

		err = kubeCRT.ServeForEver()
		CheckError(err)

		fmt.Println("Waiting")
		<-make(chan bool)
	},
}
