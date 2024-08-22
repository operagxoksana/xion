package integration_tests

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/strangelove-ventures/interchaintest/v8/conformance"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/relayer/rly"
	"os"
	"strconv"
	"testing"
	"time"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"cosmossdk.io/math"

	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	breaksIBC = true

	xionImageFrom   = "ghcr.io/burnt-labs/xion/xion"
	xionVersionFrom = "v9.0.0"
	xionImageTo     = "ghcr.io/burnt-labs/xion/heighliner"
	xionVersionTo   = "sha-962b654"
	xionUpgradeName = "v10"

	osmosisImage   = "ghcr.io/strangelove-ventures/heighliner/osmosis"
	osmosisVersion = "v25.2.1"

	relayerImage   = "ghcr.io/cosmos/relayer"
	relayerVersion = "main"
	relayerImpl    = ibc.CosmosRly

	ibcClientTrustingPeriod           = "336h"
	ibcClienttrustingPeriodPercentage = 66
	ibcClientMaxClockDrift            = "5s"

	ibcChannelSourcePort      = "transfer"
	ibcChannelDestinationPort = "transfer"
	ibcChannelOrder           = ibc.Unordered
	ibcChannelVersion         = "ics20-1"

	authority = "xion10d07y265gmmuvt4z0w9aw880jnsr700jctf8qc" // Governance authority address
)

// XionTestMinion likes bananas
type XionTestMinion struct {
	ctx context.Context

	Name          string
	DockerClient  *client.Client
	DockerNetwork string
	Reporter      *testreporter.Reporter

	Xion         *cosmos.CosmosChain
	Counterparty *cosmos.CosmosChain
	Interchain   *interchaintest.Interchain

	Relayer             ibc.Relayer
	RelayerFactory      interchaintest.RelayerFactory
	RelayerClientOpts   ibc.CreateClientOptions
	RelayerChannelOpts  ibc.CreateChannelOptions
	RelayerExecReporter *testreporter.RelayerExecReporter

	IBCClientUpgradePath []string
}

// NewXionTestMinion spawns a new XionTestMinion.
func NewXionTestMinion(t *testing.T, name string) *XionTestMinion {
	// hook into the local docker network
	ctx := context.Background()
	dockerClient, dockerNetwork := interchaintest.DockerSetup(t)

	// return a new minion
	return &XionTestMinion{
		ctx: ctx,

		Name:          name,
		DockerClient:  dockerClient,
		DockerNetwork: dockerNetwork,

		IBCClientUpgradePath: []string{"upgrade", "UpgradedIBCState"},
	}
}

