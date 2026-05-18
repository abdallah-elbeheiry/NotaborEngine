// basic_shader.frag.hlsl

struct PS_INPUT {
    float4 color    : COLOR0;
    float2 uv       : TEXCOORD0;
    float2 localPos : TEXCOORD1;
};

Texture2D    texSampler : register(t0);
SamplerState samplerState : register(s0);

cbuffer MaterialUniforms : register(b1) {
    bool  uUseTexture;
    bool  uCircleMask;
    float uCircleRadius;
    float uCircleEdge;
};

float4 main(PS_INPUT input) : SV_TARGET {
    float4 color = input.color;
    
    // Apply texture if enabled
    if (uUseTexture) {
        float4 texColor = texSampler.Sample(samplerState, input.uv);
        color = color * texColor;
    }
    
    // Apply circular mask if enabled
    if (uCircleMask) {
        float dist = length(input.localPos);
        float mask = smoothstep(
            uCircleRadius + uCircleEdge,
            uCircleRadius - uCircleEdge,
            dist
        );
        color.a *= mask;
    }
    
    return color;
}