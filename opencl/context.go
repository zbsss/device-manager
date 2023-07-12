package opencl

// #include "opencl.h"
import "C"
import (
	"context"

	pb "github.com/zbsss/device-manager/generated"
)

type Context struct {
	context C.cl_context
}

func createContext(device Device) (Context, error) {
	var emptyContext Context

	// TODO add more functionality. Super simple context creation right now
	var errInt clError
	ctx := C.clCreateContext(
		nil,
		1,
		(*C.cl_device_id)(&device.deviceID),
		nil,
		nil,
		(*C.cl_int)(&errInt),
	)
	if errInt != clSuccess {
		return emptyContext, clErrorToError(errInt)
	}

	return Context{ctx}, nil
}

func (c Context) CreateCommandQueue(device Device) (CommandQueue, error) {
	return createCommandQueue(c, device)
}

func (c Context) CreateProgramWithSource(programCode string) (Program, error) {
	return createProgramWithSource(c, programCode)
}

func (c Context) CreateBuffer(memFlags []MemFlags, size uint64) (Buffer, error) {
	ctx := context.Background()
	deviceId := "1"
	_, err := Scheduler.GetMemoryQuota(ctx, &pb.GetMemoryQuotaRequest{Device: deviceId, Memory: size})
	if err != nil {
		return Buffer{}, err
	}

	return createBuffer(c, memFlags, size)
}

func (c Context) Release() {
	C.clReleaseContext(c.context)
}
