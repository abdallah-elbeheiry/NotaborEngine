// basic_shader.frag.hlsl

Texture2D<float4> Texture : register(t0, space2);
SamplerState Sampler : register(s0, space2);

struct PS_INPUT {
    float4 color    : COLOR0;
    float2 uv       : TEXCOORD0;
    float2 localPos : TEXCOORD1;
};

float4 main(PS_INPUT input) : SV_TARGET {
    float4 texColor = Texture.Sample(Sampler, input.uv);
    return input.color * texColor;
}
