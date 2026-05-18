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

cbuffer PushConstants : register(b0) {
    float4x4 projection;
    float4x4 view;
    float4x4 model;
};

VS_OUTPUT main(VS_INPUT input) {
VS_OUTPUT output;

float4 worldPos = mul(model, float4(input.pos, 0.0, 1.0));
float4 viewPos  = mul(view, worldPos);
output.position = mul(projection, viewPos);

output.color    = input.color;
output.uv       = input.uv;
output.localPos = input.localPos;

return output;
}