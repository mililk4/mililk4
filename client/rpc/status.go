package rpc

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/spf13/viper"
	"github.com/cosmos/cosmos-sdk/client/utils"
)

func statusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE:  printNodeStatus,
	}

	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(client.FlagNode, cmd.Flags().Lookup(client.FlagNode))
	cmd.Flags().Bool(client.FlagIndentResponse, false, "Add indent to JSON response")
	return cmd
}

func getNodeStatus(cliCtx context.CLIContext) (*ctypes.ResultStatus, error) {
	// get the node
	node, err := cliCtx.GetNode()
	if err != nil {
		return &ctypes.ResultStatus{}, err
	}

	return node.Status()
}

// CMD

func printNodeStatus(cmd *cobra.Command, args []string) error {
	status, err := getNodeStatus(context.NewCLIContext())
	if err != nil {
		return err
	}

	var output []byte
	indent := viper.GetBool(client.FlagIndentResponse)
	if indent {
		output, err = cdc.MarshalJSONIndent(status, "", "  ")
	} else {
		output, err = cdc.MarshalJSON(status)
	}
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

// REST

// REST handler for node info
func NodeInfoRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(cliCtx)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		nodeInfo := status.NodeInfo
		utils.PostProcessResponse(w, cdc, nodeInfo)
	}
}

// REST handler for node syncing
func NodeSyncingRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(cliCtx)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		syncing := status.SyncInfo.CatchingUp
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte(strconv.FormatBool(syncing)))
	}
}
