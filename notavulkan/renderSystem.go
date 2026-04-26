package notavulkan

import "C"
import (
	"fmt"
	"runtime"

	"github.com/bbredesen/go-vk"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Renderer holds all Vulkan state
type Renderer struct {
	Instance vk.Instance
	Device   vk.Device
	Physical vk.PhysicalDevice

	GraphicsQueue vk.Queue
	PresentQueue  vk.Queue

	Swapchain      *Swapchain
	CommandPool    vk.CommandPool
	CommandBuffers []vk.CommandBuffer

	RenderPass   vk.RenderPass
	Framebuffers []vk.Framebuffer

	Window  *glfw.Window
	Surface vk.SurfaceKHR
}

// Swapchain holds Vulkan swapchain data
type Swapchain struct {
	Handle vk.SwapchainKHR
	Images []vk.Image
	Views  []vk.ImageView
	Extent vk.Extent2D
	Format vk.Format
}

// NewRendererInstanceOnly creates a Vulkan instance only (no device yet)
func NewRendererInstanceOnly(appName string) (*Renderer, error) {
	app := vk.ApplicationInfo{
		PApplicationName: appName,
		ApiVersion:       vk.MAKE_VERSION(1, 3, 0),
	}

	instanceCI := vk.InstanceCreateInfo{
		PApplicationInfo: &app,
		PpEnabledExtensionNames: []string{
			vk.KHR_SURFACE_EXTENSION_NAME,
			"VK_KHR_win32_surface", // <-- required for Windows
		},
	}

	instance, err := vk.CreateInstance(&instanceCI, nil)
	if err != nil {
		return nil, err
	}
	return &Renderer{Instance: instance}, nil
}

// InitializeDeviceAndSwapchain sets up the device, queues, swapchain, command pool, render pass, framebuffers
func (r *Renderer) InitializeDeviceAndSwapchain(width, height int, surface vk.SurfaceKHR) error {
	// Pick first physical device
	devices, err := vk.EnumeratePhysicalDevices(r.Instance)
	if err != nil || len(devices) == 0 {
		return err
	}
	r.Physical = devices[0]

	// Find graphics queue family
	var graphicsFamily int
	queueFamilies := vk.GetPhysicalDeviceQueueFamilyProperties(r.Physical)
	for i, q := range queueFamilies {
		if q.QueueFlags&vk.QUEUE_GRAPHICS_BIT != 0 {
			graphicsFamily = i
			break
		}
	}

	// Create logical device
	device, err := vk.CreateDevice(r.Physical, &vk.DeviceCreateInfo{
		PQueueCreateInfos: []vk.DeviceQueueCreateInfo{
			{
				QueueFamilyIndex: uint32(graphicsFamily),
				PQueuePriorities: []float32{1.0},
			},
		},
		PpEnabledExtensionNames: []string{vk.KHR_SWAPCHAIN_EXTENSION_NAME},
	}, nil)
	if err != nil {
		return err
	}
	r.Device = device
	r.GraphicsQueue = vk.GetDeviceQueue(device, uint32(graphicsFamily), 0)
	r.PresentQueue = r.GraphicsQueue

	// Create swapchain
	r.Swapchain, err = r.CreateSwapchain(width, height, surface)
	if err != nil {
		return err
	}

	// Create command pool
	pool, err := vk.CreateCommandPool(r.Device, &vk.CommandPoolCreateInfo{
		QueueFamilyIndex: uint32(graphicsFamily),
		Flags:            vk.COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT,
	}, nil)
	if err != nil {
		return err
	}
	r.CommandPool = pool

	// Allocate command buffers (one per swapchain image)
	r.CommandBuffers = make([]vk.CommandBuffer, len(r.Swapchain.Images))
	allocInfo := vk.CommandBufferAllocateInfo{
		CommandPool:        r.CommandPool,
		Level:              vk.COMMAND_BUFFER_LEVEL_PRIMARY,
		CommandBufferCount: uint32(len(r.CommandBuffers)),
	}
	r.CommandBuffers, err = vk.AllocateCommandBuffers(r.Device, &allocInfo)
	if err != nil {
		return err
	}

	// Create minimal render pass
	r.RenderPass, err = r.CreateRenderPass()
	if err != nil {
		return err
	}

	// Create framebuffers
	r.Framebuffers, err = r.CreateFramebuffers()
	if err != nil {
		return err
	}

	return nil
}

// CreateSwapchain creates swapchain and image views
func (r *Renderer) CreateSwapchain(width, height int, surface vk.SurfaceKHR) (*Swapchain, error) {
	caps, err := vk.GetPhysicalDeviceSurfaceCapabilitiesKHR(r.Physical, surface)
	if err != nil {
		return nil, err
	}

	formats, err := vk.GetPhysicalDeviceSurfaceFormatsKHR(r.Physical, surface)
	if err != nil {
		return nil, err
	}
	format := formats[0].Format
	colorSpace := formats[0].ColorSpace

	extent := vk.Extent2D{Width: uint32(width), Height: uint32(height)}
	count := caps.MinImageCount + 1
	if caps.MaxImageCount > 0 && count > caps.MaxImageCount {
		count = caps.MaxImageCount
	}

	sc, err := vk.CreateSwapchainKHR(r.Device, &vk.SwapchainCreateInfoKHR{
		Surface:         surface,
		MinImageCount:   count,
		ImageFormat:     format,
		ImageColorSpace: colorSpace,
		ImageExtent:     extent,
		ImageUsage:      vk.IMAGE_USAGE_COLOR_ATTACHMENT_BIT,
	}, nil)
	if err != nil {
		return nil, err
	}

	images, err := vk.GetSwapchainImagesKHR(r.Device, sc)
	if err != nil {
		return nil, err
	}

	views := make([]vk.ImageView, len(images))
	for i, img := range images {
		view, err := vk.CreateImageView(r.Device, &vk.ImageViewCreateInfo{
			Image:    img,
			ViewType: vk.IMAGE_VIEW_TYPE_2D,
			Format:   format,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.IMAGE_ASPECT_COLOR_BIT,
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		}, nil)
		if err != nil {
			return nil, err
		}
		views[i] = view
	}

	return &Swapchain{
		Handle: sc,
		Images: images,
		Views:  views,
		Extent: extent,
		Format: format,
	}, nil
}

// CreateRenderPass creates a minimal render pass
func (r *Renderer) CreateRenderPass() (vk.RenderPass, error) {
	attachment := vk.AttachmentDescription{
		Format:         r.Swapchain.Format,
		Samples:        vk.SAMPLE_COUNT_1_BIT,
		LoadOp:         vk.ATTACHMENT_LOAD_OP_CLEAR,
		StoreOp:        vk.ATTACHMENT_STORE_OP_STORE,
		StencilLoadOp:  vk.ATTACHMENT_LOAD_OP_DONT_CARE,
		StencilStoreOp: vk.ATTACHMENT_STORE_OP_DONT_CARE,
		InitialLayout:  vk.IMAGE_LAYOUT_UNDEFINED,
		FinalLayout:    vk.IMAGE_LAYOUT_PRESENT_SRC_KHR,
	}

	colorAttachmentRef := vk.AttachmentReference{
		Attachment: 0,
		Layout:     vk.IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL,
	}

	subpass := vk.SubpassDescription{
		PipelineBindPoint: vk.PIPELINE_BIND_POINT_GRAPHICS,
		PColorAttachments: []vk.AttachmentReference{colorAttachmentRef},
	}

	return vk.CreateRenderPass(r.Device, &vk.RenderPassCreateInfo{
		PAttachments: []vk.AttachmentDescription{attachment},
		PSubpasses:   []vk.SubpassDescription{subpass},
	}, nil)
}

// CreateFramebuffers creates one framebuffer per swapchain image
func (r *Renderer) CreateFramebuffers() ([]vk.Framebuffer, error) {
	fbs := make([]vk.Framebuffer, len(r.Swapchain.Views))
	for i, view := range r.Swapchain.Views {
		fb, err := vk.CreateFramebuffer(r.Device, &vk.FramebufferCreateInfo{
			RenderPass:   r.RenderPass,
			PAttachments: []vk.ImageView{view},
			Width:        r.Swapchain.Extent.Width,
			Height:       r.Swapchain.Extent.Height,
			Layers:       1,
		}, nil)
		if err != nil {
			return nil, err
		}
		fbs[i] = fb
	}
	return fbs, nil
}

func (r *Renderer) CreateWindowAndSurface(title string, width, height int) error {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		return fmt.Errorf("failed to initialize GLFW: %w", err)
	}

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	glfw.WindowHint(glfw.Resizable, glfw.False)

	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create GLFW window: %w", err)
	}
	r.Window = window

	// Allocate Vulkan instance handle in C memory
	instanceC := r.Instance
	surfacePtr, err := window.CreateWindowSurface(&instanceC, nil)
	if err != nil {
		return fmt.Errorf("failed to create Vulkan surface: %w", err)
	}
	r.Surface = vk.SurfaceKHR(surfacePtr)

	return nil
}

// DestroyWindow cleans up GLFW resources.
func (r *Renderer) DestroyWindow() {
	if r.Window != nil {
		r.Window.Destroy()
		glfw.Terminate()
		r.Window = nil
	}
}
