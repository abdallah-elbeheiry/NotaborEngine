struct VS_INPUT {
    float2 pos      : POSITION0;
    float4 color    : COLOR0;
    float2 uv       : TEXCOORD0;
    float2 localPos : TEXCOORD1;
};

struct VS_OUTPUT {
    float4 color    : COLOR0;
    float2 uv       : TEXCOORD0;
    float2 localPos : TEXCOORD1;
    float4 position : SV_POSITION;
};

VS_OUTPUT main(VS_INPUT input) {
    VS_OUTPUT output;

    // Simple pass-through: position already in screen space or normalized coordinates
    // For now, assume input.pos is already in clip space
    output.position = float4(input.pos, 0.0, 1.0);

    output.color    = input.color;
    output.uv       = input.uv;
    output.localPos = input.localPos;

    return output;
}