// TestXionUpgradeIBC tests a Xion software upgrade, ensuring IBC conformance prior-to and after the upgrade.
func TestXionUpgradeIBC(t *testing.T) {

	// Define Test cases
	testCases := []struct {
		name             string
		xionSpec         *interchaintest.ChainSpec
		counterpartySpec *interchaintest.ChainSpec
	}{
		{
			name: "xion-osmosis",
			xionSpec: &interchaintest.ChainSpec{
				Name:    "xion",
				Version: xionVersionFrom,
				ChainConfig: ibc.ChainConfig{
					Images: []ibc.DockerImage{
						{
							Repository: xionImageFrom,
							Version:    xionVersionFrom,
							UidGid:     "1025:1025",
						},
					},
					GasPrices:      "0.0uxion",
					GasAdjustment:  1.3,
					Type:           "cosmos",
					ChainID:        "xion-1",
					Bin:            "xiond",
					Bech32Prefix:   "xion",
					Denom:          "uxion",
					TrustingPeriod: ibcClientTrustingPeriod,
					NoHostMount:    false,
					ModifyGenesis:  ModifyInterChainGenesis(ModifyInterChainGenesisFn{ModifyGenesisShortProposals}, [][]string{{votingPeriod, maxDepositPeriod}}),
				},
			},
			counterpartySpec: &interchaintest.ChainSpec{
				Name:    "osmosis",
				Version: osmosisVersion,
				ChainConfig: ibc.ChainConfig{
					Images: []ibc.DockerImage{
						{
							Repository: osmosisImage,
							Version:    osmosisVersion,
							UidGid:     "1025:1025",
						},
					},
					Type:           "cosmos",
					Bin:            "osmosisd",
					Bech32Prefix:   "osmo",
					Denom:          "uosmo",
					GasPrices:      "0.025uosmo",
					GasAdjustment:  1.3,
					TrustingPeriod: ibcClientTrustingPeriod,
					NoHostMount:    false,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Override default embedded configuredChains.yaml
			f := OverrideConfiguredChainsYaml(t)
			defer os.Remove(f.Name())

			// spawn a minion
			x := NewXionTestMinion(t, tc.name)

			// feed the minion
			x.SetupInterchain(t, tc.xionSpec, tc.counterpartySpec)
			defer x.Interchain.Close()
			defer x.DockerClient.Close()

			// check for IBC conformance prior to the upgrade
			x.IBCConformance(t)
			// upgrade the Xion chain
			x.XionUpgrade(t)
			// check for IBC conformance after the upgrade
			x.IBCConformance(t)
		})
	}
}

// SetupInterchain configures an interchaintest.Interchain with the provided chain specs.
func (x *XionTestMinion) SetupInterchain(t *testing.T, xionSpec, counterpartySpec *interchaintest.ChainSpec) {
	// loggers and reporters
	f, err := interchaintest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(t, err)
	x.Reporter = testreporter.NewReporter(f)
	x.RelayerExecReporter = x.Reporter.RelayerExecReporter(t)

	// build chainFactory
	cf := interchaintest.NewBuiltinChainFactory(
		zaptest.NewLogger(t),
		[]*interchaintest.ChainSpec{
			xionSpec,
			counterpartySpec,
		},
	)

	// create chains
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err, "error creating chains")

	// feed configured chains to the minion
	x.Xion, x.Counterparty = chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// set relayer client options
	x.RelayerClientOpts = ibc.CreateClientOptions{
		TrustingPeriod:           ibcClientTrustingPeriod,
		TrustingPeriodPercentage: ibcClienttrustingPeriodPercentage,
		MaxClockDrift:            ibcClientMaxClockDrift,
	}
	err = x.RelayerClientOpts.Validate()
	require.NoError(t, err, "couldn't validate relayer client options: %v", err)

	// set relayer channel options
	x.RelayerChannelOpts = ibc.CreateChannelOptions{
		SourcePortName: ibcChannelSourcePort,
		DestPortName:   ibcChannelDestinationPort,
		Order:          ibcChannelOrder,
		Version:        ibcChannelVersion,
	}
	err = x.RelayerChannelOpts.Validate()
	require.NoError(t, err, "couldn't validate relayer channel options: %v", err)

	// build relayer
	rlyImage := relayer.CustomDockerImage(relayerImage, relayerVersion, rly.RlyDefaultUidGid)
	x.RelayerFactory = interchaintest.NewBuiltinRelayerFactory(relayerImpl, zaptest.NewLogger(t), rlyImage)
	x.Relayer = x.RelayerFactory.Build(t, x.DockerClient, x.DockerNetwork)

	// configure interchain
	x.Interchain = interchaintest.NewInterchain().
		AddChain(x.Xion).
		AddChain(x.Counterparty).
		AddRelayer(x.Relayer, "rly").
		AddLink(interchaintest.InterchainLink{
			Chain1:            x.Xion,
			Chain2:            x.Counterparty,
			Relayer:           x.Relayer,
			Path:              x.Name,
			CreateClientOpts:  x.RelayerClientOpts,
			CreateChannelOpts: x.RelayerChannelOpts,
		})

	// build interchain
	err = x.Interchain.Build(x.ctx, x.RelayerExecReporter, interchaintest.InterchainBuildOptions{
		TestName:          x.Name,
		Client:            x.DockerClient,
		NetworkID:         x.DockerNetwork,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  false,
	})
	require.NoError(t, err, "couldn't build interchain: %v", err)
}

// IBCConformance explodes the XionTestMinion and sends chunks downstream to ICT.
func (x *XionTestMinion) IBCConformance(t *testing.T) {
	conformance.TestChainPair(
		t,
		x.ctx,
		x.DockerClient,
		x.DockerNetwork,
		x.Xion,
		x.Counterparty,
		x.RelayerFactory,
		x.Reporter,
		x.Relayer,
		x.Name,
	)
}

