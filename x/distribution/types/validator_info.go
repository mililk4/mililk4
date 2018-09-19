package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// distribution info for a particular validator
type ValidatorDistInfo struct {
	OperatorAddr sdk.ValAddress `json:"operator_addr"`

	GlobalWithdrawalHeight int64    `json:"global_withdrawal_height"` // last height this validator withdrew from the global pool
	Pool                   DecCoins `json:"pool"`                     // rewards owed to delegators, commission has already been charged (includes proposer reward)
	PoolCommission         DecCoins `json:"pool_commission"`          // commission collected by this validator (pending withdrawal)

	DelAccum TotalAccum `json:"del_accum"` // total proposer pool accumulation factor held by delegators
}

// update total delegator accumululation
func (vi ValidatorDistInfo) UpdateTotalDelAccum(height int64, totalDelShares sdk.Dec) ValidatorDistInfo {
	vi.DelAccum = vi.DelAccum.Update(height, totalDelShares)
	return vi
}

// XXX TODO Update dec logic
// move any available accumulated fees in the Global to the validator's pool
func (vi ValidatorDistInfo) TakeFeePoolRewards(fp FeePool, height int64, totalBonded, vdTokens,
	commissionRate sdk.Dec) (ValidatorDistInfo, FeePool) {

	fp.UpdateTotalValAccum(height, totalBondedShares)

	// update the validators pool
	blocks = height - vi.GlobalWithdrawalHeight
	vi.GlobalWithdrawalHeight = height
	accum = sdk.NewDec(blocks).Mul(vdTokens)
	withdrawalTokens := fp.Pool.Mul(accum).Quo(fp.TotalValAccum)
	commission := withdrawalTokens.Mul(commissionRate)

	fp.TotalValAccum = fp.TotalValAccum.Sub(accum)
	fp.Pool = fp.Pool.Sub(withdrawalTokens)
	vi.PoolCommission = vi.PoolCommission.Add(commission)
	vi.PoolCommissionFree = vi.PoolCommissionFree.Add(withdrawalTokens.Sub(commission))

	return vi, fp
}

// withdraw commission rewards
func (vi ValidatorDistInfo) WithdrawCommission(g Global, height int64,
	totalBonded, vdTokens, commissionRate Dec) (vio ValidatorDistInfo, fpo FeePool, withdrawn DecCoins) {

	fp = vi.TakeFeePoolRewards(fp, height, totalBonded, vdTokens, commissionRate)

	withdrawalTokens := vi.PoolCommission
	vi.PoolCommission = 0

	return vi, fp, withdrawalTokens
}
