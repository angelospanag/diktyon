"use client";

import type { SimulationLinkDatum, SimulationNodeDatum } from "d3-force";
import {
  forceCenter,
  forceCollide,
  forceLink,
  forceManyBody,
  forceSimulation,
} from "d3-force";
import { Building2, ShieldAlert, User } from "lucide-react";
import type React from "react";
import {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from "react";
import type {
  Edge as GraphEdge,
  Node as GraphNode,
  Response as GraphResponse,
} from "@/client/types.gen";
import { Pushpin } from "@/components/Pushpin";
import {
  distressFlags,
  historicalLabel,
  isHistorical,
  mergeMeta,
  PIN_COLOR,
  tiltAngle,
} from "@/lib/graph";

// ── Types ──────────────────────────────────────────────────────────────────────

interface SimNode extends SimulationNodeDatum {
  id: string;
  graphNode: GraphNode;
  x: number;
  y: number;
}

interface D3Link extends SimulationLinkDatum<SimNode> {
  kind: string;
}

// ── Layout ─────────────────────────────────────────────────────────────────────

function runLayout(nodes: SimNode[], edges: GraphEdge[]): SimNode[] {
  const links: D3Link[] = edges.map((e) => ({
    source: e.source,
    target: e.target,
    kind: e.kind,
  }));

  forceSimulation<SimNode>(nodes)
    .force(
      "link",
      forceLink<SimNode, D3Link>(links)
        .id((d) => d.id)
        .distance((d: D3Link) => {
          const s = d.source as SimNode;
          const t = d.target as SimNode;
          return s.graphNode?.type === "company" ||
            t.graphNode?.type === "company"
            ? 160
            : 240;
        })
        .strength(0.9),
    )
    .force(
      "charge",
      forceManyBody<SimNode>().strength((d: SimNode) =>
        d.graphNode.type === "company" ? -1800 : -400,
      ),
    )
    .force("center", forceCenter(0, 0))
    .force("collide", forceCollide<SimNode>(100).iterations(2))
    .stop()
    .tick(300);

  for (const n of nodes) {
    delete n.fx;
    delete n.fy;
  }
  return nodes;
}

function layoutSubset(
  allNodes: SimNode[],
  allEdges: GraphEdge[],
  showHist: boolean,
) {
  if (showHist) return { nodes: allNodes, edges: allEdges };
  const nodes = allNodes.filter((n) => !isHistorical(n.graphNode));
  const ids = new Set(nodes.map((n) => n.id));
  return {
    nodes,
    edges: allEdges.filter((e) => ids.has(e.source) && ids.has(e.target)),
  };
}

// ── Card colours ───────────────────────────────────────────────────────────────

const NODE_BG: Record<GraphNode["type"], string> = {
  company: "linear-gradient(155deg,#ffffff 0%,#f5ede0 100%)",
  officer: "linear-gradient(135deg,#f4f7ff 0%,#dce8f5 100%)",
  psc: "linear-gradient(155deg,#fff8f8 0%,#fde8e8 100%)",
};
const NODE_BORDER: Record<GraphNode["type"], string> = {
  company: "#c8b898",
  officer: "#8fa8cc",
  psc: "#d88888",
};
const BADGE_COLOR: Record<GraphNode["type"], string> = {
  company: "#1a1a2e",
  officer: "#2a4a7f",
  psc: "#8a1a1a",
};
const NODE_ICON: Record<GraphNode["type"], React.ReactNode> = {
  company: <Building2 size={14} strokeWidth={1.8} />,
  officer: <User size={14} strokeWidth={1.8} />,
  psc: <ShieldAlert size={14} strokeWidth={1.8} />,
};

// ── Polaroid card ──────────────────────────────────────────────────────────────

function PolaroidCard({
  node,
  selected,
  dimmed,
  onClick,
  onExpand,
  onDragStart,
  onHover,
}: {
  node: SimNode;
  selected: boolean;
  dimmed: boolean;
  onClick: (n: GraphNode) => void;
  onExpand: (n: GraphNode) => void;
  onDragStart: (id: string, e: React.MouseEvent) => void;
  onHover: (id: string | null) => void;
}) {
  const gn = node.graphNode;
  const tilt = tiltAngle(gn.id);
  const historical = isHistorical(gn);
  const flags = distressFlags(gn);

  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: draggable graph node
    // biome-ignore lint/a11y/useKeyWithClickEvents: drag interaction, not a simple click target
    <div
      onMouseDown={(e) => {
        e.stopPropagation();
        onDragStart(node.id, e);
      }}
      onClick={(e) => {
        e.stopPropagation();
        onClick(gn);
      }}
      onMouseEnter={() => onHover(node.id)}
      onMouseLeave={() => onHover(null)}
      style={{
        position: "absolute",
        left: node.x,
        top: node.y,
        transform: `translate(-50%, -50%) rotate(${tilt}deg)`,
        cursor: "grab",
        userSelect: "none",
        opacity: dimmed ? 0.15 : historical ? 0.72 : 1,
        transition: "opacity 0.18s ease",
      }}
    >
      {/* Pushpin */}
      <div
        style={{
          position: "absolute",
          top: -16,
          left: "50%",
          transform: "translateX(-50%)",
          zIndex: 1,
        }}
      >
        <Pushpin color={PIN_COLOR[gn.type]} />
      </div>

      {/* Distress-flag badge — overdue filings / insolvency history / charges */}
      {flags.length > 0 && (
        <div
          title={flags.join(", ")}
          style={{
            position: "absolute",
            top: -7,
            left: -7,
            zIndex: 2,
            width: 18,
            height: 18,
            borderRadius: "50%",
            background: "#8a1a1a",
            color: "#fff",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontSize: 10,
            fontWeight: 700,
            border: "1.5px solid #faf6f0",
            boxShadow: "0 1px 4px rgba(0,0,0,0.3)",
          }}
        >
          ⚠
        </div>
      )}

      {/* Card body — animated on first mount via globals.css .polaroid-card */}
      <div
        className="polaroid-card"
        style={{
          position: "relative",
          width: 160,
          boxSizing: "border-box",
          padding: "9px 12px 10px",
          background: NODE_BG[gn.type],
          border: `${selected ? 2.5 : 1.5}px ${historical || gn.type === "psc" ? "dashed" : "solid"} ${selected ? "#cc2200" : historical ? "#9c8a72" : NODE_BORDER[gn.type]}`,
          borderRadius: 5,
          boxShadow: selected
            ? "3px 5px 22px rgba(204,34,0,0.38)"
            : "3px 5px 14px rgba(58,32,0,0.26)",
          fontFamily: "Inter,system-ui,sans-serif",
          filter: historical ? "grayscale(0.4)" : undefined,
        }}
      >
        {/* Case-file stamp for resigned officers / ceased PSCs */}
        {historical && (
          <div
            title={historicalLabel(gn) ?? undefined}
            style={{
              position: "absolute",
              top: 6,
              right: 6,
              transform: "rotate(-10deg)",
              border: "1.5px solid rgba(138,26,26,0.55)",
              borderRadius: 3,
              padding: "1px 4px",
              fontFamily: "JetBrains Mono, monospace",
              fontSize: 8,
              fontWeight: 700,
              letterSpacing: "0.06em",
              textTransform: "uppercase",
              color: "rgba(138,26,26,0.6)",
              pointerEvents: "none",
            }}
          >
            {gn.meta.resigned_on ? "Resigned" : "Ceased"}
          </div>
        )}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 4,
            fontSize: 10,
            fontWeight: 700,
            letterSpacing: "0.12em",
            textTransform: "uppercase",
            color: BADGE_COLOR[gn.type],
            opacity: 0.75,
            marginBottom: 4,
          }}
        >
          {NODE_ICON[gn.type]}
          {gn.type === "officer" ? "officer" : gn.type}
        </div>

        <div
          style={{
            fontSize: 14,
            fontWeight: 600,
            color: "#1a1a2e",
            lineHeight: 1.3,
            wordBreak: "break-word",
          }}
        >
          {gn.label}
        </div>

        {gn.type === "company" && gn.meta.status && (
          <div
            style={{
              marginTop: 5,
              fontSize: 10,
              fontWeight: 600,
              letterSpacing: "0.08em",
              textTransform: "uppercase",
              color: gn.meta.status === "active" ? "#2a6e2a" : "#7a4a1a",
            }}
          >
            {gn.meta.status}
          </div>
        )}
        {gn.type === "officer" && gn.meta.role && (
          <div style={{ marginTop: 4, fontSize: 10, color: "#445577" }}>
            {gn.meta.role}
          </div>
        )}
        {gn.type === "psc" && gn.meta.natures_of_control?.[0] && (
          <div style={{ marginTop: 4, fontSize: 10, color: "#7a2a2a" }}>
            {gn.meta.natures_of_control[0].replaceAll("-", " ")}
          </div>
        )}

        {/* Expand affordance */}
        {!gn.meta.expanded &&
          ((gn.type === "company" && !!gn.meta.company_number) ||
            (gn.type === "officer" && !!gn.meta.officer_id)) && (
            <button
              type="button"
              onMouseDown={(e) => e.stopPropagation()}
              onClick={(e) => {
                e.stopPropagation();
                onExpand(gn);
              }}
              title="Expand network"
              style={{
                position: "absolute",
                bottom: 6,
                right: 7,
                width: 16,
                height: 16,
                borderRadius: "50%",
                background: "#cc2200",
                color: "#fff",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontSize: 13,
                fontWeight: 700,
                lineHeight: 1,
                boxShadow: "0 1px 3px rgba(0,0,0,0.25)",
                border: "none",
                cursor: "pointer",
                padding: 0,
              }}
            >
              +
            </button>
          )}
      </div>
    </div>
  );
}

