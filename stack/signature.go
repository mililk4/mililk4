package stack

import (
	"fmt"

	crypto "github.com/tendermint/go-crypto"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/types"
)

// app name for auth
const (
	NameSigs = "sigs"
)

type Signatures struct{}

func (_ Signatures) Name() string {
	return NameSigs
}

var _ Middleware = Signatures{}

func SigPerm(addr []byte) basecoin.Actor {
	return basecoin.NewActor(NameSigs, addr)
}

// Signed allows us to use txs.OneSig and txs.MultiSig (and others??)
type Signed interface {
	basecoin.TxLayer
	Signers() ([]crypto.PubKey, error)
}

func (h Signatures) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.CheckTx(ctx2, store, tnext)
}

func (h Signatures) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.DeliverTx(ctx2, store, tnext)
}

func addSigners(ctx basecoin.Context, sigs []crypto.PubKey) basecoin.Context {
	perms := make([]basecoin.Actor, len(sigs))
	fmt.Printf("Add %d signers\n", len(sigs))
	for i, s := range sigs {
		fmt.Printf("Add %X\n", s.Address())
		perms[i] = SigPerm(s.Address())
	}
	// add the signers to the context and continue
	return ctx.WithPermissions(perms...)
}

func getSigners(tx basecoin.Tx) ([]crypto.PubKey, basecoin.Tx, error) {
	fmt.Println("getSigners")
	stx, ok := tx.Unwrap().(Signed)
	if !ok {
		fmt.Printf("Not okay: %#v\n", tx)
		return nil, basecoin.Tx{}, errors.Unauthorized()
	}
	sig, err := stx.Signers()
	return sig, stx.Next(), err
}
