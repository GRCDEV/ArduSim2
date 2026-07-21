export function getServiceColor(serviceId: string, isDark: boolean) {
  let hash = 0;
  for (let i = 0; i < serviceId.length; i++) {
    hash = serviceId.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  const lightness = isDark ? 30 : 45;
  const borderLightness = isDark ? 45 : 35;

  return {
    backgroundColor: `hsl(${hue}, 65%, ${lightness}%)`,
    color: "#ffffff",
    borderColor: `hsl(${hue}, 65%, ${borderLightness}%)`,
  };
}
