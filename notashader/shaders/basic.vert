#version 460 core

layout(location = 0) in vec2 aPos;
layout(location = 1) in vec4 aColor;
layout(location = 2) in vec2 aUV;
layout(location = 3) in vec2 aLocalPos; // new attribute for local space

out vec4 vColor;
out vec2 vUV;
out vec2 vLocalPos;

void main() {
    gl_Position = vec4(aPos, 0.0, 1.0); // aPos is already transformed by model
    vColor = aColor;
    vUV = aUV;
    vLocalPos = aLocalPos; // keep local coordinates
}