// XionUpgrade attempts to upgrade a chain, and optionally handles breaking IBC changes.
func (x *XionTestMinion) XionUpgrade(t *testing.T) {
	// waiting on blocks with finite context
	ctxTimeout, ctxTimeoutCancel := context.WithTimeout(x.ctx, time.Second*45)
	defer ctxTimeoutCancel()

	// Fund proposer
	fundAmount := math.NewInt(20_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, x.ctx, "default", fundAmount, x.Xion)
	chainUser := users[0]

	// determine halt height
	currentHeight, err := x.Xion.Height(x.ctx)
	require.NoErrorf(t, err, "couldn't get chain height: %v", err)
	haltHeight := currentHeight + haltHeightDelta - 3

	if breaksIBC {
		// submit IBC upgrade govprop
		err = x.submitIBCSoftwareUpgradeProposal(t, chainUser, haltHeight, currentHeight)
		require.NoErrorf(t, err, "couldn't submit IBC upgrade proposal: %v", err)
	}

	// submit chain upgrade govprop
	err = x.submitSoftwareUpgradeProposal(t, chainUser, haltHeight, currentHeight)
	require.NoErrorf(t, err, "couldn't submit xionUpgrade proposal: %v", err)

	// confirm chain halt
	_ = testutil.WaitForBlocks(ctxTimeout, int(haltHeight-currentHeight), x.Xion)
	currentHeight, err = x.Xion.Height(x.ctx)
	require.NoErrorf(t, err, "couldn't get chain height after chain should have halted: %v", err)
	// ERR CONSENSUS FAILURE!!! err="UPGRADE \"v10\" NEEDED at height: 80: Software Upgrade v10" module=consensus
	// INF Timed out dur=2000 height=81 module=consensus round=0 step=RoundStepPropose
	require.GreaterOrEqualf(t, currentHeight, haltHeight, "currentHeight: %d is not >= to haltHeight: %d", currentHeight, haltHeight)

	// upgrade all nodes
	err = x.Xion.StopAllNodes(x.ctx)
	require.NoErrorf(t, err, "couldn't stop nodes: %v", err)
	x.Xion.UpgradeVersion(x.ctx, x.DockerClient, xionImageTo, xionVersionTo)

	if breaksIBC {
		// upgrade foreign IBC clients
		x.upgradeCounterpartyClients()
	}

	// reboot nodes
	err = x.Xion.StartAllNodes(x.ctx)
	require.NoErrorf(t, err, "couldn't reboot nodes: %v", err)

	// banana?? 🍌🍌🍌
	err = testutil.WaitForBlocks(ctxTimeout, int(blocksAfterUpgrade), x.Xion)
	require.NoError(t, err, "chain did not produce blocks after upgrade")
}

// submitIBCSoftwareUpgradeProposal submits and passes an IBCSoftwareUpgrade govprop.
func (x *XionTestMinion) submitIBCSoftwareUpgradeProposal(
	t *testing.T,
	chainUser ibc.Wallet,
	currentHeight int64,
	haltHeight int64,
) (error error) {
	// https://github.com/cosmos/ibc-go/blob/main/docs/docs/01-ibc/05-upgrades/01-quick-guide.md#step-by-step-upgrade-process-for-sdk-chains

	// An UpgradedClientState must be provided to perform an IBC breaking upgrade.
	// This will make the chain commit to the correct upgraded (self) client state
	// before the upgrade occurs, so that connecting chains can verify that the
	// new upgraded client is valid by verifying a proof on the previous version
	// of the chain.

	// The UpgradePlan must specify an upgrade height only (no upgrade time),
	// and the ClientState should only include the fields common to all valid clients
	// (chain-specified parameters) and zero out any client-customizable fields
	// (such as TrustingPeriod).

	upgradeInfo := fmt.Sprintf("Software Upgrade %s", xionUpgradeName)

	upgradedClientState := &ibctm.ClientState{
		ChainId:                      x.Xion.Config().ChainID,
		UpgradePath:                  x.IBCClientUpgradePath,
		AllowUpdateAfterExpiry:       true,
		AllowUpdateAfterMisbehaviour: true,
	}
	upgradedClientStateAny, err := ibcclienttypes.PackClientState(upgradedClientState)
	require.NoError(t, err, "couldn't pack upgraded client state: %v", err)

	// Set upgrade plan
	plan := upgradetypes.Plan{
		Name:   xionUpgradeName,
		Height: haltHeight,
		Info:   upgradeInfo,
	}

	// Legacy upgrade / ibc-go v7 and earlier
	upgrade := &ibcclienttypes.UpgradeProposal{
		Title:               upgradeInfo,
		Description:         upgradeInfo,
		Plan:                plan,
		UpgradedClientState: upgradedClientStateAny,
	}

	// IBCSoftwareUpgrade / ibc-go v8 and later
	//upgrade := &ibcclienttypes.MsgIBCSoftwareUpgrade{
	//	Plan:                plan,
	//	UpgradedClientState: upgradedClientStateAny,
	//	Signer:              authority,
	//}

	// Get proposer addr and keyname
	address, err := x.Xion.GetAddress(x.ctx, chainUser.KeyName())
	require.NoError(t, err)
	proposerAddr, err := sdk.Bech32ifyAddressBytes(x.Xion.Config().Bech32Prefix, address)
	require.NoError(t, err)
	proposerKeyname := chainUser.KeyName()

	// Build govprop
	proposal, err := x.Xion.BuildProposal(
		[]cosmos.ProtoMessage{upgrade},
		upgradeInfo,
		upgradeInfo,
		"",
		fmt.Sprintf("%d%s", 10_000_000, x.Xion.Config().Denom),
		proposerAddr,
		true,
	)
	require.NoError(t, err)

	// Submit govprop
	err = x.submitGovprop(t, proposerKeyname, proposal, currentHeight)
	require.NoError(t, err, "couldn't submit govprop: %v", err)

	// Upon passing the governance proposal, the upgrade module will commit the
	// UpgradedClient under the key:
	//
	// upgrade/UpgradedIBCState/{upgradeHeight}/upgradedClient.
	//
	// On the block right before the upgrade height, the upgrade module will also commit
	// an initial consensus state for the next chain under the key:
	//
	// upgrade/UpgradedIBCState/{upgradeHeight}/upgradedConsState.
	//
	// Once the chain reaches the upgrade height and halts, a relayer can upgrade
	// the counterparty clients to the last block of the old chain. They can then submit
	// the proofs of the UpgradedClient and UpgradedConsensusState against this last block
	// and upgrade the counterparty client.

	return err
}