// ── SVG string edge ────────────────────────────────────────────────────────────

function StringEdge({
  src,
  tgt,
  kind,
  role,
  edgeId,
  dimmed,
}: {
  src: SimNode;
  tgt: SimNode;
  kind: string;
  role?: string;
  edgeId: string;
  dimmed: boolean;
}) {
  // Always route left-to-right so the textPath label reads correctly
  const [a, b] =
    src.x <= tgt.x
      ? [
          { x: src.x, y: src.y },
          { x: tgt.x, y: tgt.y },
        ]
      : [
          { x: tgt.x, y: tgt.y },
          { x: src.x, y: src.y },
        ];

  const sag = Math.hypot(b.x - a.x, b.y - a.y) * 0.07;
  const mx = (a.x + b.x) / 2;
  const my = (a.y + b.y) / 2 - sag;
  const d = `M ${a.x} ${a.y} Q ${mx} ${my} ${b.x} ${b.y}`;
  // Highlight path sits 1px above the body to simulate the round cross-section of string
  const dHi = `M ${a.x} ${a.y} Q ${mx} ${my - 1} ${b.x} ${b.y}`;
  const label =
    kind === "officer_of" ? (role?.toLowerCase() ?? "officer") : "controls";
  const dashes = kind === "psc_of" ? "8 5" : undefined;

  return (
    <g style={{ opacity: dimmed ? 0.08 : 1, transition: "opacity 0.18s ease" }}>
      {/* defs path is unfiltered so textPath follows the true bezier */}
      <defs>
        <path id={edgeId} d={d} />
      </defs>
      {/* String body — dark crimson, slightly thick */}
      <path
        d={d}
        fill="none"
        stroke="#8b1500"
        strokeWidth={3.5}
        strokeOpacity={0.88}
        strokeLinecap="round"
        strokeDasharray={dashes}
        filter="url(#string-texture)"
      />
      {/* Highlight — lighter stripe along the top of the string */}
      <path
        d={dHi}
        fill="none"
        stroke="#e05030"
        strokeWidth={0.9}
        strokeOpacity={0.5}
        strokeLinecap="round"
        strokeDasharray={dashes}
        filter="url(#string-texture)"
      />
      <text
        fontSize={10}
        fill="#8b1500"
        fillOpacity={0.9}
        stroke="#faf6f0"
        strokeWidth={2.5}
        strokeOpacity={0.85}
        style={{
          fontFamily: "JetBrains Mono, monospace",
          letterSpacing: "0.08em",
          paintOrder: "stroke fill",
        }}
      >
        <textPath href={`#${edgeId}`} startOffset="50%" textAnchor="middle">
          {label}
        </textPath>
      </text>
    </g>
  );
}

