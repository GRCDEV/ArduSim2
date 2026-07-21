export const UAV_COLORS = [
  "#ef4444",
  "#3b82f6",
  "#10b981",
  "#f59e0b",
  "#8b5cf6",
  "#ec4899",
  "#14b8a6",
  "#f97316",
];

export function getUavColor(idx: number) {
  return UAV_COLORS[idx % UAV_COLORS.length];
}

export const RGB_UAV_COLORS = [
  [239, 68, 68], // #ef4444
  [59, 130, 246], // #3b82f6
  [16, 185, 129], // #10b981
  [245, 158, 11], // #f59e0b
  [139, 92, 246], // #8b5cf6
  [236, 72, 153], // #ec4899
  [20, 184, 166], // #14b8a6
  [249, 115, 22], // #f97316
];

export function getRgbUavColor(idx: number) {
  return RGB_UAV_COLORS[idx % RGB_UAV_COLORS.length];
}