// submitSoftwareUpgradeProposal submits and passes a SoftwareUpgrade govprop.
func (x *XionTestMinion) submitSoftwareUpgradeProposal(
	t *testing.T,
	chainUser ibc.Wallet,
	currentHeight int64,
	haltHeight int64,
) (error error) {
	upgradeInfo := fmt.Sprintf("Software Upgrade %s", xionUpgradeName)

	// Get proposer addr and keyname
	proposerKeyname := chainUser.KeyName()
	address, err := x.Xion.GetAddress(x.ctx, proposerKeyname)
	require.NoError(t, err)
	proposerAddr, err := sdk.Bech32ifyAddressBytes(x.Xion.Config().Bech32Prefix, address)
	require.NoError(t, err)

	// Build SoftwareUpgrade message
	plan := upgradetypes.Plan{
		Name:   xionUpgradeName,
		Height: haltHeight,
		Info:   upgradeInfo,
	}
	upgrade := upgradetypes.MsgSoftwareUpgrade{
		Authority: authority,
		Plan:      plan,
	}

	// Build govprop
	proposal, err := x.Xion.BuildProposal(
		[]cosmos.ProtoMessage{&upgrade},
		upgradeInfo,
		upgradeInfo,
		"",
		fmt.Sprintf("%d%s", 10_000_000, x.Xion.Config().Denom),
		proposerAddr,
		true,
	)
	require.NoError(t, err)

	// Submit govprop
	err = x.submitGovprop(t, proposerKeyname, proposal, currentHeight)
	require.NoError(t, err, "couldn't submit govprop: %v", err)

	return err
}

// submitGovprop submits a cosmos.TxProposalv1 and ensures it passes.
func (x *XionTestMinion) submitGovprop(
	t *testing.T,
	proposerKeyname string,
	proposal cosmos.TxProposalv1,
	currentHeight int64,
) (err error) {

	// Submit govprop
	tx, err := x.Xion.SubmitProposal(x.ctx, proposerKeyname, proposal)
	require.NoError(t, err)

	// Ensure prop exists and is vote-able
	propId, err := strconv.Atoi(tx.ProposalID)
	require.NoError(t, err, "couldn't convert proposal ID to int: %v", err)
	prop, err := x.Xion.GovQueryProposal(x.ctx, uint64(propId))
	require.NoError(t, err, "couldn't query proposal: %v", err)
	require.Equal(t, govv1beta1.StatusVotingPeriod, prop.Status)

	// Vote on govprop
	err = x.Xion.VoteOnProposalAllValidators(x.ctx, prop.ProposalId, cosmos.ProposalVoteYes)
	require.NoErrorf(t, err, "couldn't submit votes: %v", err)

	// Ensure govprop passed
	_, err = cosmos.PollForProposalStatus(x.ctx, x.Xion, currentHeight, currentHeight+haltHeightDelta, prop.ProposalId, govv1beta1.StatusPassed)
	require.NoErrorf(t, err, "couldn't poll for proposal status: %v", err)

	return err
}

func (x *XionTestMinion) upgradeCounterpartyClients() {
	// https://github.com/cosmos/ibc-go/blob/main/docs/docs/01-ibc/05-upgrades/01-quick-guide.md#step-by-step-upgrade-process-for-relayers-upgrading-counterparty-clients

	// Once the upgrading chain has committed to upgrading,
	// relayers must wait till the chain halts at the upgrade height before upgrading counterparty clients.
	//
	// This is because chains may reschedule or cancel upgrade plans before they occur.
	// Thus, relayers must wait till the chain reaches the upgrade height and halts
	// before they can be sure the upgrade will take place.
	//
	// Thus, the upgrade process for relayers trying to upgrade the counterparty clients is as follows:
	// - Wait for the upgrading chain to reach the upgrade height and halt
	// - Query a full node for the proofs of UpgradedClient and UpgradedConsensusState at the last height of the old chain.
	// - Update the counterparty client to the last height of the old chain using the UpdateClient msg.
	// - Submit an UpgradeClient msg to the counterparty chain with the UpgradedClient, UpgradedConsensusState and their respective proofs.
	// - Submit an UpdateClient msg to the counterparty chain with a header from the new upgraded chain.
}
