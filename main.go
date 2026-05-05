package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"time"

	"stream_resigner/internal/lbryd"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func resigner(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	client := lbryd.NewClient("http://127.0.0.1:5279")

	accountList, err := client.AccountList(ctx, 1, 1000)
	if err != nil {
		panic(err)
	}
	logrus.Println("------accounts------")
	for i, a := range accountList.Items {
		accountBalance, err := client.AccountBalance(ctx, &a.ID)
		if err != nil {
			panic(err)
		}
		logrus.Printf("%d) account id: %s - balance: %s - account name: %s", i, a.ID, accountBalance.Available, a.Name)
	}

	chanMap := make(map[string]string, 10)

	unspent, err := client.ChannelList(ctx, 1, 1000, lbryd.ChannelListOptions{IsSpent: lbryd.Ptr(false)})
	if err != nil {
		panic(err)
	}
	logrus.Println("------unspent channels------")
	for _, c := range unspent.Items {
		chanMap[c.ClaimID] = c.Name
		logrus.Infof("%s - claim_id: %s - outpoint: %s:%d - thumbnail url: %s", c.Name, c.ClaimID, c.Txid, c.Nout, thumbnailURL(c.Value))
	}

	spent, err := client.ChannelList(ctx, 1, 1000, lbryd.ChannelListOptions{IsSpent: lbryd.Ptr(true)})
	if err != nil {
		panic(err)
	}
	logrus.Println("------spent channels------")
	for _, c := range spent.Items {
		if _, ok := chanMap[c.ClaimID]; !ok {
			chanMap[c.ClaimID] = c.Name
		}
		logrus.Infof("%s - claim_id: %s - outpoint: %s:%d - thumbnail url: %s", c.Name, c.ClaimID, c.Txid, c.Nout, thumbnailURL(c.Value))
	}

	streams, err := client.StreamList(ctx, 1, 100000)
	if err != nil {
		panic(err)
	}
	logrus.Println("------streams without valid signatures------")

	streamsToResign := make([]lbryd.Stream, 0, len(streams.Items))
	for _, s := range streams.Items {
		if s.IsSpent {
			continue
		}
		if !s.IsChannelSignatureValid && s.SigningChannel != nil && s.SigningChannel.ChannelID != nil {
			channelID := *s.SigningChannel.ChannelID
			if channelID == invalidChannelID {
				streamsToResign = append(streamsToResign, s)
			}
			logrus.Printf("%s - invalid channel: %s (%s)", s.Name, chanMap[channelID], channelID)
		}
	}
	if newChannelID != "" && fundingAccountID != "" {
		logrus.Println("------preparing funds ------")
		err = ensureEnoughUTXOs(ctx, client, fundingAccountID, len(streamsToResign))
		if err != nil {
			panic(err)
		}
		logrus.Println("------updating ------")
		for _, s := range streamsToResign {
			res, err := client.StreamUpdate(ctx, s.ClaimID, lbryd.StreamUpdateOptions{
				ChannelID:         &newChannelID,
				FundingAccountIDs: []string{fundingAccountID},
				ClearTags:         lbryd.Ptr(false),
				ClearLanguages:    lbryd.Ptr(false),
				ClearLocations:    lbryd.Ptr(false),
			})
			if err != nil {
				logrus.Errorln(err.Error())
				continue
			}
			logrus.Infof("successful update. TXID: %s", res.Txid)
		}
	} else {
		logrus.Infof("would have updated %d streams if --to and --funding-account were passed", len(streamsToResign))
	}
}

func thumbnailURL(v *lbryd.ClaimValue) string {
	if v == nil {
		return ""
	}
	return v.Thumbnail.URL
}

func ensureEnoughUTXOs(ctx context.Context, client *lbryd.Client, spendAccount string, target int) error {
	utxolist, err := client.UTXOList(ctx, &spendAccount, 1, 10000)
	if err != nil {
		return err
	} else if utxolist == nil {
		return errors.New("no response")
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
		balance, err := client.AccountBalance(ctx, &spendAccount)
		if err != nil {
			return err
		} else if balance == nil {
			return errors.New("no response")
		}

		balanceAmount, err := strconv.ParseFloat(balance.Available, 64)
		if err != nil {
			return err
		}
		//this is dumb but sometimes the balance is negative and it breaks everything, so let's check again
		if balanceAmount < 0 {
			logrus.Infof("negative balance of %.2f found. Waiting to retry...", balanceAmount)
			time.Sleep(10 * time.Second)
			balanceAmount, err = strconv.ParseFloat(balance.Available, 64)
			if err != nil {
				return err
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
		availableBalance, _ := strconv.ParseFloat(balance.Available, 64)
		logrus.Infof("Splitting balance of %.3f evenly between %d UTXOs", availableBalance, desiredUTXOCount)

		broadcastFee := 0.1
		prefillTx, err := client.AccountFund(ctx, spendAccount, spendAccount, fmt.Sprintf("%.4f", balanceAmount-broadcastFee), desiredUTXOCount, false)
		if err != nil {
			return err
		} else if prefillTx == nil {
			return errors.New("no response")
		}
		err = waitForNewBlock(ctx, client)
		if err != nil {
			return err
		}
	} else if confirmedCount == 0 {
		err = waitForNewBlock(ctx, client)
		if err != nil {
			return err
		}
	}
	return nil
}

func waitForNewBlock(ctx context.Context, client *lbryd.Client) error {
	status, err := client.Status(ctx)
	if err != nil {
		return err
	}

	for status.Wallet.Blocks == 0 || status.Wallet.BlocksBehind != 0 {
		time.Sleep(5 * time.Second)
		status, err = client.Status(ctx)
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
		status, err = client.Status(ctx)
		if err != nil {
			return err
		}
	}
	time.Sleep(5 * time.Second)
	return nil
}
