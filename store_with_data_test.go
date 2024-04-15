package wasmtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreWithData(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()
	store := NewStoreWithData[any](engine, nil)
	defer store.Close()
}

func TestStoreWithDataInterruptWasm(t *testing.T) {
	config := NewConfig()
	config.SetEpochInterruption(true)
	store := NewStoreWithData[any](NewEngineWithConfig(config), nil)
	store.SetEpochDeadline(1)
	wasm, err := Wat2Wasm(`
	  (import "" "" (func))
	  (func
	    call 0
	    (loop br 0))
	  (start 1)
	`)
	require.NoError(t, err)
	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)
	engine := store.Engine
	f := WrapFunc(store, func() {
		engine.IncrementEpoch()
	})
	instance, err := NewInstance(store, module, []AsExtern{f})
	require.Nil(t, instance)

	require.Error(t, err)
	trap := err.(*Trap)
	require.NotNil(t, trap)
}

func TestStoreWithDataFuelConsumed(t *testing.T) {
	engine := NewEngine()
	store := NewStoreWithData[any](engine, nil)

	fuel, enable := store.GetFuel()
	require.Error(t, enable)
	require.Equal(t, fuel, uint64(0))
}

func TestStoreWithDataAddFuel(t *testing.T) {
	config := NewConfig()
	config.SetConsumeFuel(true)
	engine := NewEngineWithConfig(config)
	store := NewStoreWithData[any](engine, nil)

	fuel, enable := store.GetFuel()
	require.NoError(t, enable)
	require.Equal(t, fuel, uint64(0))

	const add_fuel = 3
	err := store.SetFuel(add_fuel)
	require.NoError(t, err)
}

func TestStoreWithDataLimiterMemorySizeFail(t *testing.T) {
	engine := NewEngine()
	store := NewStoreWithData[any](engine, nil)

	store.Limiter(2*64*1024, -1, -1, -1, -1)
	wasm, err := Wat2Wasm(`
	(module
	  (memory 3)
	)
	`)
	require.NoError(t, err)

	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	_, err = NewInstance(store, module, []AsExtern{})
	require.Error(t, err, "memory minimum size of 3 pages exceeds memory limits")
}

func TestStoreWithDataLimiterMemorySizeSuccess(t *testing.T) {
	engine := NewEngine()
	store := NewStoreWithData[any](engine, nil)

	store.Limiter(4*64*1024, -1, -1, -1, -1)
	wasm, err := Wat2Wasm(`
	(module
	  (memory 3)
	)
	`)
	require.NoError(t, err)

	module, err := NewModule(store.Engine, wasm)
	require.NoError(t, err)

	_, err = NewInstance(store, module, []AsExtern{})
	require.NoError(t, err)
}

func TestStoreWithDataRetrieval(t *testing.T) {
	engine := NewEngine()
	type testStruct struct {
		x int
	}
	data := testStruct{1}
	store := NewStoreWithData[testStruct](engine, data)
	require.Equal(t, data, store.Data())
}
