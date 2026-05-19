// basic_shader.frag.hlsl

Texture2D<float4> Texture : register(t0, space2);
SamplerState Sampler : register(s0, space2);

struct PS_INPUT {
    float4 color    : COLOR0;
    float2 uv       : TEXCOORD0;
    float2 localPos : TEXCOORD1;
    float2 circleMask : TEXCOORD2;
    float useCircle : TEXCOORD3;
};

float4 main(PS_INPUT input) : SV_TARGET {
    float4 texColor = Texture.Sample(Sampler, input.uv);
    float4 color = input.color * texColor;

    if (input.useCircle > 0.5) {
        float radius = input.circleMask.x;
        float edge = max(input.circleMask.y, 0.0001);
        float dist = length(input.localPos);
        float alpha = 1.0 - smoothstep(radius - edge, radius, dist);
        clip(alpha - 0.001);
        color.a *= alpha;
    }

    return color;
}
