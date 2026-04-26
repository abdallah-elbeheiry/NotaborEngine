@fragment
fn main(@location(1) fragColor: vec3<f32>) -> @location(0) vec4<f32> {
    return vec4<f32>(fragColor, 1.0);
}