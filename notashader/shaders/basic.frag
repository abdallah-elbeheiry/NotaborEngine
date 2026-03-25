#version 460 core

in vec4 vColor;
in vec2 vUV;
in vec2 vLocalPos;

out vec4 FragColor;

uniform sampler2D uTexture;
uniform bool uUseTexture = false;
uniform bool uCircleMask = false;
uniform float uCircleRadius = 0.5;
uniform float uCircleEdge = 0.01;

void main() {
    vec4 color = vColor;

    if (uUseTexture) {
        color *= texture(uTexture, vUV);
    }

    if (uCircleMask) {
        float dist = length(vLocalPos);
        float alpha = 1.0 - smoothstep(uCircleRadius - uCircleEdge, uCircleRadius, dist);
        if (alpha <= 0.0)
        discard;
        color.a *= alpha;
    }

    FragColor = color;
}