package root

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	coreheader "cosmossdk.io/core/header"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/sqlite"
)

const (
	testStoreKey  = "test_store_key"
	testStoreKey2 = "test_store_key2"
	testStoreKey3 = "test_store_key3"
)

var (
	testStoreKeyBytes  = []byte(testStoreKey)
	testStoreKey2Bytes = []byte(testStoreKey2)
	testStoreKey3Bytes = []byte(testStoreKey3)
)

type RootStoreTestSuite struct {
	suite.Suite

	rootStore store.RootStore
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, &RootStoreTestSuite{})
}

func (s *RootStoreTestSuite) SetupTest() {
	noopLog := log.NewNopLogger()

	sqliteDB, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)
	ss := storage.NewStorageStore(sqliteDB, nil, noopLog)

	tree := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	tree2 := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	tree3 := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	sc, err := commitment.NewCommitStore(map[string]commitment.Tree{testStoreKey: tree, testStoreKey2: tree2, testStoreKey3: tree3}, dbm.NewMemDB(), nil, noopLog)
	s.Require().NoError(err)

	rs, err := New(noopLog, ss, sc, nil, nil)
	s.Require().NoError(err)

	s.rootStore = rs
}

func (s *RootStoreTestSuite) newStoreWithPruneConfig(config *store.PruneOptions) {
	noopLog := log.NewNopLogger()

	sqliteDB, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)
	ss := storage.NewStorageStore(sqliteDB, config, noopLog)

	tree := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	tree2 := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	tree3 := iavl.NewIavlTree(dbm.NewMemDB(), noopLog, iavl.DefaultConfig())
	sc, err := commitment.NewCommitStore(map[string]commitment.Tree{testStoreKey: tree, testStoreKey2: tree2, testStoreKey3: tree3}, dbm.NewMemDB(), config, noopLog)
	s.Require().NoError(err)

	rs, err := New(noopLog, ss, sc, nil, nil)
	s.Require().NoError(err)

	s.rootStore = rs
}

func (s *RootStoreTestSuite) newStoreWithDBMount(config *store.PruneOptions, dbm1, dbm2, dbm3, dbm4 *dbm.MemDB) {
	noopLog := log.NewNopLogger()

	sqliteDB, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)
	ss := storage.NewStorageStore(sqliteDB, config, noopLog)

	tree := iavl.NewIavlTree(dbm1, noopLog, iavl.DefaultConfig())
	tree2 := iavl.NewIavlTree(dbm2, noopLog, iavl.DefaultConfig())
	tree3 := iavl.NewIavlTree(dbm3, noopLog, iavl.DefaultConfig())
	sc, err := commitment.NewCommitStore(map[string]commitment.Tree{testStoreKey: tree, testStoreKey2: tree2, testStoreKey3: tree3}, dbm4, config, noopLog)
	s.Require().NoError(err)

	rs, err := New(noopLog, ss, sc, nil, nil)
	s.Require().NoError(err)

	s.rootStore = rs
}

func (s *RootStoreTestSuite) TearDownTest() {
	err := s.rootStore.Close()
	s.Require().NoError(err)
}

func (s *RootStoreTestSuite) TestGetStateCommitment() {
	s.Require().Equal(s.rootStore.GetStateCommitment(), s.rootStore.(*Store).stateCommitment)
}

func (s *RootStoreTestSuite) TestGetStateStorage() {
	s.Require().Equal(s.rootStore.GetStateStorage(), s.rootStore.(*Store).stateStorage)
}

func (s *RootStoreTestSuite) TestSetInitialVersion() {
	s.Require().NoError(s.rootStore.SetInitialVersion(100))
}

func (s *RootStoreTestSuite) TestSetCommitHeader() {
	h := &coreheader.Info{
		Height:  100,
		Hash:    []byte("foo"),
		ChainID: "test",
	}
	s.rootStore.SetCommitHeader(h)

	s.Require().Equal(h, s.rootStore.(*Store).commitHeader)
}

func (s *RootStoreTestSuite) TestQuery() {
	_, err := s.rootStore.Query([]byte{}, 1, []byte("foo"), true)
	s.Require().Error(err)

	// write and commit a changeset
	cs := corestore.NewChangeset()
	cs.Add(testStoreKeyBytes, []byte("foo"), []byte("bar"), false)

	workingHash, err := s.rootStore.WorkingHash(cs)
	s.Require().NoError(err)
	s.Require().NotNil(workingHash)

	commitHash, err := s.rootStore.Commit(cs)
	s.Require().NoError(err)
	s.Require().NotNil(commitHash)
	s.Require().Equal(workingHash, commitHash)

	// ensure the proof is non-nil for the corresponding version
	result, err := s.rootStore.Query([]byte(testStoreKey), 1, []byte("foo"), true)
	s.Require().NoError(err)
	s.Require().NotNil(result.ProofOps)
	s.Require().Equal([]byte("foo"), result.ProofOps[0].Key)
}

