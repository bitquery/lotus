package state

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	"github.com/stretchr/testify/require"
	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/network"
	builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"

	"github.com/filecoin-project/lotus/build/buildconstants"
	"github.com/filecoin-project/lotus/chain/types"
)

func BenchmarkStateTreeSet(b *testing.B) {
	cst := cbor.NewMemCborStore()
	st, err := NewStateTree(cst, types.StateTreeVersion1)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		a, err := address.NewIDAddress(uint64(b.N))
		if err != nil {
			b.Fatal(err)
		}
		err = st.SetActor(a, &types.Actor{
			Balance: types.NewInt(1258812523),
			Code:    builtin2.StorageMinerActorCodeID,
			Head:    builtin2.AccountActorCodeID,
			Nonce:   uint64(b.N),
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStateTreeSetFlush(b *testing.B) {
	cst := cbor.NewMemCborStore()
	sv, err := VersionForNetwork(buildconstants.TestNetworkVersion)
	if err != nil {
		b.Fatal(err)
	}

	st, err := NewStateTree(cst, sv)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		a, err := address.NewIDAddress(uint64(b.N))
		if err != nil {
			b.Fatal(err)
		}
		err = st.SetActor(a, &types.Actor{
			Balance: types.NewInt(1258812523),
			Code:    builtin2.StorageMinerActorCodeID,
			Head:    builtin2.AccountActorCodeID,
			Nonce:   uint64(b.N),
		})
		if err != nil {
			b.Fatal(err)
		}
		if _, err := st.Flush(context.TODO()); err != nil {
			b.Fatal(err)
		}
	}
}

func TestResolveCache(t *testing.T) {
	cst := cbor.NewMemCborStore()
	sv, err := VersionForNetwork(buildconstants.TestNetworkVersion)
	if err != nil {
		t.Fatal(err)
	}

	st, err := NewStateTree(cst, sv)
	if err != nil {
		t.Fatal(err)
	}
	nonId := address.NewForTestGetter()()
	id, _ := address.NewIDAddress(1000)

	st.lookupIDFun = func(a address.Address) (address.Address, error) {
		if a == nonId {
			return id, nil
		}
		return address.Undef, types.ErrActorNotFound
	}

	err = st.SetActor(nonId, &types.Actor{Nonce: 1})
	if err != nil {
		t.Fatal(err)
	}

	{
		err = st.Snapshot(context.TODO())
		if err != nil {
			t.Fatal(err)
		}
		act, err := st.GetActor(nonId)
		if err != nil {
			t.Fatal(err)
		}
		if act.Nonce != 1 {
			t.Fatalf("expected nonce 1, got %d", act.Nonce)
		}
		err = st.SetActor(nonId, &types.Actor{Nonce: 2})
		if err != nil {
			t.Fatal(err)
		}

		act, err = st.GetActor(nonId)
		if err != nil {
			t.Fatal(err)
		}
		if act.Nonce != 2 {
			t.Fatalf("expected nonce 2, got %d", act.Nonce)
		}

		if err := st.Revert(); err != nil {
			t.Fatal(err)
		}
		st.ClearSnapshot()
	}

	act, err := st.GetActor(nonId)
	if err != nil {
		t.Fatal(err)
	}
	if act.Nonce != 1 {
		t.Fatalf("expected nonce 1, got %d", act.Nonce)
	}

	{
		err = st.Snapshot(context.TODO())
		if err != nil {
			t.Fatal(err)
		}
		act, err := st.GetActor(nonId)
		if err != nil {
			t.Fatal(err)
		}
		if act.Nonce != 1 {
			t.Fatalf("expected nonce 1, got %d", act.Nonce)
		}
		err = st.SetActor(nonId, &types.Actor{Nonce: 2})
		if err != nil {
			t.Fatal(err)
		}

		act, err = st.GetActor(nonId)
		if err != nil {
			t.Fatal(err)
		}
		if act.Nonce != 2 {
			t.Fatalf("expected nonce 2, got %d", act.Nonce)
		}
		st.ClearSnapshot()
	}

	act, err = st.GetActor(nonId)
	if err != nil {
		t.Fatal(err)
	}
	if act.Nonce != 2 {
		t.Fatalf("expected nonce 2, got %d", act.Nonce)
	}

}

func BenchmarkStateTree10kGetActor(b *testing.B) {
	cst := cbor.NewMemCborStore()
	sv, err := VersionForNetwork(buildconstants.TestNetworkVersion)
	if err != nil {
		b.Fatal(err)
	}

	st, err := NewStateTree(cst, sv)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 10000; i++ {
		a, err := address.NewIDAddress(uint64(i))
		if err != nil {
			b.Fatal(err)
		}
		err = st.SetActor(a, &types.Actor{
			Balance: types.NewInt(1258812523 + uint64(i)),
			Code:    builtin2.StorageMinerActorCodeID,
			Head:    builtin2.AccountActorCodeID,
			Nonce:   uint64(i),
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	if _, err := st.Flush(context.TODO()); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		a, err := address.NewIDAddress(uint64(b.N % 10000))
		if err != nil {
			b.Fatal(err)
		}

		_, err = st.GetActor(a)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestSetCache(t *testing.T) {
	cst := cbor.NewMemCborStore()
	sv, err := VersionForNetwork(buildconstants.TestNetworkVersion)
	if err != nil {
		t.Fatal(err)
	}

	st, err := NewStateTree(cst, sv)
	if err != nil {
		t.Fatal(err)
	}

	a, err := address.NewIDAddress(uint64(222))
	if err != nil {
		t.Fatal(err)
	}

	act := &types.Actor{
		Balance: types.NewInt(0),
		Code:    builtin2.StorageMinerActorCodeID,
		Head:    builtin2.AccountActorCodeID,
		Nonce:   0,
	}

	err = st.SetActor(a, act)
	if err != nil {
		t.Fatal(err)
	}

	act.Nonce = 1

	outact, err := st.GetActor(a)
	if err != nil {
		t.Fatal(err)
	}

	if outact.Nonce == 1 {
		t.Error("nonce should not have updated")
	}
}

func TestSnapshots(t *testing.T) {
	ctx := context.Background()
	cst := cbor.NewMemCborStore()

	sv, err := VersionForNetwork(buildconstants.TestNetworkVersion)
	if err != nil {
		t.Fatal(err)
	}

	st, err := NewStateTree(cst, sv)
	if err != nil {
		t.Fatal(err)
	}

	var addrs []address.Address
	//for _, a := range []string{"t15ocrptbu4i5qucjvvwecihd7fqqgzb27pz5l5zy", "t1dpyvgavvl3f4ujlk6odedss54z6rt5gyuknsuva", "t1feiejbkcvozy7iltt2pxzuoq4d2kpbsusugan7a", "t3rgjfqybjx7bahuhfv7nwfg3tlm4i4zyvldfirjvzm5z5xwjoqbj3rfi2mpmlxpqwxxxafgpkjilqzpg7cefa"} {
	for _, a := range []string{"t0100", "t0101", "t0102", "t0103"} {
		addr, err := address.NewFromString(a)
		if err != nil {
			t.Fatal(err)
		}
		addrs = append(addrs, addr)
	}

	if err := st.Snapshot(ctx); err != nil {
		t.Fatal(err)
	}

	if err := st.SetActor(addrs[0], &types.Actor{Code: builtin2.AccountActorCodeID, Head: builtin2.AccountActorCodeID, Balance: types.NewInt(55)}); err != nil {
		t.Fatal(err)
	}

	{ // sub call that will fail
		if err := st.Snapshot(ctx); err != nil {
			t.Fatal(err)
		}

		if err := st.SetActor(addrs[1], &types.Actor{Code: builtin2.AccountActorCodeID, Head: builtin2.AccountActorCodeID, Balance: types.NewInt(77)}); err != nil {
			t.Fatal(err)
		}

		if err := st.Revert(); err != nil {
			t.Fatal(err)
		}
		st.ClearSnapshot()
	}

	// more operations in top level call...
	if err := st.SetActor(addrs[2], &types.Actor{Code: builtin2.AccountActorCodeID, Head: builtin2.AccountActorCodeID, Balance: types.NewInt(123)}); err != nil {
		t.Fatal(err)
	}

	{ // sub call that succeeds
		if err := st.Snapshot(ctx); err != nil {
			t.Fatal(err)
		}

		if err := st.SetActor(addrs[3], &types.Actor{Code: builtin2.AccountActorCodeID, Head: builtin2.AccountActorCodeID, Balance: types.NewInt(5)}); err != nil {
			t.Fatal(err)
		}

		st.ClearSnapshot()
	}

	st.ClearSnapshot()

	if _, err := st.Flush(ctx); err != nil {
		t.Fatal(err)
	}

	assertHas(t, st, addrs[0])
	assertNotHas(t, st, addrs[1])
	assertHas(t, st, addrs[2])
	assertHas(t, st, addrs[3])
}

func assertHas(t *testing.T, st *StateTree, addr address.Address) {
	_, err := st.GetActor(addr)
	if err != nil {
		t.Fatal(err)
	}
}

func assertNotHas(t *testing.T, st *StateTree, addr address.Address) {
	_, err := st.GetActor(addr)
	if err == nil {
		t.Fatal("shouldnt have found actor", addr)
	}
}

func TestStateTreeConsistency(t *testing.T) {
	cst := cbor.NewMemCborStore()

	// TODO: ActorUpgrade: this test tests pre actors v2

	sv, err := VersionForNetwork(network.Version3)
	if err != nil {
		t.Fatal(err)
	}

	st, err := NewStateTree(cst, sv)
	if err != nil {
		t.Fatal(err)
	}

	var addrs []address.Address
	for i := 100; i < 150; i++ {
		a, err := address.NewIDAddress(uint64(i))
		if err != nil {
			t.Fatal(err)
		}

		addrs = append(addrs, a)
	}

	randomCid, err := cid.Decode("bafy2bzacecu7n7wbtogznrtuuvf73dsz7wasgyneqasksdblxupnyovmtwxxu")
	if err != nil {
		t.Fatal(err)
	}

	for i, a := range addrs {
		err := st.SetActor(a, &types.Actor{
			Code:    randomCid,
			Head:    randomCid,
			Balance: types.NewInt(uint64(10000 + i)),
			Nonce:   uint64(1000 - i),
		})
		if err != nil {
			t.Fatalf("while setting actor: %+v", err)
		}
	}

	root, err := st.Flush(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("root is: ", root)
	if root.String() != "bafy2bzaceb2bhqw75pqp44efoxvlnm73lnctq6djair56bfn5x3gw56epcxbi" {
		t.Fatal("MISMATCH!")
	}
}

// referenceFullScanDiff replicates the exhaustive full-scan semantics of Diff
// (iterate every actor in newTree, look it up in oldTree, emit it valued by its
// new state if added or modified). It is the oracle the fast path must match.
func referenceFullScanDiff(t *testing.T, oldTree, newTree *StateTree) map[string]types.Actor {
	out := map[string]types.Actor{}
	var ncval, ocval cbg.Deferred
	buf := bytes.NewReader(nil)
	err := newTree.root.ForEach(&ncval, func(k string) error {
		addr, err := address.NewFromBytes([]byte(k))
		require.NoError(t, err)

		found, err := oldTree.root.Get(abi.AddrKey(addr), &ocval)
		require.NoError(t, err)
		if found && bytes.Equal(ocval.Raw, ncval.Raw) {
			return nil
		}

		var act types.Actor
		buf.Reset(ncval.Raw)
		require.NoError(t, act.UnmarshalCBOR(buf))
		buf.Reset(nil)
		out[addr.String()] = act
		return nil
	})
	require.NoError(t, err)
	return out
}

// TestDiffFastPathMatchesFullScan verifies that the hamt.Diff-based fast path in
// Diff returns exactly the same data as the exhaustive full scan, across adds,
// modifications, deletions, and unchanged actors.
func TestDiffFastPathMatchesFullScan(t *testing.T) {
	ctx := context.Background()
	cst := cbor.NewMemCborStore()

	mkAddr := func(i uint64) address.Address {
		a, err := address.NewIDAddress(i)
		require.NoError(t, err)
		return a
	}

	// Build an old version-5 state tree with enough actors to span multiple HAMT
	// levels, so structural sharing is actually exercised.
	oldSt, err := NewStateTree(cst, types.StateTreeVersion5)
	require.NoError(t, err)

	const n = 4000
	for i := uint64(1); i <= n; i++ {
		require.NoError(t, oldSt.SetActor(mkAddr(i), &types.Actor{
			Code:    builtin2.AccountActorCodeID,
			Head:    builtin2.AccountActorCodeID,
			Nonce:   i,
			Balance: types.NewInt(1000 + i),
		}))
	}
	oldRoot, err := oldSt.Flush(ctx)
	require.NoError(t, err)

	// Derive a new tree: modify a few, add a few, delete a few.
	newSt, err := LoadStateTree(cst, oldRoot)
	require.NoError(t, err)

	modified := []uint64{1, 50, 999, 2500, n}
	for _, i := range modified {
		require.NoError(t, newSt.SetActor(mkAddr(i), &types.Actor{
			Code:    builtin2.AccountActorCodeID,
			Head:    builtin2.AccountActorCodeID,
			Nonce:   i + 100000,
			Balance: types.NewInt(7777),
		}))
	}
	added := []uint64{n + 1, n + 2, n + 3, n + 7, n + 9}
	for _, i := range added {
		require.NoError(t, newSt.SetActor(mkAddr(i), &types.Actor{
			Code:    builtin2.AccountActorCodeID,
			Head:    builtin2.AccountActorCodeID,
			Nonce:   i,
			Balance: types.NewInt(42),
		}))
	}
	deleted := []uint64{2, 3, 1234}
	for _, i := range deleted {
		require.NoError(t, newSt.DeleteActor(mkAddr(i)))
	}
	newRoot, err := newSt.Flush(ctx)
	require.NoError(t, err)

	oldLoaded, err := LoadStateTree(cst, oldRoot)
	require.NoError(t, err)
	newLoaded, err := LoadStateTree(cst, newRoot)
	require.NoError(t, err)
	require.Equal(t, types.StateTreeVersion5, newLoaded.version)

	reference := referenceFullScanDiff(t, oldLoaded, newLoaded)

	// The fast path in isolation must equal the oracle.
	fast, err := diffActorTreeFast(ctx, oldLoaded, newLoaded)
	require.NoError(t, err)
	require.Equal(t, reference, fast)

	// The public Diff (which now routes v5 trees through the fast path) too.
	got, err := Diff(ctx, oldLoaded, newLoaded)
	require.NoError(t, err)
	require.Equal(t, reference, got)

	// Spot-check the contract: adds + modifications present, deletions absent.
	require.Len(t, fast, len(modified)+len(added))
	for _, i := range modified {
		require.Contains(t, fast, mkAddr(i).String())
	}
	for _, i := range added {
		require.Contains(t, fast, mkAddr(i).String())
	}
	for _, i := range deleted {
		require.NotContains(t, fast, mkAddr(i).String())
	}
}
