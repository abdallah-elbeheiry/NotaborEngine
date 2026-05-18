#version 450

// Vertex input attributes (must match Vulkan vertex input state)
layout(location = 0) in vec2 inPos;
layout(location = 1) in vec4 inColor;
layout(location = 2) in vec2 inUV;
layout(location = 3) in vec2 inLocalPos;

// Output to fragment shader
layout(location = 0) out VS_OUT {
    vec4 color;
    vec2 uv;
    vec2 localPos;
} vs_out;

// Push constants (equivalent to OpenGL uniforms)
layout(push_constant) uniform PushConstants {
    mat4 projection;
    mat4 view;
    mat4 model;
} pc;

void main()
{
    // Transform vertex position
    vec4 worldPos = pc.model * vec4(inPos, 0.0, 1.0);
    vec4 viewPos = pc.view * worldPos;
    gl_Position = pc.projection * viewPos;
    
    // Pass data to fragment shader
    vs_out.color = inColor;
    vs_out.uv = inUV;
    vs_out.localPos = inLocalPos;
}