func (s *RootStoreTestSuite) TestGetFallback() {
	sc := s.rootStore.GetStateCommitment()

	// create a changeset and commit it to SC ONLY
	cs := corestore.NewChangeset()
	cs.Add(testStoreKeyBytes, []byte("foo"), []byte("bar"), false)

	err := sc.WriteBatch(cs)
	s.Require().NoError(err)

	ci := sc.WorkingCommitInfo(1)
	_, err = sc.Commit(ci.Version)
	s.Require().NoError(err)

	// ensure we can query for the key, which should fallback to SC
	qResult, err := s.rootStore.Query(testStoreKeyBytes, 1, []byte("foo"), false)
	s.Require().NoError(err)
	s.Require().Equal([]byte("bar"), qResult.Value)

	// non-existent key
	qResult, err = s.rootStore.Query(testStoreKeyBytes, 1, []byte("non_existent_key"), false)
	s.Require().NoError(err)
	s.Require().Nil(qResult.Value)
}

func (s *RootStoreTestSuite) TestQueryProof() {
	cs := corestore.NewChangeset()
	// testStoreKey
	cs.Add(testStoreKeyBytes, []byte("key1"), []byte("value1"), false)
	cs.Add(testStoreKeyBytes, []byte("key2"), []byte("value2"), false)
	// testStoreKey2
	cs.Add(testStoreKey2Bytes, []byte("key3"), []byte("value3"), false)
	// testStoreKey3
	cs.Add(testStoreKey3Bytes, []byte("key4"), []byte("value4"), false)

	// commit
	_, err := s.rootStore.WorkingHash(cs)
	s.Require().NoError(err)
	_, err = s.rootStore.Commit(cs)
	s.Require().NoError(err)

	// query proof for testStoreKey
	result, err := s.rootStore.Query(testStoreKeyBytes, 1, []byte("key1"), true)
	s.Require().NoError(err)
	s.Require().NotNil(result.ProofOps)
	cInfo, err := s.rootStore.GetStateCommitment().GetCommitInfo(1)
	s.Require().NoError(err)
	storeHash := cInfo.GetStoreCommitID(testStoreKeyBytes).Hash
	treeRoots, err := result.ProofOps[0].Run([][]byte{[]byte("value1")})
	s.Require().NoError(err)
	s.Require().Equal(treeRoots[0], storeHash)
	expRoots, err := result.ProofOps[1].Run([][]byte{storeHash})
	s.Require().NoError(err)
	s.Require().Equal(expRoots[0], cInfo.Hash())
}

func (s *RootStoreTestSuite) TestLoadVersion() {
	// write and commit a few changesets
	for v := 1; v <= 5; v++ {
		val := fmt.Sprintf("val%03d", v) // val001, val002, ..., val005

		cs := corestore.NewChangeset()
		cs.Add(testStoreKeyBytes, []byte("key"), []byte(val), false)

		workingHash, err := s.rootStore.WorkingHash(cs)
		s.Require().NoError(err)
		s.Require().NotNil(workingHash)

		commitHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(commitHash)
		s.Require().Equal(workingHash, commitHash)
	}

	// ensure the latest version is correct
	latest, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(5), latest)

	// attempt to load a non-existent version
	err = s.rootStore.LoadVersion(6)
	s.Require().Error(err)

	// attempt to load a previously committed version
	err = s.rootStore.LoadVersion(3)
	s.Require().NoError(err)

	// ensure the latest version is correct
	latest, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), latest)

	// query state and ensure values returned are based on the loaded version
	_, ro, err := s.rootStore.StateLatest()
	s.Require().NoError(err)

	reader, err := ro.GetReader(testStoreKeyBytes)
	s.Require().NoError(err)
	val, err := reader.Get([]byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("val003"), val)

	// attempt to write and commit a few changesets
	for v := 4; v <= 5; v++ {
		val := fmt.Sprintf("overwritten_val%03d", v) // overwritten_val004, overwritten_val005

		cs := corestore.NewChangeset()
		cs.Add(testStoreKeyBytes, []byte("key"), []byte(val), false)

		workingHash, err := s.rootStore.WorkingHash(cs)
		s.Require().NoError(err)
		s.Require().NotNil(workingHash)

		commitHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().NotNil(commitHash)
		s.Require().Equal(workingHash, commitHash)
	}

	// ensure the latest version is correct
	latest, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(5), latest)

	// query state and ensure values returned are based on the loaded version
	_, ro, err = s.rootStore.StateLatest()
	s.Require().NoError(err)

	reader, err = ro.GetReader(testStoreKeyBytes)
	s.Require().NoError(err)
	val, err = reader.Get([]byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("overwritten_val005"), val)
}

