#version 450

// Input from vertex shader
layout(location = 0) in VS_OUT {
    vec4 color;
    vec2 uv;
    vec2 localPos;
} fs_in;

// Textures and samplers
layout(set = 0, binding = 0) uniform sampler2D texSampler;

// Uniforms (using descriptor sets for Vulkan compatibility)
layout(set = 0, binding = 1) uniform MaterialUniforms {
    bool uUseTexture;
    bool uCircleMask;
    float uCircleRadius;
    float uCircleEdge;
} material;

// Output
layout(location = 0) out vec4 outColor;

void main()
{
    vec4 color = fs_in.color;
    
    // Apply texture if enabled
    if (material.uUseTexture) {
        vec4 texColor = texture(texSampler, fs_in.uv);
        color = color * texColor;
    }
    
    // Apply circular mask if enabled
    if (material.uCircleMask) {
        float dist = length(fs_in.localPos);
        float mask = smoothstep(
            material.uCircleRadius + material.uCircleEdge,
            material.uCircleRadius - material.uCircleEdge,
            dist
        );
        color.a *= mask;
    }
    
    outColor = color;
}
