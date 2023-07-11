package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/zbsss/device-manager/opencl"
)

var kernelSource = `
__kernel void square(
   __global float* input,
   __global float* output,
   const unsigned int count)
{
   int i = get_global_id(0);
   if(i < count)
       output[i] = input[i] * input[i];
}
`

const (
	deviceType = opencl.DeviceTypeAll

	dataSize = 128

	programCode = `
kernel void kern(global float* out, global float* a, global float* b)
{
	size_t i = get_global_id(0);
	out[i] = a[i] + b[i];
}
`
)

func printHeader(name string) {
	fmt.Println(strings.ToUpper(name))
	for _ = range name {
		fmt.Print("=")
	}
	fmt.Println()
}

func printInfo(platform opencl.Platform, device opencl.Device) {
	var platformName string
	err := platform.GetInfo(opencl.PlatformName, &platformName)
	if err != nil {
		panic(err)
	}

	var vendor string
	err = device.GetInfo(opencl.DeviceVendor, &vendor)
	if err != nil {
		panic(err)
	}

	fmt.Println()
	printHeader("Using")
	fmt.Println("Platform:", platformName)
	fmt.Println("Vendor:  ", vendor)
}

func main() {
	platforms, err := opencl.GetPlatforms()
	if err != nil {
		panic(err)
	}

	printHeader("Platforms")

	foundDevice := false

	var platform opencl.Platform
	var device opencl.Device
	var name string
	for _, curPlatform := range platforms {
		err = curPlatform.GetInfo(opencl.PlatformName, &name)
		if err != nil {
			panic(err)
		}

		var devices []opencl.Device
		devices, err = curPlatform.GetDevices(deviceType)
		if err != nil {
			panic(err)
		}

		// Use the first available device
		if len(devices) > 0 && !foundDevice {
			var available bool
			err = devices[0].GetInfo(opencl.DeviceAvailable, &available)
			if err == nil && available {
				platform = curPlatform
				device = devices[0]
				foundDevice = true
			}
		}

		version := curPlatform.GetVersion()
		fmt.Printf("Name: %v, devices: %v, version: %v\n", name, len(devices), version)
	}

	if !foundDevice {
		panic("No device found")
	}

	printInfo(platform, device)

	var ctx opencl.Context
	ctx, err = device.CreateContext()
	if err != nil {
		panic(err)
	}
	defer ctx.Release()

	var commandQueue opencl.CommandQueue
	commandQueue, err = ctx.CreateCommandQueue(device)
	if err != nil {
		panic(err)
	}
	defer commandQueue.Release()

	var program opencl.Program
	program, err = ctx.CreateProgramWithSource(programCode)
	if err != nil {
		panic(err)
	}
	defer program.Release()

	var logs string
	err = program.Build(device, &logs)
	if err != nil {
		fmt.Println(logs)
		panic(err)
	}

	kernel, err := program.CreateKernel("kern")
	if err != nil {
		panic(err)
	}
	defer kernel.Release()

	buffer, err := ctx.CreateBuffer([]opencl.MemFlags{opencl.MemWriteOnly}, dataSize*4)
	if err != nil {
		panic(err)
	}
	defer buffer.Release()

	buffer1, err := ctx.CreateBuffer([]opencl.MemFlags{opencl.MemReadOnly}, dataSize*4)
	if err != nil {
		panic(err)
	}
	defer buffer1.Release()

	buffer2, err := ctx.CreateBuffer([]opencl.MemFlags{opencl.MemReadOnly}, dataSize*4)
	if err != nil {
		panic(err)
	}
	defer buffer2.Release()

	err = kernel.SetArg(0, buffer.Size(), &buffer)
	if err != nil {
		panic(err)
	}

	err = kernel.SetArg(1, buffer1.Size(), &buffer1)
	if err != nil {
		panic(err)
	}

	err = kernel.SetArg(2, buffer2.Size(), &buffer2)
	if err != nil {
		panic(err)
	}

	write_data := make([]float32, dataSize)
	for i := 0; i < dataSize; i++ {
		write_data[i] = float32(i)
	}

	for {
		err = commandQueue.EnqueueWriteBuffer(buffer1, true, write_data)
		if err != nil {
			panic(err)
		}
		err = commandQueue.EnqueueWriteBuffer(buffer2, true, write_data)
		if err != nil {
			panic(err)
		}

		err = commandQueue.EnqueueNDRangeKernel(kernel, 1, []uint64{dataSize})
		if err != nil {
			panic(err)
		}

		data := make([]float32, dataSize)

		err = commandQueue.EnqueueReadBuffer(buffer, true, data)
		if err != nil {
			panic(err)
		}

		commandQueue.Flush()
		commandQueue.Finish()

		fmt.Println()
		printHeader("Output")
		for _, item := range data {
			fmt.Printf("%v ", item)
		}
		fmt.Println()

		time.Sleep(60 * time.Second)
	}
}
