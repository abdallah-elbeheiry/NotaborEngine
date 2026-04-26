@vertex
fn main(@location(0) inPos: vec2<f32>,
        @location(1) inColor: vec3<f32>) ->
        @builtin(position) vec4<f32> {
    return vec4<f32>(inPos, 0.0, 1.0);
}