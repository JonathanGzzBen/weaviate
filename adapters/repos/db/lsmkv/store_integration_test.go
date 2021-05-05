package lsmkv

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreLifecycle(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	dirName := fmt.Sprintf("./testdata/%d", rand.Intn(10000000))
	os.MkdirAll(dirName, 0o777)
	defer func() {
		err := os.RemoveAll(dirName)
		fmt.Println(err)
	}()

	t.Run("cycle 1", func(t *testing.T) {
		store, err := New(dirName)
		require.Nil(t, err)

		err = store.CreateOrLoadBucket("bucket1", StrategyReplace)
		require.Nil(t, err)

		b1 := store.Bucket("bucket1")
		require.NotNil(t, b1)

		err = b1.Put([]byte("name"), []byte("Jane Doe"))
		require.Nil(t, err)

		err = store.CreateOrLoadBucket("bucket2", StrategyReplace)
		require.Nil(t, err)

		b2 := store.Bucket("bucket2")
		require.NotNil(t, b2)

		err = b2.Put([]byte("foo"), []byte("bar"))
		require.Nil(t, err)

		err = store.Shutdown(context.Background())
		require.Nil(t, err)
	})

	t.Run("cycle 2", func(t *testing.T) {
		store, err := New(dirName)
		require.Nil(t, err)

		err = store.CreateOrLoadBucket("bucket1", StrategyReplace)
		require.Nil(t, err)

		b1 := store.Bucket("bucket1")
		require.NotNil(t, b1)

		err = store.CreateOrLoadBucket("bucket2", StrategyReplace)
		require.Nil(t, err)

		b2 := store.Bucket("bucket2")
		require.NotNil(t, b2)

		res, err := b1.Get([]byte("name"))
		require.Nil(t, err)
		assert.Equal(t, []byte("Jane Doe"), res)

		res, err = b2.Get([]byte("foo"))
		require.Nil(t, err)
		assert.Equal(t, []byte("bar"), res)

		err = store.Shutdown(context.Background())
		require.Nil(t, err)
	})
}