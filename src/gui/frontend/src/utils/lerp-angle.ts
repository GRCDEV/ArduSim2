export function lerpAngle(a: number, b: number, t: number) {
  const diff = ((b - a + 180) % 360) - 180;
  return a + diff * t;
}
