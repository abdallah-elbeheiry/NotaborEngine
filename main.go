package main

import (
	"fmt"
	"log"
	"runtime"

	"NotaborEngine/notavulkan"

	"github.com/bbredesen/go-vk"
)

func init() {
	runtime.LockOSThread() // required by GLFW
}

func main() {
	// 1. Create the renderer with a Vulkan instance
	renderer, err := notavulkan.NewRendererInstanceOnly("VulkanApp")
	if err != nil {
		log.Fatalln("failed to create Vulkan instance:", err)
	}

	// 2. Create window and surface (handled internally by notavulkan)
	width, height := 800, 600
	if err := renderer.CreateWindowAndSurface("Vulkan GLFW", width, height); err != nil {
		log.Fatalln("failed to create window and surface:", err)
	}

	// 3. Initialize device, swapchain, and other Vulkan resources
	if err := renderer.InitializeDeviceAndSwapchain(width, height, renderer.Surface); err != nil {
		log.Fatalln("failed to initialize device and swapchain:", err)
	}

	fmt.Printf("Vulkan initialized!\n")
	fmt.Printf("Instance: %v\nDevice: %v\nSwapchain images: %d\n",
		renderer.Instance, renderer.Device, len(renderer.Swapchain.Images))

	// 4. Record a test command buffer
	cmdBuf := renderer.CommandBuffers[0]
	beginInfo := vk.CommandBufferBeginInfo{}
	if err := vk.BeginCommandBuffer(cmdBuf, &beginInfo); err != nil {
		log.Fatalln("failed to begin command buffer:", err)
	}

	clearColor := vk.ClearValue{
		Color: vk.ClearColorValue{TypeFloat32: [4]float32{0.1, 0.2, 0.3, 1.0}},
	}

	rpBegin := vk.RenderPassBeginInfo{
		RenderPass:   renderer.RenderPass,
		Framebuffer:  renderer.Framebuffers[0],
		RenderArea:   vk.Rect2D{Extent: renderer.Swapchain.Extent},
		PClearValues: []vk.ClearValue{clearColor},
	}

	vk.CmdBeginRenderPass(cmdBuf, &rpBegin, vk.SUBPASS_CONTENTS_INLINE)
	vk.CmdEndRenderPass(cmdBuf)

	if err := vk.EndCommandBuffer(cmdBuf); err != nil {
		log.Fatalln("failed to end command buffer:", err)
	}

	fmt.Println("Command buffer recorded, ready for submission (not submitted).")

	// Keep the window open briefly (you'll replace this with a proper event loop)
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()

	// Cleanup
	renderer.DestroyWindow()
	// TODO: Add Vulkan resource cleanup (device, instance, etc.)
}
