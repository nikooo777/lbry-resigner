package main

import (
	"fmt"
	"os"

	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/extras/jsonrpc"
	"github.com/lbryio/lbry.go/v2/extras/util"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	invalidChannelID string
	newChannelID     string
	fundingAccountID string
)

func main() {
	cmd := &cobra.Command{
		Use:   "resigner",
		Short: "automatically re-sign streams with invalid signing channels. Only pass --from to dry run",
		Run:   resigner,
		Args:  cobra.RangeArgs(0, 0),
	}

	cmd.Flags().StringVar(&invalidChannelID, "from", "", "claimID of the old channel")
	cmd.Flags().StringVar(&newChannelID, "to", "", "claimID of the new channel to sign streams with")
	cmd.Flags().StringVar(&fundingAccountID, "funding-account", "", "id of the funding account used to pay for the transaction")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func resigner(cmd *cobra.Command, args []string) {
	client := jsonrpc.NewClient("http://127.0.0.1:5279")

	accountList, err := client.AccountList(1, 1000)
	if err != nil {
		panic(errors.Err(err))
	}
	logrus.Println("------accounts------")
	for i, a := range accountList.Items {
		accountBalance, err := client.AccountBalance(&a.ID)
		if err != nil {
			panic(errors.Err(err))
		}
		logrus.Printf("%d) account id: %s - balance: %s - account name: %s", i, a.ID, accountBalance.Available.String(), a.Name)
	}

	chanMap := make(map[string]string, 10)
	unspentChannels, err := client.ChannelList(nil, 1, 1000, nil, false)
	if err != nil {
		panic(errors.Err(err))
	}
	spentChannels, err := client.ChannelList(nil, 1, 1000, nil, true)
	if err != nil {
		panic(errors.Err(err))
	}
	logrus.Println("------unspent channels------")

	for _, c := range unspentChannels.Items {
		chanMap[c.ClaimID] = c.Name
		if c.IsSpent {
			continue
		}
		logrus.Infof("%s - %s - %s:%d - %s", c.Name, c.ClaimID, c.Txid, c.Nout, c.Value.GetThumbnail())
	}
	logrus.Println("------spent channels------")

	for _, c := range spentChannels.Items {
		_, ok := chanMap[c.ClaimID]
		if !ok {
			chanMap[c.ClaimID] = c.Name
		}
		if !c.IsSpent {
			continue
		}
		logrus.Infof("%s - %s - %s:%d - %s", c.Name, c.ClaimID, c.Txid, c.Nout, c.Value.GetThumbnail())
	}

	streams, err := client.StreamList(nil, 1, 100000, false)
	if err != nil {
		panic(errors.Err(err))
	}
	logrus.Println("------streams without valid signatures------")

	streamsToResign := make([]jsonrpc.Claim, 0, len(streams.Items))
	for _, s := range streams.Items {
		if s.IsSpent {
			continue
		}
		if !s.IsChannelSignatureValid && s.SigningChannel != nil && s.SigningChannel.ChannelID != "" {
			if s.SigningChannel.ChannelID == invalidChannelID {
				streamsToResign = append(streamsToResign, s)
			}
			logrus.Printf("%s - invalid channel: %s (%s)", s.Name, chanMap[s.SigningChannel.ChannelID], s.SigningChannel.ChannelID)
		}
	}
	logrus.Println("------updating ------")
	if newChannelID != "" && fundingAccountID != "" {
		for _, s := range streamsToResign {
			streamCreateOptions := jsonrpc.StreamCreateOptions{
				ClaimCreateOptions: jsonrpc.ClaimCreateOptions{
					FundingAccountIDs: []string{fundingAccountID},
				},
				ChannelID: &newChannelID,
			}
			res, err := client.StreamUpdate(s.ClaimID, jsonrpc.StreamUpdateOptions{
				ClearTags:           util.PtrToBool(false),
				ClearLanguages:      util.PtrToBool(false),
				ClearLocations:      util.PtrToBool(false),
				StreamCreateOptions: &streamCreateOptions,
			})
			if err != nil {
				logrus.Errorln(errors.FullTrace(err))
				continue
			}
			logrus.Infof("successful update. TXID: %s", res.Txid)
		}
	} else {
		logrus.Infof("would have updated %d streams if --to and --funding-account were passed", len(streamsToResign))
	}
}