func (s *RootStoreTestSuite) TestCommit() {
	lv, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Zero(lv)

	// perform changes
	cs := corestore.NewChangeset()
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)
	}

	// committing w/o calling WorkingHash should error
	_, err = s.rootStore.Commit(cs)
	s.Require().Error(err)

	// execute WorkingHash and Commit
	wHash, err := s.rootStore.WorkingHash(cs)
	s.Require().NoError(err)

	cHash, err := s.rootStore.Commit(cs)
	s.Require().NoError(err)
	s.Require().Equal(wHash, cHash)

	// ensure latest version is updated
	lv, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), lv)

	// perform reads on the updated root store
	_, ro, err := s.rootStore.StateLatest()
	s.Require().NoError(err)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		reader, err := ro.GetReader(testStoreKeyBytes)
		s.Require().NoError(err)
		result, err := reader.Get([]byte(key))
		s.Require().NoError(err)

		s.Require().Equal([]byte(val), result)
	}
}

func (s *RootStoreTestSuite) TestStateAt() {
	// write keys over multiple versions
	for v := uint64(1); v <= 5; v++ {
		// perform changes
		cs := corestore.NewChangeset()
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
			val := fmt.Sprintf("val%03d_%03d", i, v) // val000_1, val001_1, ..., val099_1

			cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)
		}

		// execute WorkingHash and Commit
		wHash, err := s.rootStore.WorkingHash(cs)
		s.Require().NoError(err)

		cHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().Equal(wHash, cHash)
	}

	lv, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(5), lv)

	// ensure we can read state correctly at each version
	for v := uint64(1); v <= 5; v++ {
		ro, err := s.rootStore.StateAt(v)
		s.Require().NoError(err)

		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
			val := fmt.Sprintf("val%03d_%03d", i, v) // val000_1, val001_1, ..., val099_1

			reader, err := ro.GetReader(testStoreKeyBytes)
			s.Require().NoError(err)
			result, err := reader.Get([]byte(key))
			s.Require().NoError(err)
			s.Require().Equal([]byte(val), result)
		}
	}
}

