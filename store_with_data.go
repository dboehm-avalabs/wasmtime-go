package wasmtime

// #include <wasmtime.h>
// #include "shims.h"
import "C"
import (
	"runtime"
)

// StoreWithData is a general group of wasm instances, and many objects
// must all be created with and reference the same `Store`
type StoreWithData[T any] struct {
	Store
}

// StoreWithDatalike represents types that can be used to contextually reference a
// `Store`.
//
// This interface is implemented by `*Store` and `*Caller` and is pervasively
// used throughout this library. You'll want to pass one of those two objects
// into functions that take a `StoreWithDatalike`.
type StoreWithDatalike[T any] interface {
	// Returns the wasmtime context pointer this store is attached to.
	Context() *C.wasmtime_context_t
	Data() T
}

// NewStore creates a new `Store` from the configuration provided in `engine`
func NewStoreWithData[T any](engine *Engine, data T) *StoreWithData[T] {
	// Allocate an index for this store and allocate some internal data to go with
	// the store.
	gStoreLock.Lock()
	idx := gStoreSlab.allocate()
	gStoreMap[idx] = &storeData{engine: engine, data: data}
	gStoreLock.Unlock()

	ptr := C.go_store_new(engine.ptr(), C.size_t(idx))
	store := &StoreWithData[T]{
		Store{
			_ptr:   ptr,
			Engine: engine,
		},
	}
	runtime.SetFinalizer(store, func(store *StoreWithData[T]) {
		store.Close()
	})
	return store
}

// Returns the underlying `data` that this store references
func (store *StoreWithData[T]) Data() T {
	return getDataInStore(store).data.(T)
}
