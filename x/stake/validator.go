package stake

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

// Validator defines the total amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// validator, the validator is credited with a Delegation whose number of
// bond shares is based on the amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
type Validator struct {
	Status  sdk.BondStatus `json:"status"`  // bonded status
	Address sdk.Address    `json:"address"` // sender of BondTx - UnbondTx returns here
	PubKey  crypto.PubKey  `json:"pub_key"` // pubkey of validator

	PShares         PoolShares `json:"pool_shares"`      // total shares for tokens held in the pool
	DelegatorShares sdk.Rat    `json:"delegator_shares"` // total shares issued to a validator's delegators

	Description        Description `json:"description"`            // description terms for the validator
	BondHeight         int64       `json:"validator_bond_height"`  // earliest height as a bonded validator
	BondIntraTxCounter int16       `json:"validator_bond_counter"` // block-local tx index of validator change
	ProposerRewardPool sdk.Coins   `json:"proposer_reward_pool"`   // XXX reward pool collected from being the proposer

	Commission            sdk.Rat `json:"commission"`              // XXX the commission rate of fees charged to any delegators
	CommissionMax         sdk.Rat `json:"commission_max"`          // XXX maximum commission rate which this validator can ever charge
	CommissionChangeRate  sdk.Rat `json:"commission_change_rate"`  // XXX maximum daily increase of the validator commission
	CommissionChangeToday sdk.Rat `json:"commission_change_today"` // XXX commission rate change today, reset each day (UTC time)

	// fee related
	PrevBondedShares sdk.Rat `json:"prev_bonded_shares"` // total shares of a global hold pools
}

// Validators - list of Validators
type Validators []Validator

// NewValidator - initialize a new validator
func NewValidator(address sdk.Address, pubKey crypto.PubKey, description Description) Validator {
	return Validator{
		Status:                sdk.Unbonded,
		Address:               address,
		PubKey:                pubKey,
		PShares:               NewUnbondedShares(sdk.ZeroRat()),
		DelegatorShares:       sdk.ZeroRat(),
		Description:           description,
		BondHeight:            int64(0),
		BondIntraTxCounter:    int16(0),
		ProposerRewardPool:    sdk.Coins{},
		Commission:            sdk.ZeroRat(),
		CommissionMax:         sdk.ZeroRat(),
		CommissionChangeRate:  sdk.ZeroRat(),
		CommissionChangeToday: sdk.ZeroRat(),
		PrevBondedShares:      sdk.ZeroRat(),
	}
}

// only the vitals - does not check bond height of IntraTxCounter
func (v Validator) equal(c2 Validator) bool {
	return v.Status == c2.Status &&
		v.PubKey.Equals(c2.PubKey) &&
		bytes.Equal(v.Address, c2.Address) &&
		v.PShares.Equal(c2.PShares) &&
		v.DelegatorShares.Equal(c2.DelegatorShares) &&
		v.Description == c2.Description &&
		//v.BondHeight == c2.BondHeight &&
		//v.BondIntraTxCounter == c2.BondIntraTxCounter && // counter is always changing
		v.ProposerRewardPool.IsEqual(c2.ProposerRewardPool) &&
		v.Commission.Equal(c2.Commission) &&
		v.CommissionMax.Equal(c2.CommissionMax) &&
		v.CommissionChangeRate.Equal(c2.CommissionChangeRate) &&
		v.CommissionChangeToday.Equal(c2.CommissionChangeToday) &&
		v.PrevBondedShares.Equal(c2.PrevBondedShares)
}

// intended to be used with require/assert:  require.True(ValEq(...))
func ValEq(t *testing.T, exp, got Validator) (*testing.T, bool, string, Validator, Validator) {
	return t, exp.equal(got), "expected:\t%v\ngot:\t\t%v", exp, got
}

// Description - description fields for a validator
type Description struct {
	Moniker  string `json:"moniker"`
	Identity string `json:"identity"`
	Website  string `json:"website"`
	Details  string `json:"details"`
}

func NewDescription(moniker, identity, website, details string) Description {
	return Description{
		Moniker:  moniker,
		Identity: identity,
		Website:  website,
		Details:  details,
	}
}

//XXX updateDescription function
//XXX enforce limit to number of description characters

// abci validator from stake validator type
func (v Validator) abciValidator(cdc *wire.Codec) abci.Validator {
	return abci.Validator{
		PubKey: v.PubKey.Bytes(),
		Power:  v.PShares.Bonded().Evaluate(),
	}
}

// abci validator from stake validator type
// with zero power used for validator updates
func (v Validator) abciValidatorZero(cdc *wire.Codec) abci.Validator {
	return abci.Validator{
		PubKey: v.PubKey.Bytes(),
		Power:  0,
	}
}

