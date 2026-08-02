// Shared corkboard pushpin — a glossy head with a metal needle below it.
// Used on every Polaroid card (GraphCanvas), the selected-node panel
// (NodePanel), and the static intro/error cards (Home) so the "pinned paper"
// look is consistent everywhere, not just on graph nodes.
export function Pushpin({
  color,
  size = 16,
}: {
  color: string;
  size?: number;
}) {
  const height = size * (22 / 16);
  return (
    <svg
      aria-hidden="true"
      width={size}
      height={height}
      viewBox="0 0 16 22"
      style={{ display: "block" }}
    >
      <circle
        cx={8}
        cy={7.5}
        r={6.5}
        fill={color}
        stroke="rgba(0,0,0,0.22)"
        strokeWidth={0.8}
      />
      <circle cx={6.2} cy={5.4} r={2.4} fill="rgba(255,255,255,0.36)" />
      <line
        x1={8}
        y1={13.5}
        x2={8}
        y2={22}
        stroke="#666"
        strokeWidth={1.3}
        strokeLinecap="round"
      />
    </svg>
  );
}
