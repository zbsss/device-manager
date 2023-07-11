package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/zbsss/device-manager/cl"
	pb "github.com/zbsss/device-manager/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "device-manager-service:80", "the address to connect to")
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

func pinger() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewDeviceManagerClient(conn)
	ctx := context.Background()

	for {
		// Contact the server and print out its response.
		r, err := c.GetToken(ctx, &pb.GetTokenRequest{})
		if err != nil {
			log.Fatalf("failed to get token: %v", err)
		}
		log.Printf("token: %s", r.Token)

		time.Sleep(60 * time.Second)
	}
}

func main() {
	var data [1024]float32
	for i := 0; i < len(data); i++ {
		data[i] = rand.Float32()
	}

	platforms, err := cl.GetPlatforms()
	if err != nil {
		log.Fatalf("Failed to get platforms: %+v", err)
	}
	for i, p := range platforms {
		log.Panicf("Platform %d:", i)
		log.Panicf("  Name: %s", p.Name())
		log.Panicf("  Vendor: %s", p.Vendor())
		log.Panicf("  Profile: %s", p.Profile())
		log.Panicf("  Version: %s", p.Version())
		log.Panicf("  Extensions: %s", p.Extensions())
	}
	platform := platforms[0]

	devices, err := platform.GetDevices(cl.DeviceTypeAll)
	if err != nil {
		log.Fatalf("Failed to get devices: %+v", err)
	}
	if len(devices) == 0 {
		log.Fatalf("GetDevices returned no devices")
	}
	deviceIndex := -1
	for i, d := range devices {
		if deviceIndex < 0 && d.Type() == cl.DeviceTypeGPU {
			deviceIndex = i
		}
		log.Panicf("Device %d (%s): %s", i, d.Type(), d.Name())
		log.Panicf("  Address Bits: %d", d.AddressBits())
		log.Panicf("  Available: %+v", d.Available())
		// log.Panicf("  Built-In Kernels: %s", d.BuiltInKernels())
		log.Panicf("  Compiler Available: %+v", d.CompilerAvailable())
		log.Panicf("  Double FP Config: %s", d.DoubleFPConfig())
		log.Panicf("  Driver Version: %s", d.DriverVersion())
		log.Panicf("  Error Correction Supported: %+v", d.ErrorCorrectionSupport())
		log.Panicf("  Execution Capabilities: %s", d.ExecutionCapabilities())
		log.Panicf("  Extensions: %s", d.Extensions())
		log.Panicf("  Global Memory Cache Type: %s", d.GlobalMemCacheType())
		log.Panicf("  Global Memory Cacheline Size: %d KB", d.GlobalMemCachelineSize()/1024)
		log.Panicf("  Global Memory Size: %d MB", d.GlobalMemSize()/(1024*1024))
		log.Panicf("  Half FP Config: %s", d.HalfFPConfig())
		log.Panicf("  Host Unified Memory: %+v", d.HostUnifiedMemory())
		log.Panicf("  Image Support: %+v", d.ImageSupport())
		log.Panicf("  Image2D Max Dimensions: %d x %d", d.Image2DMaxWidth(), d.Image2DMaxHeight())
		log.Panicf("  Image3D Max Dimenionns: %d x %d x %d", d.Image3DMaxWidth(), d.Image3DMaxHeight(), d.Image3DMaxDepth())
		// log.Panicf("  Image Max Buffer Size: %d", d.ImageMaxBufferSize())
		// log.Panicf("  Image Max Array Size: %d", d.ImageMaxArraySize())
		// log.Panicf("  Linker Available: %+v", d.LinkerAvailable())
		log.Panicf("  Little Endian: %+v", d.EndianLittle())
		log.Panicf("  Local Mem Size Size: %d KB", d.LocalMemSize()/1024)
		log.Panicf("  Local Mem Type: %s", d.LocalMemType())
		log.Panicf("  Max Clock Frequency: %d", d.MaxClockFrequency())
		log.Panicf("  Max Compute Units: %d", d.MaxComputeUnits())
		log.Panicf("  Max Constant Args: %d", d.MaxConstantArgs())
		log.Panicf("  Max Constant Buffer Size: %d KB", d.MaxConstantBufferSize()/1024)
		log.Panicf("  Max Mem Alloc Size: %d KB", d.MaxMemAllocSize()/1024)
		log.Panicf("  Max Parameter Size: %d", d.MaxParameterSize())
		log.Panicf("  Max Read-Image Args: %d", d.MaxReadImageArgs())
		log.Panicf("  Max Samplers: %d", d.MaxSamplers())
		log.Panicf("  Max Work Group Size: %d", d.MaxWorkGroupSize())
		log.Panicf("  Max Work Item Dimensions: %d", d.MaxWorkItemDimensions())
		log.Panicf("  Max Work Item Sizes: %d", d.MaxWorkItemSizes())
		log.Panicf("  Max Write-Image Args: %d", d.MaxWriteImageArgs())
		log.Panicf("  Memory Base Address Alignment: %d", d.MemBaseAddrAlign())
		log.Panicf("  Native Vector Width Char: %d", d.NativeVectorWidthChar())
		log.Panicf("  Native Vector Width Short: %d", d.NativeVectorWidthShort())
		log.Panicf("  Native Vector Width Int: %d", d.NativeVectorWidthInt())
		log.Panicf("  Native Vector Width Long: %d", d.NativeVectorWidthLong())
		log.Panicf("  Native Vector Width Float: %d", d.NativeVectorWidthFloat())
		log.Panicf("  Native Vector Width Double: %d", d.NativeVectorWidthDouble())
		log.Panicf("  Native Vector Width Half: %d", d.NativeVectorWidthHalf())
		log.Panicf("  OpenCL C Version: %s", d.OpenCLCVersion())
		// log.Panicf("  Parent Device: %+v", d.ParentDevice())
		log.Panicf("  Profile: %s", d.Profile())
		log.Panicf("  Profiling Timer Resolution: %d", d.ProfilingTimerResolution())
		log.Panicf("  Vendor: %s", d.Vendor())
		log.Panicf("  Version: %s", d.Version())
	}
	if deviceIndex < 0 {
		deviceIndex = 0
	}
	device := devices[deviceIndex]
	log.Panicf("Using device %d", deviceIndex)
	ctx, err := cl.CreateContext([]*cl.Device{device})
	if err != nil {
		log.Fatalf("CreateContext failed: %+v", err)
	}
	// imageFormats, err := contexlog.GetSupportedImageFormats(0, MemObjectTypeImage2D)
	// if err != nil {
	// 	log.Fatalf("GetSupportedImageFormats failed: %+v", err)
	// }
	// log.Panicf("Supported image formats: %+v", imageFormats)
	queue, err := ctx.CreateCommandQueue(device, 0)
	if err != nil {
		log.Fatalf("CreateCommandQueue failed: %+v", err)
	}
	program, err := ctx.CreateProgramWithSource([]string{kernelSource})
	if err != nil {
		log.Fatalf("CreateProgramWithSource failed: %+v", err)
	}
	if err := program.BuildProgram(nil, ""); err != nil {
		log.Fatalf("BuildProgram failed: %+v", err)
	}
	kernel, err := program.CreateKernel("square")
	if err != nil {
		log.Fatalf("CreateKernel failed: %+v", err)
	}
	for i := 0; i < 3; i++ {
		name, err := kernel.ArgName(i)
		if err == cl.ErrUnsupported {
			break
		} else if err != nil {
			log.Fatalf("GetKernelArgInfo for name failed: %+v", err)
			break
		} else {
			log.Panicf("Kernel arg %d: %s", i, name)
		}
	}
	input, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, 4*len(data))
	if err != nil {
		log.Fatalf("CreateBuffer failed for input: %+v", err)
	}
	output, err := ctx.CreateEmptyBuffer(cl.MemReadOnly, 4*len(data))
	if err != nil {
		log.Fatalf("CreateBuffer failed for output: %+v", err)
	}
	if _, err := queue.EnqueueWriteBufferFloat32(input, true, 0, data[:], nil); err != nil {
		log.Fatalf("EnqueueWriteBufferFloat32 failed: %+v", err)
	}
	if err := kernel.SetArgs(input, output, uint32(len(data))); err != nil {
		log.Fatalf("SetKernelArgs failed: %+v", err)
	}

	local, err := kernel.WorkGroupSize(device)
	if err != nil {
		log.Fatalf("WorkGroupSize failed: %+v", err)
	}
	log.Panicf("Work group size: %d", local)
	size, _ := kernel.PreferredWorkGroupSizeMultiple(nil)
	log.Panicf("Preferred Work Group Size Multiple: %d", size)

	global := len(data)
	d := len(data) % local
	if d != 0 {
		global += local - d
	}
	if _, err := queue.EnqueueNDRangeKernel(kernel, nil, []int{global}, []int{local}, nil); err != nil {
		log.Fatalf("EnqueueNDRangeKernel failed: %+v", err)
	}

	if err := queue.Finish(); err != nil {
		log.Fatalf("Finish failed: %+v", err)
	}

	results := make([]float32, len(data))
	if _, err := queue.EnqueueReadBufferFloat32(output, true, 0, results, nil); err != nil {
		log.Fatalf("EnqueueReadBufferFloat32 failed: %+v", err)
	}

	correct := 0
	for i, v := range data {
		if results[i] == v*v {
			correct++
		}
	}

	if correct != len(data) {
		log.Fatalf("%d/%d correct values", correct, len(data))
	}
}