// update the location of the shares within a validator if its bond status has changed
func (v Validator) UpdateSharesLocation(p Pool) (Validator, Pool) {
	var tokens int64

	switch {
	case v.PShares.Kind == ShareUnbonded:
		if v.Status == sdk.Unbonded {
			return v, p
		}
		p, tokens = p.removeSharesUnbonded(v.PShares.Amount)

	case v.PShares.Kind == ShareUnbonding:
		if v.Status == sdk.Unbonding {
			return v, p
		}
		p, tokens = p.removeSharesUnbonding(v.PShares.Amount)

	case v.PShares.Kind == ShareBonded:
		if v.Status == sdk.Bonded { // return if nothing needs switching
			return v, p
		}
		p, tokens = p.removeSharesBonded(v.PShares.Amount)
	}

	switch v.Status {
	case sdk.Unbonded, sdk.Revoked:
		p, v.PShares = p.addTokensUnbonded(tokens)
	case sdk.Unbonding:
		p, v.PShares = p.addTokensUnbonding(tokens)
	case sdk.Bonded:
		p, v.PShares = p.addTokensBonded(tokens)
	}
	return v, p
}

// XXX TEST
// get the power or potential power for a validator
// if bonded, the power is the BondedShares
// if not bonded, the power is the amount of bonded shares which the
//    the validator would have it was bonded
func (v Validator) EquivalentBondedShares(p Pool) (eqBondedShares sdk.Rat) {
	return v.PShares.ToBonded(p).Amount
}

//_________________________________________________________________________________________________________

// XXX Audit this function further to make sure it's correct
// add tokens to a validator
func (v Validator) addTokensFromDel(p Pool,
	amount int64) (validator2 Validator, p2 Pool, issuedDelegatorShares sdk.Rat) {

	exRate := v.DelegatorShareExRate(p) // bshr/delshr

	var poolShares PoolShares
	var equivalentBondedShares sdk.Rat
	switch v.Status {
	case sdk.Unbonded, sdk.Revoked:
		p, poolShares = p.addTokensUnbonded(amount)
	case sdk.Unbonding:
		p, poolShares = p.addTokensUnbonding(amount)
	case sdk.Bonded:
		p, poolShares = p.addTokensBonded(amount)
	}
	v.PShares.Amount = v.PShares.Amount.Add(poolShares.Amount)
	equivalentBondedShares = poolShares.ToBonded(p).Amount

	issuedDelegatorShares = equivalentBondedShares.Quo(exRate) // bshr/(bshr/delshr) = delshr
	v.DelegatorShares = v.DelegatorShares.Add(issuedDelegatorShares)

	return v, p, issuedDelegatorShares
}

// remove delegator shares from a validator
// NOTE this function assumes the shares have already been updated for the validator status
func (v Validator) removeDelShares(p Pool,
	delShares sdk.Rat) (validator2 Validator, p2 Pool, createdCoins int64) {

	amount := v.DelegatorShareExRate(p).Mul(delShares)
	eqBondedSharesToRemove := NewBondedShares(amount)
	v.DelegatorShares = v.DelegatorShares.Sub(delShares)

	switch v.Status {
	case sdk.Unbonded, sdk.Revoked:
		unbondedShares := eqBondedSharesToRemove.ToUnbonded(p).Amount
		p, createdCoins = p.removeSharesUnbonded(unbondedShares)
		v.PShares.Amount = v.PShares.Amount.Sub(unbondedShares)
	case sdk.Unbonding:
		unbondingShares := eqBondedSharesToRemove.ToUnbonding(p).Amount
		p, createdCoins = p.removeSharesUnbonding(unbondingShares)
		v.PShares.Amount = v.PShares.Amount.Sub(unbondingShares)
	case sdk.Bonded:
		p, createdCoins = p.removeSharesBonded(eqBondedSharesToRemove.Amount)
		v.PShares.Amount = v.PShares.Amount.Sub(eqBondedSharesToRemove.Amount)
	}
	return v, p, createdCoins
}

// get the exchange rate of tokens over delegator shares
// UNITS: eq-val-bonded-shares/delegator-shares
func (v Validator) DelegatorShareExRate(p Pool) sdk.Rat {
	if v.DelegatorShares.IsZero() {
		return sdk.OneRat()
	}
	eqBondedShares := v.PShares.ToBonded(p).Amount
	return eqBondedShares.Quo(v.DelegatorShares)
}

//______________________________________________________________________

// ensure fulfills the sdk validator types
var _ sdk.Validator = Validator{}

// nolint - for sdk.Validator
func (v Validator) GetStatus() sdk.BondStatus { return v.Status }
func (v Validator) GetAddress() sdk.Address   { return v.Address }
func (v Validator) GetPubKey() crypto.PubKey  { return v.PubKey }
func (v Validator) GetPower() sdk.Rat         { return v.PShares.Bonded() }
func (v Validator) GetBondHeight() int64      { return v.BondHeight }
