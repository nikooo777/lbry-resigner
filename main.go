package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

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
		logrus.Infof("%s - claim_id: %s - outpoint: %s:%d - thumbnail url: %s", c.Name, c.ClaimID, c.Txid, c.Nout, c.Value.GetThumbnail().GetUrl())
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
		logrus.Infof("%s - claim_id: %s - outpoint: %s:%d - thumbnail url: %s", c.Name, c.ClaimID, c.Txid, c.Nout, c.Value.GetThumbnail().GetUrl())
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
	if newChannelID != "" && fundingAccountID != "" {
		logrus.Println("------preparing funds ------")
		err = ensureEnoughUTXOs(client, fundingAccountID, len(streamsToResign))
		if err != nil {
			panic(errors.Err(err))
		}
		logrus.Println("------updating ------")
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

func ensureEnoughUTXOs(client *jsonrpc.Client, spendAccount string, target int) error {
	utxolist, err := client.UTXOList(&spendAccount, 1, 10000)
	if err != nil {
		return err
	} else if utxolist == nil {
		return errors.Err("no response")
	}

	count := 0
	confirmedCount := 0

	for _, utxo := range utxolist.Items {
		amount, _ := strconv.ParseFloat(utxo.Amount, 64)
		if utxo.IsMyOutput && utxo.Type == "payment" && amount > 0.001 {
			if utxo.Confirmations > 0 {
				confirmedCount++
			}
			count++
		}
	}
	logrus.Infof("utxo count: %d (%d confirmed) out of %d needed", count, confirmedCount, target)
	if count < target {
		balance, err := client.AccountBalance(&spendAccount)
		if err != nil {
			return err
		} else if balance == nil {
			return errors.Err("no response")
		}

		balanceAmount, err := strconv.ParseFloat(balance.Available.String(), 64)
		if err != nil {
			return errors.Err(err)
		}
		//this is dumb but sometimes the balance is negative and it breaks everything, so let's check again
		if balanceAmount < 0 {
			logrus.Infof("negative balance of %.2f found. Waiting to retry...", balanceAmount)
			time.Sleep(10 * time.Second)
			balanceAmount, err = strconv.ParseFloat(balance.Available.String(), 64)
			if err != nil {
				return errors.Err(err)
			}
		}
		maxUTXOs := uint64(math.Min(float64(target), 500))
		desiredUTXOCount := uint64(math.Floor((balanceAmount) / 0.01))
		if desiredUTXOCount > maxUTXOs {
			desiredUTXOCount = maxUTXOs
		}
		if desiredUTXOCount < uint64(confirmedCount) {
			return nil
		}
		availableBalance, _ := balance.Available.Float64()
		logrus.Infof("Splitting balance of %.3f evenly between %d UTXOs", availableBalance, desiredUTXOCount)

		broadcastFee := 0.1
		prefillTx, err := client.AccountFund(spendAccount, spendAccount, fmt.Sprintf("%.4f", balanceAmount-broadcastFee), desiredUTXOCount, false)
		if err != nil {
			return err
		} else if prefillTx == nil {
			return errors.Err("no response")
		}
		err = waitForNewBlock(client)
		if err != nil {
			return err
		}
	} else if confirmedCount == 0 {
		err = waitForNewBlock(client)
		if err != nil {
			return err
		}
	}
	return nil
}

func waitForNewBlock(client *jsonrpc.Client) error {
	status, err := client.Status()
	if err != nil {
		return err
	}

	for status.Wallet.Blocks == 0 || status.Wallet.BlocksBehind != 0 {
		time.Sleep(5 * time.Second)
		status, err = client.Status()
		if err != nil {
			return err
		}
	}

	currentBlock := status.Wallet.Blocks
	for i := 0; status.Wallet.Blocks <= currentBlock; i++ {
		if i%3 == 0 {
			logrus.Printf("Waiting for new block (%d)...", currentBlock+1)
		}

		time.Sleep(10 * time.Second)
		status, err = client.Status()
		if err != nil {
			return err
		}
	}
	time.Sleep(5 * time.Second)
	return nil
}
