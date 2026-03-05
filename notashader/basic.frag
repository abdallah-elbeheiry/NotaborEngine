#version 460 core

in vec2 vUV;
out vec4 FragColor;

uniform sampler2D mainTexture;

void main() {
    FragColor = texture(mainTexture, vUV);
}