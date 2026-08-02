import { Building2, ShieldAlert, User } from "lucide-react";
import type { ReactNode } from "react";
import type { Node as GraphNode } from "@/client/types.gen";
import { isHistorical } from "@/lib/graph";

interface Props {
  nodes: GraphNode[];
  onHoverNode?: (id: string | null) => void;
  onClickNode?: (node: GraphNode) => void;
}

const SECTION_BORDER: Record<GraphNode["type"], string> = {
  company: "#c8a050",
  officer: "#8fa8cc",
  psc: "#d88888",
};

export function GraphRoster({ nodes, onHoverNode, onClickNode }: Props) {
  if (nodes.length === 0) return null;

  const companies = nodes
    .filter((n) => n.type === "company")
    .sort((a, b) => a.label.localeCompare(b.label));
  const officers = nodes
    .filter((n) => n.type === "officer")
    .sort((a, b) => a.label.localeCompare(b.label));
  const pscs = nodes
    .filter((n) => n.type === "psc")
    .sort((a, b) => a.label.localeCompare(b.label));

  return (
    <div
      data-print-hide
      className="absolute bottom-5 right-5 z-20 select-none overflow-y-auto"
      style={{
        background: "linear-gradient(175deg,#ffffff 0%,#faf6f0 100%)",
        border: "1px solid #d4c4a8",
        borderRadius: 4,
        boxShadow: "2px 4px 16px rgba(40,20,0,0.22)",
        padding: "10px 14px",
        fontFamily: "Inter, system-ui, sans-serif",
        transform: "rotate(0.4deg)",
        width: 200,
        maxHeight: "55vh",
      }}
    >
      <p
        style={{
          fontSize: 10,
          fontWeight: 700,
          letterSpacing: "0.16em",
          textTransform: "uppercase",
          color: "#1a1a2e",
          opacity: 0.55,
          marginBottom: 9,
        }}
      >
        Board · {nodes.length}
      </p>

      <RosterSection
        label="Companies"
        icon={<Building2 size={10} strokeWidth={1.8} color="#8a6500" />}
        items={companies}
        onHoverNode={onHoverNode}
        onClickNode={onClickNode}
      />
      <RosterSection
        label="Officers"
        icon={<User size={10} strokeWidth={1.8} color="#1a4a8a" />}
        items={officers}
        onHoverNode={onHoverNode}
        onClickNode={onClickNode}
      />
      <RosterSection
        label="PSCs"
        icon={<ShieldAlert size={10} strokeWidth={1.8} color="#8a1a1a" />}
        items={pscs}
        onHoverNode={onHoverNode}
        onClickNode={onClickNode}
      />
    </div>
  );
}

function RosterSection({
  label,
  icon,
  items,
  onHoverNode,
  onClickNode,
}: {
  label: string;
  icon: ReactNode;
  items: GraphNode[];
  onHoverNode?: (id: string | null) => void;
  onClickNode?: (node: GraphNode) => void;
}) {
  if (items.length === 0) return null;
  const borderColor = SECTION_BORDER[items[0].type];
  return (
    <div style={{ marginBottom: 9 }}>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 5,
          marginBottom: 5,
        }}
      >
        {icon}
        <span
          style={{
            fontSize: 9,
            fontWeight: 700,
            letterSpacing: "0.1em",
            textTransform: "uppercase",
            color: "#1a1a2e",
            opacity: 0.5,
          }}
        >
          {label} ({items.length})
        </span>
      </div>
      <ul
        style={{
          listStyle: "none",
          padding: 0,
          margin: 0,
          display: "flex",
          flexDirection: "column",
          gap: 3,
        }}
      >
        {items.map((n) => {
          const historical = isHistorical(n);
          const statusTag = n.meta.resigned_on
            ? "resigned"
            : n.meta.ceased_on
              ? "ceased"
              : null;
          return (
            <li
              key={n.id}
              onMouseEnter={() => onHoverNode?.(n.id)}
              onMouseLeave={() => onHoverNode?.(null)}
              onClick={() => onClickNode?.(n)}
              onKeyDown={(e) => {
                if (e.key === "Enter" || e.key === " ") onClickNode?.(n);
              }}
              style={{
                fontSize: 11,
                color: historical ? "#78716c" : "#1a1a2e",
                lineHeight: 1.3,
                paddingLeft: 6,
                borderLeft: `2px solid ${historical ? "#c8b89a" : borderColor}`,
                cursor: onClickNode ? "pointer" : "default",
                transition: "opacity 0.1s ease",
                display: "flex",
                alignItems: "baseline",
                gap: 4,
              }}
            >
              <span>{n.label}</span>
              {statusTag && (
                <span
                  style={{
                    fontSize: 8.5,
                    fontWeight: 700,
                    letterSpacing: "0.08em",
                    textTransform: "uppercase",
                    color: "rgba(138,26,26,0.5)",
                    flexShrink: 0,
                  }}
                >
                  {statusTag}
                </span>
              )}
            </li>
          );
        })}
      </ul>
    </div>
  );
}