// ── GraphCanvas ────────────────────────────────────────────────────────────────

export interface GraphCanvasHandle {
  addGraph: (graph: GraphResponse) => void;
  setGraph: (graph: GraphResponse) => void;
  fit: () => void;
  setHoveredNode: (id: string | null) => void;
  markExpanded: (id: string) => void;
}

interface Props {
  onNodeClick: (node: GraphNode) => void;
  onExpand: (node: GraphNode) => void;
  selectedNodeId?: string | null;
  showResigned?: boolean;
}

export const GraphCanvas = forwardRef<GraphCanvasHandle, Props>(
  function GraphCanvas(
    { onNodeClick, onExpand, selectedNodeId, showResigned = false },
    ref,
  ) {
    const [nodes, setNodes] = useState<SimNode[]>([]);
    const [edges, setEdges] = useState<GraphEdge[]>([]);
    const [hoveredNodeId, setHoveredNodeId] = useState<string | null>(null);

    const nodesRef = useRef<SimNode[]>([]);
    const edgesRef = useRef<GraphEdge[]>([]);
    const containerRef = useRef<HTMLDivElement>(null);

    const panRef = useRef({ x: 0, y: 0 });
    const zoomRef = useRef(1);
    const [pan, setPan] = useState({ x: 0, y: 0 });
    const [zoom, setZoom] = useState(1);

    const showResignedRef = useRef(showResigned);
    useEffect(() => {
      showResignedRef.current = showResigned;
    }, [showResigned]);

    const dragging = useRef(false);
    const dragOrigin = useRef({ x: 0, y: 0 });

    const draggingNodeId = useRef<string | null>(null);
    const dragNodeOrigin = useRef({ x: 0, y: 0 });
    const dragMouseOrigin = useRef({ x: 0, y: 0 });

    // ── Fit ──────────────────────────────────────────────────────────────────

    const fitToNodes = useCallback((n: SimNode[]) => {
      const el = containerRef.current;
      if (!el || n.length === 0) return;
      const { width, height } = el.getBoundingClientRect();

      let minX = Infinity,
        maxX = -Infinity,
        minY = Infinity,
        maxY = -Infinity;
      for (const node of n) {
        minX = Math.min(minX, node.x - 90);
        maxX = Math.max(maxX, node.x + 90);
        minY = Math.min(minY, node.y - 55);
        maxY = Math.max(maxY, node.y + 55);
      }

      const pad = 60;
      const scale = Math.min(
        (width - pad * 2) / (maxX - minX),
        (height - pad * 2) / (maxY - minY),
        1.4,
      );
      const cx = (minX + maxX) / 2;
      const cy = (minY + maxY) / 2;

      panRef.current = { x: -cx * scale, y: -cy * scale };
      zoomRef.current = scale;
      setPan({ ...panRef.current });
      setZoom(scale);
    }, []);

    useEffect(() => {
      if (nodesRef.current.length === 0) return;
      const { nodes: ln, edges: le } = layoutSubset(
        nodesRef.current,
        edgesRef.current,
        showResigned,
      );
      if (ln.length === 0) return;
      runLayout(ln, le);
      setNodes([...nodesRef.current]);
      requestAnimationFrame(() => fitToNodes(ln));
    }, [showResigned, fitToNodes]);

    // ── Graph helpers ─────────────────────────────────────────────────────────

    function applyGraph(n: SimNode[], e: GraphEdge[]) {
      nodesRef.current = n;
      edgesRef.current = e;
      setNodes([...n]);
      setEdges([...e]);
    }

    useImperativeHandle(ref, () => ({
      setGraph(graph) {
        const allSimNodes: SimNode[] = (graph.nodes ?? []).map((n) => ({
          id: n.id,
          graphNode: n,
          x: (Math.random() - 0.5) * 300,
          y: (Math.random() - 0.5) * 300,
        }));
        const allEdges = graph.edges ?? [];
        const { nodes: ln, edges: le } = layoutSubset(
          allSimNodes,
          allEdges,
          showResignedRef.current,
        );
        runLayout(ln, le);
        applyGraph(allSimNodes, allEdges);
        requestAnimationFrame(() => fitToNodes(ln));
      },

      addGraph(graph) {
        const existingMap = new Map(nodesRef.current.map((n) => [n.id, n]));
        const edgeKey = (e: GraphEdge) => `${e.source}→${e.target}:${e.kind}`;

        const newNodes: SimNode[] = [];
        for (const n of graph.nodes ?? []) {
          const existing = existingMap.get(n.id);
          if (existing) {
            existing.graphNode = {
              ...n,
              meta: mergeMeta(existing.graphNode.meta, n.meta),
            };
          } else {
            const anchor = nodesRef.current[0] ?? { x: 0, y: 0 };
            newNodes.push({
              id: n.id,
              graphNode: n,
              x: anchor.x + (Math.random() - 0.5) * 200,
              y: anchor.y + (Math.random() - 0.5) * 200,
            });
          }
        }

        const newEdges = (graph.edges ?? []).filter(
          (e) => !edgesRef.current.some((ex) => edgeKey(ex) === edgeKey(e)),
        );
        if (newNodes.length === 0 && newEdges.length === 0) return;

        const pinned = nodesRef.current.map((n) => ({
          ...n,
          fx: n.x,
          fy: n.y,
        }));
        const allEdges = [...edgesRef.current, ...newEdges];
        const laid = runLayout([...pinned, ...newNodes], allEdges);

        applyGraph(laid, allEdges);
        requestAnimationFrame(() => {
          const { nodes: visible } = layoutSubset(
            nodesRef.current,
            [],
            showResignedRef.current,
          );
          fitToNodes(visible);
        });
      },

      fit() {
        fitToNodes(nodesRef.current);
      },

      setHoveredNode(id) {
        setHoveredNodeId(id);
      },

      markExpanded(id) {
        nodesRef.current = nodesRef.current.map((n) =>
          n.id === id
            ? {
                ...n,
                graphNode: {
                  ...n.graphNode,
                  meta: { ...n.graphNode.meta, expanded: true },
                },
              }
            : n,
        );
        setNodes([...nodesRef.current]);
      },
    }));

    // ── Helpers ───────────────────────────────────────────────────────────────

    const toSceneCoords = useCallback((clientX: number, clientY: number) => {
      const rect = containerRef.current?.getBoundingClientRect();
      if (!rect) return { x: 0, y: 0 };
      return {
        x:
          (clientX - rect.left - rect.width / 2 - panRef.current.x) /
          zoomRef.current,
        y:
          (clientY - rect.top - rect.height / 2 - panRef.current.y) /
          zoomRef.current,
      };
    }, []);

    // ── Canvas pan ────────────────────────────────────────────────────────────

    const onMouseDown = useCallback((e: React.MouseEvent) => {
      if (e.button !== 0) return;
      dragging.current = true;
      dragOrigin.current = {
        x: e.clientX - panRef.current.x,
        y: e.clientY - panRef.current.y,
      };
    }, []);

    const onMouseMove = useCallback(
      (e: React.MouseEvent) => {
        if (draggingNodeId.current) {
          const mouse = toSceneCoords(e.clientX, e.clientY);
          const dx = mouse.x - dragMouseOrigin.current.x;
          const dy = mouse.y - dragMouseOrigin.current.y;
          const newX = dragNodeOrigin.current.x + dx;
          const newY = dragNodeOrigin.current.y + dy;
          const id = draggingNodeId.current;

          const refNode = nodesRef.current.find((n) => n.id === id);
          if (refNode) {
            refNode.x = newX;
            refNode.y = newY;
          }
          setNodes((prev) =>
            prev.map((n) => (n.id === id ? { ...n, x: newX, y: newY } : n)),
          );
        } else if (dragging.current) {
          panRef.current = {
            x: e.clientX - dragOrigin.current.x,
            y: e.clientY - dragOrigin.current.y,
          };
          setPan({ ...panRef.current });
        }
      },
      [toSceneCoords],
    );

    const onMouseUp = useCallback(() => {
      dragging.current = false;
      draggingNodeId.current = null;
    }, []);

    // ── Node drag start ───────────────────────────────────────────────────────

    const handleNodeDragStart = useCallback(
      (id: string, e: React.MouseEvent) => {
        if (e.button !== 0) return;
        const node = nodesRef.current.find((n) => n.id === id);
        if (!node) return;
        draggingNodeId.current = id;
        dragNodeOrigin.current = { x: node.x, y: node.y };
        dragMouseOrigin.current = toSceneCoords(e.clientX, e.clientY);
        setHoveredNodeId(null);
      },
      [toSceneCoords],
    );

    // ── Zoom (wheel) ──────────────────────────────────────────────────────────

    useEffect(() => {
      const el = containerRef.current;
      if (!el) return;

      function onWheel(e: WheelEvent) {
        e.preventDefault();
        if (!el) return;
        const rect = el.getBoundingClientRect();
        const mouseX = e.clientX - rect.left - rect.width / 2;
        const mouseY = e.clientY - rect.top - rect.height / 2;
        const factor = e.deltaY > 0 ? 0.9 : 1.1;
        const newZoom = Math.min(Math.max(zoomRef.current * factor, 0.08), 5);
        const ratio = newZoom / zoomRef.current;

        panRef.current = {
          x: mouseX + (panRef.current.x - mouseX) * ratio,
          y: mouseY + (panRef.current.y - mouseY) * ratio,
        };
        zoomRef.current = newZoom;
        setPan({ ...panRef.current });
        setZoom(newZoom);
      }

      el.addEventListener("wheel", onWheel, { passive: false });
      return () => el.removeEventListener("wheel", onWheel);
    }, []);

    // ── Print: fit before, restore after ─────────────────────────────────────

    useEffect(() => {
      let saved: { pan: { x: number; y: number }; zoom: number } | null = null;

      function onBefore() {
        saved = { pan: { ...panRef.current }, zoom: zoomRef.current };
        fitToNodes(
          nodesRef.current.filter(
            (n) => showResigned || !isHistorical(n.graphNode),
          ),
        );
      }

      function onAfter() {
        if (!saved) return;
        panRef.current = saved.pan;
        zoomRef.current = saved.zoom;
        setPan({ ...saved.pan });
        setZoom(saved.zoom);
        saved = null;
      }

      window.addEventListener("beforeprint", onBefore);
      window.addEventListener("afterprint", onAfter);
      return () => {
        window.removeEventListener("beforeprint", onBefore);
        window.removeEventListener("afterprint", onAfter);
      };
    }, [showResigned, fitToNodes]);

    // ── Render ────────────────────────────────────────────────────────────────

    // Resigned officers / ceased PSCs stay in the simulation (so toggling
    // doesn't jitter the layout of visible nodes) but are hidden from
    // rendering, the roster, and hover/dim logic unless the toggle is on.
    const visibleNodes = showResigned
      ? nodes
      : nodes.filter((n) => !isHistorical(n.graphNode));
    const visibleIds = new Set(visibleNodes.map((n) => n.id));
    const visibleEdges = showResigned
      ? edges
      : edges.filter(
          (e) => visibleIds.has(e.source) && visibleIds.has(e.target),
        );

    const nodeMap = new Map(visibleNodes.map((n) => [n.id, n]));

    // Nodes directly connected to the hovered node — computed once per render.
    const neighborIds: Set<string> | null = hoveredNodeId
      ? new Set(
          visibleEdges.flatMap((e) =>
            e.source === hoveredNodeId
              ? [e.target]
              : e.target === hoveredNodeId
                ? [e.source]
                : [],
          ),
        )
      : null;

    const isNodeDimmed = (id: string) =>
      neighborIds !== null && id !== hoveredNodeId && !neighborIds.has(id);

    const isEdgeDimmed = (e: GraphEdge) =>
      neighborIds !== null &&
      e.source !== hoveredNodeId &&
      e.target !== hoveredNodeId;

    return (
      <div
        ref={containerRef}
        role="application"
        aria-label="Corporate network graph"
        className="cork-board"
        style={{
          width: "100%",
          height: "100%",
          overflow: "hidden",
          position: "relative",
          cursor: dragging.current ? "grabbing" : "grab",
        }}
        onMouseDown={onMouseDown}
        onMouseMove={onMouseMove}
        onMouseUp={onMouseUp}
        onMouseLeave={() => {
          onMouseUp();
          setHoveredNodeId(null);
        }}
      >
        <div
          style={{
            position: "absolute",
            left: "50%",
            top: "50%",
            transform: `translate(${pan.x}px,${pan.y}px) scale(${zoom})`,
            transformOrigin: "0 0",
            willChange: "transform",
          }}
        >
          {/* SVG string layer */}
          <svg
            aria-hidden="true"
            width={1}
            height={1}
            overflow="visible"
            style={{
              position: "absolute",
              left: 0,
              top: 0,
              pointerEvents: "none",
            }}
          >
            <defs>
              <filter
                id="string-texture"
                x="-5%"
                y="-5%"
                width="110%"
                height="110%"
              >
                <feTurbulence
                  type="fractalNoise"
                  baseFrequency="0.035"
                  numOctaves="2"
                  seed="4"
                  result="noise"
                />
                <feDisplacementMap
                  in="SourceGraphic"
                  in2="noise"
                  scale="1.8"
                  xChannelSelector="R"
                  yChannelSelector="G"
                />
              </filter>
            </defs>
            {visibleEdges.map((e) => {
              const src = nodeMap.get(e.source);
              const tgt = nodeMap.get(e.target);
              if (!src || !tgt) return null;
              const key = `${e.source}→${e.target}:${e.kind}`;
              const edgeId = key.replace(/[^a-zA-Z0-9]/g, "_");
              return (
                <StringEdge
                  key={key}
                  edgeId={edgeId}
                  src={src}
                  tgt={tgt}
                  kind={e.kind}
                  role={src.graphNode.meta?.role}
                  dimmed={isEdgeDimmed(e)}
                />
              );
            })}
          </svg>

          {/* Cards */}
          {visibleNodes.map((node) => (
            <PolaroidCard
              key={node.id}
              node={node}
              selected={selectedNodeId === node.id}
              dimmed={isNodeDimmed(node.id)}
              onClick={onNodeClick}
              onExpand={onExpand}
              onDragStart={handleNodeDragStart}
              onHover={setHoveredNodeId}
            />
          ))}
        </div>
      </div>
    );
  },
);
