package opencl

// #include "opencl.h"
import "C"
import (
	"context"
	"unsafe"

	pb "github.com/zbsss/device-manager/generated"
)

type MemFlags uint64

const (
	MemReadWrite MemFlags = C.CL_MEM_READ_WRITE
	MemWriteOnly          = C.CL_MEM_WRITE_ONLY
	MemReadOnly           = C.CL_MEM_READ_ONLY
	// ...
)

type Buffer struct {
	buffer C.cl_mem
}

func createBuffer(context Context, flags []MemFlags, size uint64) (Buffer, error) {
	// AND together all flags
	flagBitField := uint64(0)
	for _, flag := range flags {
		flagBitField &= uint64(flag)
	}

	var errInt clError
	buffer := C.clCreateBuffer(
		context.context,
		C.cl_mem_flags(flagBitField),
		C.size_t(size),
		nil,
		(*C.cl_int)(&errInt),
	)
	if errInt != clSuccess {
		return Buffer{}, clErrorToError(errInt)
	}

	return Buffer{buffer}, nil
}

func (b Buffer) Size() uint64 {
	return uint64(C.sizeof_cl_mem)
}

func (b Buffer) Release() {
	var size C.size_t
	C.clGetMemObjectInfo(b.buffer, C.CL_MEM_SIZE, C.size_t(unsafe.Sizeof(size)), unsafe.Pointer(&size), nil)

	C.clReleaseMemObject(b.buffer)

	ctx := context.Background()
	Scheduler.ReturnMemoryQuota(ctx, &pb.ReturnMemoryQuotaRequest{Device: DeviceId, Pod: ClientId, Memory: uint64(size)})
}
