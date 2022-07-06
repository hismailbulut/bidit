package bidutil

import (
	"strconv"
	"sync/atomic"
)

type AtomicInt int32

func (atomicInt *AtomicInt) Get() int {
	return int(atomic.LoadInt32((*int32)(atomicInt)))
}

func (atomicInt *AtomicInt) Set(v int) {
	atomic.StoreInt32((*int32)(atomicInt), int32(v))
}

func (atomicInt AtomicInt) Add(v int) AtomicInt {
	ai2 := AtomicInt(0)
	ai2.Set(atomicInt.Get() + v)
	return ai2
}

func (atomicInt AtomicInt) Sub(v int) AtomicInt {
	ai2 := AtomicInt(0)
	ai2.Set(atomicInt.Get() - v)
	return ai2
}

func (atomicInt *AtomicInt) Increment() {
	atomicInt.Set(atomicInt.Get() + 1)
}

func (atomicInt *AtomicInt) Decrement() {
	atomicInt.Set(atomicInt.Get() - 1)
}

func (atomicInt AtomicInt) String() string {
	return strconv.Itoa(atomicInt.Get())
}

type AtomicBool int32

func (atomicBool *AtomicBool) Get() bool {
	v := atomic.LoadInt32((*int32)(atomicBool))
	if v == 1 {
		return true
	} else if v == 0 {
		return false
	}
	panic("unknown boolean value")
}

func (atomicBool *AtomicBool) Set(v bool) {
	if v {
		atomic.StoreInt32((*int32)(atomicBool), 1)
	} else {
		atomic.StoreInt32((*int32)(atomicBool), 0)
	}
}

func (atomicBool AtomicBool) String() string {
	if atomicBool.Get() {
		return "true"
	}
	return "false"
}
