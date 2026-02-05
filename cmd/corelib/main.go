package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"taskpp/core/bind"
)

var (
	handleSeq uint64
	mu        sync.RWMutex
	cores     = make(map[uint64]*bind.Core)
)

//export Core_New
func Core_New(configJSON *C.char) C.uint64_t {
	cfg := ""
	if configJSON != nil {
		cfg = C.GoString(configJSON)
	}
	core := bind.NewCore(cfg)
	id := atomic.AddUint64(&handleSeq, 1)
	mu.Lock()
	cores[id] = core
	mu.Unlock()
	return C.uint64_t(id)
}

//export Core_Destroy
func Core_Destroy(handle C.uint64_t) {
	id := uint64(handle)
	mu.Lock()
	delete(cores, id)
	mu.Unlock()
}

//export Core_Open
func Core_Open(handle C.uint64_t) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.Open())
}

//export Core_Close
func Core_Close(handle C.uint64_t) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.Close())
}

//export Core_InitKeys
func Core_InitKeys(handle C.uint64_t, passphrase *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.InitKeys(cGoString(passphrase)))
}

//export Core_UnlockKeys
func Core_UnlockKeys(handle C.uint64_t, passphrase *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.UnlockKeys(cGoString(passphrase)))
}

//export Core_ListTasks
func Core_ListTasks(handle C.uint64_t, filterJSON *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.ListTasks(cGoString(filterJSON)))
}

//export Core_CreateTask
func Core_CreateTask(handle C.uint64_t, taskJSON *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.CreateTask(cGoString(taskJSON)))
}

//export Core_UpdateTask
func Core_UpdateTask(handle C.uint64_t, taskJSON *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.UpdateTask(cGoString(taskJSON)))
}

//export Core_DeleteTask
func Core_DeleteTask(handle C.uint64_t, taskID *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.DeleteTask(cGoString(taskID)))
}

//export Core_ReorderTasks
func Core_ReorderTasks(handle C.uint64_t, reorderJSON *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.ReorderTasks(cGoString(reorderJSON)))
}

//export Core_SetDueDate
func Core_SetDueDate(handle C.uint64_t, taskID *C.char, dueDate *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.SetDueDate(cGoString(taskID), cGoString(dueDate)))
}

//export Core_SetCompleted
func Core_SetCompleted(handle C.uint64_t, taskID *C.char, completed C.int) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.SetCompleted(cGoString(taskID), completed != 0))
}

//export Core_ExportEvents
func Core_ExportEvents(handle C.uint64_t, sinceSeq C.longlong) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.ExportEvents(int64(sinceSeq)))
}

//export Core_ImportEvents
func Core_ImportEvents(handle C.uint64_t, eventsJSON *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.ImportEvents(cGoString(eventsJSON)))
}

//export Core_GetSyncState
func Core_GetSyncState(handle C.uint64_t) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.GetSyncState())
}

//export Core_DebugDecryptEvent
func Core_DebugDecryptEvent(handle C.uint64_t, payloadBase64 *C.char) *C.char {
	core := getCore(handle)
	if core == nil {
		return cError("core not found")
	}
	return cString(core.DebugDecryptEvent(cGoString(payloadBase64)))
}

//export Core_FreeString
func Core_FreeString(str *C.char) {
	if str != nil {
		C.free(unsafe.Pointer(str))
	}
}

func getCore(handle C.uint64_t) *bind.Core {
	id := uint64(handle)
	mu.RLock()
	core := cores[id]
	mu.RUnlock()
	return core
}

func cGoString(value *C.char) string {
	if value == nil {
		return ""
	}
	return C.GoString(value)
}

func cString(value string) *C.char {
	return C.CString(value)
}

func cError(message string) *C.char {
	return C.CString(`{"error":"` + message + `"}`)
}

func main() {}
