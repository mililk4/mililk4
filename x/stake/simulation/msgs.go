package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// SimulateMsgCreateValidator
func SimulateMsgCreateValidator(m auth.AccountMapper, k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		description := stake.Description{
			Moniker: simulation.RandStringOfLength(r, 10),
		}
		key := keys[r.Intn(len(keys))]
		pubkey := key.PubKey()
		address := sdk.AccAddress(pubkey.Address())
		amount := m.GetAccount(ctx, address).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = sdk.NewInt(int64(r.Intn(int(amount.Int64()))))
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "nop", nil
		}
		msg := stake.MsgCreateValidator{
			Description:   description,
			ValidatorAddr: address,
			DelegatorAddr: address,
			PubKey:        pubkey,
			Delegation:    sdk.NewIntCoin(denom, amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgCreateValidator/%v", result.IsOK()))
		// require.True(t, result.IsOK(), "expected OK result but instead got %v", result)
		action = fmt.Sprintf("TestMsgCreateValidator: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgEditValidator
func SimulateMsgEditValidator(k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		description := stake.Description{
			Moniker:  simulation.RandStringOfLength(r, 10),
			Identity: simulation.RandStringOfLength(r, 10),
			Website:  simulation.RandStringOfLength(r, 10),
			Details:  simulation.RandStringOfLength(r, 10),
		}
		key := keys[r.Intn(len(keys))]
		pubkey := key.PubKey()
		address := sdk.AccAddress(pubkey.Address())
		msg := stake.MsgEditValidator{
			Description:   description,
			ValidatorAddr: address,
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgEditValidator/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgEditValidator: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgDelegate
func SimulateMsgDelegate(m auth.AccountMapper, k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		validatorKey := keys[r.Intn(len(keys))]
		validatorAddress := sdk.AccAddress(validatorKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = sdk.NewInt(int64(r.Intn(int(amount.Int64()))))
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "nop", nil
		}
		msg := stake.MsgDelegate{
			DelegatorAddr: delegatorAddress,
			ValidatorAddr: validatorAddress,
			Delegation:    sdk.NewIntCoin(denom, amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgDelegate/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgDelegate: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgBeginUnbonding
func SimulateMsgBeginUnbonding(m auth.AccountMapper, k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		validatorKey := keys[r.Intn(len(keys))]
		validatorAddress := sdk.AccAddress(validatorKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = sdk.NewInt(int64(r.Intn(int(amount.Int64()))))
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "nop", nil
		}
		msg := stake.MsgBeginUnbonding{
			DelegatorAddr: delegatorAddress,
			ValidatorAddr: validatorAddress,
			SharesAmount:  sdk.NewRatFromInt(amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgBeginUnbonding/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgBeginUnbonding: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgCompleteUnbonding
func SimulateMsgCompleteUnbonding(k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		validatorKey := keys[r.Intn(len(keys))]
		validatorAddress := sdk.AccAddress(validatorKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		msg := stake.MsgCompleteUnbonding{
			DelegatorAddr: delegatorAddress,
			ValidatorAddr: validatorAddress,
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgCompleteUnbonding/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgCompleteUnbonding with %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgBeginRedelegate
func SimulateMsgBeginRedelegate(m auth.AccountMapper, k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		sourceValidatorKey := keys[r.Intn(len(keys))]
		sourceValidatorAddress := sdk.AccAddress(sourceValidatorKey.PubKey().Address())
		destValidatorKey := keys[r.Intn(len(keys))]
		destValidatorAddress := sdk.AccAddress(destValidatorKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		// TODO
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = sdk.NewInt(int64(r.Intn(int(amount.Int64()))))
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "nop", nil
		}
		msg := stake.MsgBeginRedelegate{
			DelegatorAddr:    delegatorAddress,
			ValidatorSrcAddr: sourceValidatorAddress,
			ValidatorDstAddr: destValidatorAddress,
			SharesAmount:     sdk.NewRatFromInt(amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgBeginRedelegate/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgBeginRedelegate: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgCompleteRedelegate
func SimulateMsgCompleteRedelegate(k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		validatorSrcKey := keys[r.Intn(len(keys))]
		validatorSrcAddress := sdk.AccAddress(validatorSrcKey.PubKey().Address())
		validatorDstKey := keys[r.Intn(len(keys))]
		validatorDstAddress := sdk.AccAddress(validatorDstKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		msg := stake.MsgCompleteRedelegate{
			DelegatorAddr:    delegatorAddress,
			ValidatorSrcAddr: validatorSrcAddress,
			ValidatorDstAddr: validatorDstAddress,
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgCompleteRedelegate/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgCompleteRedelegate with %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulationSetup
func SimulationSetup(mapp *mock.App, k stake.Keeper) simulation.RandSetup {
	return func(r *rand.Rand, privKeys []crypto.PrivKey) {
		ctx := mapp.NewContext(false, abci.Header{})
		stake.InitGenesis(ctx, k, stake.DefaultGenesisState())
		params := k.GetParams(ctx)
		denom := params.BondDenom
		loose := sdk.ZeroInt()
		mapp.AccountMapper.IterateAccounts(ctx, func(acc auth.Account) bool {
			balance := sdk.NewInt(int64(r.Intn(1000000)))
			acc.SetCoins(acc.GetCoins().Plus(sdk.Coins{sdk.NewIntCoin(denom, balance)}))
			mapp.AccountMapper.SetAccount(ctx, acc)
			loose = loose.Add(balance)
			return false
		})
		pool := k.GetPool(ctx)
		pool.LooseTokens = pool.LooseTokens.Add(sdk.NewRat(loose.Int64(), 1))
		k.SetPool(ctx, pool)
	}
}