func (s *RootStoreTestSuite) TestPrune() {
	// perform changes
	cs := corestore.NewChangeset()
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)
	}

	testCases := []struct {
		name        string
		numVersions int64
		po          store.PruneOptions
		deleted     []uint64
		saved       []uint64
	}{
		{"prune nothing", 10, *store.DefaultPruneOptions(), nil, []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"prune everything", 12, store.PruneOptions{
			KeepRecent: 1,
			Interval:   10,
		}, []uint64{1, 2, 3, 4, 5, 6, 7, 8}, []uint64{9, 10, 11, 12}},
		{"prune some; no batch", 10, store.PruneOptions{
			KeepRecent: 2,
			Interval:   1,
		}, []uint64{1, 2, 3, 4, 6, 5, 7}, []uint64{8, 9, 10}},
		{"prune some; small batch", 10, store.PruneOptions{
			KeepRecent: 2,
			Interval:   3,
		}, []uint64{1, 2, 3, 4, 5, 6}, []uint64{7, 8, 9, 10}},
		{"prune some; large batch", 10, store.PruneOptions{
			KeepRecent: 2,
			Interval:   11,
		}, nil, []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	for _, tc := range testCases {
		tc := tc

		s.newStoreWithPruneConfig(&tc.po)

		// write keys over multiple versions
		for i := int64(0); i < tc.numVersions; i++ {
			// execute WorkingHash and Commit
			wHash, err := s.rootStore.WorkingHash(cs)
			s.Require().NoError(err)

			cHash, err := s.rootStore.Commit(cs)
			s.Require().NoError(err)
			s.Require().Equal(wHash, cHash)
		}

		for _, v := range tc.saved {
			ro, err := s.rootStore.StateAt(v)
			s.Require().NoError(err, "expected no error when loading height %d at test %s", v, tc.name)

			for i := 0; i < 10; i++ {
				key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
				val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

				reader, err := ro.GetReader(testStoreKeyBytes)
				s.Require().NoError(err)
				result, err := reader.Get([]byte(key))
				s.Require().NoError(err)
				s.Require().Equal([]byte(val), result, "value should be equal for test: %s", tc.name)
			}
		}

		for _, v := range tc.deleted {
			_, err := s.rootStore.StateAt(v)
			s.Require().Error(err, "expected error when loading height %d at test %s", v, tc.name)
		}
	}

}

func (s *RootStoreTestSuite) TestMultiStore_Pruning_SameHeightsTwice() {
	// perform changes
	cs := corestore.NewChangeset()
	cs.Add(testStoreKeyBytes, []byte("key"), []byte("val"), false)

	const (
		numVersions uint64 = 10
		keepRecent  uint64 = 1
		interval    uint64 = 10
	)

	s.newStoreWithPruneConfig(&store.PruneOptions{
		KeepRecent: keepRecent,
		Interval:   interval,
	})
	s.Require().NoError(s.rootStore.LoadLatestVersion())

	for i := uint64(0); i < numVersions; i++ {
		// execute WorkingHash and Commit
		wHash, err := s.rootStore.WorkingHash(cs)
		s.Require().NoError(err)

		cHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().Equal(wHash, cHash)
	}

	latestVer, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(numVersions, latestVer)

	for v := uint64(1); v < numVersions-uint64(keepRecent); v++ {
		err := s.rootStore.LoadVersion(v)
		s.Require().Error(err, "expected error when loading pruned height: %d", v)
	}

	for v := (numVersions - uint64(keepRecent)); v < numVersions; v++ {
		err := s.rootStore.LoadVersion(v)
		s.Require().NoError(err, "expected no error when loading height: %d", v)
	}

	// Get latest
	err = s.rootStore.LoadVersion(numVersions - 1)
	s.Require().NoError(err)

	// Test pruning the same heights again
	_, err = s.rootStore.WorkingHash(cs)
	s.Require().NoError(err)
	_, err = s.rootStore.Commit(cs)
	s.Require().NoError(err)

	// Ensure that can commit one more height with no panic
	_, err = s.rootStore.WorkingHash(cs)
	s.Require().NoError(err)
	_, err = s.rootStore.Commit(cs)
	s.Require().NoError(err)
}

func (s *RootStoreTestSuite) TestMultiStore_PruningRestart() {
	// perform changes
	cs := corestore.NewChangeset()
	cs.Add(testStoreKeyBytes, []byte("key"), []byte("val"), false)

	pruneOpt := store.PruneOptions{
		KeepRecent: 2,
		Interval:   11,
	}

	db1 := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	db3 := dbm.NewMemDB()
	db4 := dbm.NewMemDB()

	s.newStoreWithDBMount(&pruneOpt, db1, db2, db3, db4)
	s.Require().NoError(s.rootStore.LoadLatestVersion())

	// Commit enough to build up heights to prune, where on the next block we should
	// batch delete.
	for i := uint64(0); i < 10; i++ {
		// execute WorkingHash and Commit
		wHash, err := s.rootStore.WorkingHash(cs)
		s.Require().NoError(err)

		cHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().Equal(wHash, cHash)
	}

	latestVer, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)

	ok, actualHeightToPrune := pruneOpt.ShouldPrune(latestVer)
	s.Require().False(ok)
	s.Require().Equal(uint64(0), actualHeightToPrune)

	// "restart"
	s.newStoreWithDBMount(&pruneOpt, db1, db2, db3, db4)
	err = s.rootStore.LoadLatestVersion()
	s.Require().NoError(err)

	latestVer, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)

	ok, actualHeightToPrune = pruneOpt.ShouldPrune(latestVer)
	s.Require().False(ok)
	s.Require().Equal(uint64(0), actualHeightToPrune)

	// commit one more block and ensure the heights have been pruned
	// execute WorkingHash and Commit
	wHash, err := s.rootStore.WorkingHash(cs)
	s.Require().NoError(err)

	cHash, err := s.rootStore.Commit(cs)
	s.Require().NoError(err)
	s.Require().Equal(wHash, cHash)

	latestVer, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)

	ok, actualHeightToPrune = pruneOpt.ShouldPrune(latestVer)
	s.Require().True(ok)
	s.Require().Equal(uint64(8), actualHeightToPrune)

	for v := uint64(1); v <= actualHeightToPrune; v++ {
		err := s.rootStore.LoadVersion(v)
		s.Require().Error(err, "expected error when loading height: %d", v)
	}
}

