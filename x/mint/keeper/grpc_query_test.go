package keeper_test

import (
	gocontext "context"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/burnt-labs/xion/x/mint"
	"github.com/burnt-labs/xion/x/mint/keeper"
	minttestutil "github.com/burnt-labs/xion/x/mint/testutil"
	"github.com/burnt-labs/xion/x/mint/types"
)

type MintTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	queryClient types.QueryClient
	mintKeeper  keeper.Keeper
}

func (suite *MintTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	store := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	accountKeeper := minttestutil.NewMockAccountKeeper(ctrl)
	bankKeeper := minttestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := minttestutil.NewMockStakingKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("mint").Return(sdk.AccAddress{})

	suite.mintKeeper = keeper.NewKeeper(
		encCfg.Codec,
		store,
		stakingKeeper,
		accountKeeper,
		bankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	err := suite.mintKeeper.SetParams(suite.ctx, types.DefaultParams())
	suite.Require().NoError(err)
	suite.mintKeeper.SetMinter(suite.ctx, types.DefaultInitialMinter())

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, suite.mintKeeper)

	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *MintTestSuite) TestGRPCParams() {
	params, err := suite.queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(params.Params, suite.mintKeeper.GetParams(suite.ctx))

	inflation, err := suite.queryClient.Inflation(gocontext.Background(), &types.QueryInflationRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(inflation.Inflation, suite.mintKeeper.GetMinter(suite.ctx).Inflation)

	annualProvisions, err := suite.queryClient.AnnualProvisions(gocontext.Background(), &types.QueryAnnualProvisionsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(annualProvisions.AnnualProvisions, suite.mintKeeper.GetMinter(suite.ctx).AnnualProvisions)
}

func TestMintTestSuite(t *testing.T) {
	suite.Run(t, new(MintTestSuite))
}