func (s *RootStoreTestSuite) TestMultiStoreRestart() {
	db1 := dbm.NewMemDB()
	db2 := dbm.NewMemDB()
	db3 := dbm.NewMemDB()
	db4 := dbm.NewMemDB()

	s.newStoreWithDBMount(nil, db1, db2, db3, db4)
	s.Require().NoError(s.rootStore.LoadLatestVersion())

	// perform changes
	for i := uint64(1); i < 3; i++ {
		cs := corestore.NewChangeset()
		key := fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d_%03d", i, 1) // val000_1, val001_1, ..., val099_1

		cs.Add(testStoreKeyBytes, []byte(key), []byte(val), false)

		key = fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
		val = fmt.Sprintf("val%03d_%03d", i, 2) // val000_1, val001_1, ..., val099_1

		cs.Add(testStoreKey2Bytes, []byte(key), []byte(val), false)

		key = fmt.Sprintf("key%03d", i)         // key000, key001, ..., key099
		val = fmt.Sprintf("val%03d_%03d", i, 3) // val000_1, val001_1, ..., val099_1

		cs.Add(testStoreKey3Bytes, []byte(key), []byte(val), false)

		// execute WorkingHash and Commit
		wHash, err := s.rootStore.WorkingHash(cs)
		s.Require().NoError(err)

		cHash, err := s.rootStore.Commit(cs)
		s.Require().NoError(err)
		s.Require().Equal(wHash, cHash)

		latestVer, err := s.rootStore.GetLatestVersion()
		s.Require().NoError(err)
		s.Require().Equal(i, latestVer)
	}

	// more changes
	cs1 := corestore.NewChangeset()
	key := fmt.Sprintf("key%03d", 3)         // key000, key001, ..., key099
	val := fmt.Sprintf("val%03d_%03d", 3, 1) // val000_1, val001_1, ..., val099_1

	cs1.Add(testStoreKeyBytes, []byte(key), []byte(val), false)

	key = fmt.Sprintf("key%03d", 3)         // key000, key001, ..., key099
	val = fmt.Sprintf("val%03d_%03d", 3, 2) // val000_1, val001_1, ..., val099_1

	cs1.Add(testStoreKey2Bytes, []byte(key), []byte(val), false)

	// execute WorkingHash and Commit
	wHash, err := s.rootStore.WorkingHash(cs1)
	s.Require().NoError(err)

	cHash, err := s.rootStore.Commit(cs1)
	s.Require().NoError(err)
	s.Require().Equal(wHash, cHash)

	latestVer, err := s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), latestVer)

	cs2 := corestore.NewChangeset()
	key = fmt.Sprintf("key%03d", 4)         // key000, key001, ..., key099
	val = fmt.Sprintf("val%03d_%03d", 4, 3) // val000_1, val001_1, ..., val099_1

	cs2.Add(testStoreKey3Bytes, []byte(key), []byte(val), false)

	// execute WorkingHash and Commit
	wHash, err = s.rootStore.WorkingHash(cs2)
	s.Require().NoError(err)

	cHash, err = s.rootStore.Commit(cs2)
	s.Require().NoError(err)
	s.Require().Equal(wHash, cHash)

	latestVer, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), latestVer)

	// "restart"
	s.newStoreWithDBMount(nil, db1, db2, db3, db4)
	err = s.rootStore.LoadLatestVersion()
	s.Require().Nil(err)

	latestVer, err = s.rootStore.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), latestVer)

	_, ro, err := s.rootStore.StateLatest()
	s.Require().Nil(err)
	reader, err := ro.GetReader(testStoreKeyBytes)
	s.Require().NoError(err)
	result, err := reader.Get([]byte(fmt.Sprintf("key%03d", 3)))
	s.Require().NoError(err)
	s.Require().Equal([]byte(fmt.Sprintf("val%03d_%03d", 3, 1)), result, "value should be equal")

	reader, err = ro.GetReader(testStoreKey2Bytes)
	s.Require().NoError(err)
	result, err = reader.Get([]byte(fmt.Sprintf("key%03d", 2)))
	s.Require().NoError(err)
	s.Require().Equal([]byte(fmt.Sprintf("val%03d_%03d", 4, 3)), result, "value should be equal")

	reader, err = ro.GetReader(testStoreKey3Bytes)
	s.Require().NoError(err)
	result, err = reader.Get([]byte(fmt.Sprintf("key%03d", 4)))
	s.Require().NoError(err)
	s.Require().Equal([]byte(fmt.Sprintf("val%03d_%03d", 2, 2)), result, "value should be equal")
}

func (s *RootStoreTestSuite) TestUnevenStoresHeightCheck() {

}