"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { getCompanyGraph, getOfficerGraph } from "@/client/sdk.gen";
import type {
  Node as GraphNode,
  Response as GraphResponse,
  SearchResult,
} from "@/client/types.gen";
import { GraphCanvas, type GraphCanvasHandle } from "@/components/GraphCanvas";
import { GraphLegend } from "@/components/GraphLegend";
import { GraphRoster } from "@/components/GraphRoster";
import { NodePanel } from "@/components/NodePanel";
import { Pushpin } from "@/components/Pushpin";
import { SearchBox } from "@/components/SearchBox";
import { isHistorical, mergeMeta, randomLoadingCopy } from "@/lib/graph";

interface HomeProps {
  rootCompany?: string;
}

const EXAMPLES = [
  {
    name: "Marks and Spencer",
    number: "00214436",
    hint: "British institution",
  },
  { name: "Waitrose Limited", number: "00099405", hint: "British grocer" },
  { name: "Waterstones", number: "00610095", hint: "Booksellers" },
  { name: "Joules Limited", number: "02934327", hint: "British fashion brand" },
];

export function Home({ rootCompany }: HomeProps) {
  const router = useRouter();
  const country = "uk";

  const canvasRef = useRef<GraphCanvasHandle>(null);
  const [selected, setSelected] = useState<GraphNode | null>(null);
  const [expanding, setExpanding] = useState(false);
  const [loadingCopy] = useState(randomLoadingCopy);
  const [allNodes, setAllNodes] = useState<GraphNode[]>([]);
  const [showResigned, setShowResigned] = useState(false);

  const {
    data: rootGraph,
    isFetching,
    isError,
  } = useQuery({
    queryKey: ["company-graph", country, rootCompany],
    queryFn: () =>
      getCompanyGraph({
        path: { company_number: rootCompany ?? "" },
        query: { country },
        throwOnError: true,
      }).then((r) => r.data),
    enabled: !!rootCompany,
    staleTime: 5 * 60_000,
  });

  useEffect(() => {
    if (!rootGraph) return;
    canvasRef.current?.setGraph(rootGraph);
    setSelected(null);
    setAllNodes(rootGraph.nodes ?? []);
  }, [rootGraph]);

  const handleSelect = useCallback(
    (result: SearchResult) => {
      router.push(`/${country}/company/${result.company_number}`);
    },
    [router],
  );

  const handleNodeClick = useCallback((node: GraphNode) => {
    setSelected(node);
  }, []);

  const handleExpand = useCallback(async (node: GraphNode) => {
    setExpanding(true);
    try {
      let graph: GraphResponse | undefined;
      if (node.type === "company" && node.meta.company_number)
        graph = (
          await getCompanyGraph({
            path: { company_number: node.meta.company_number },
            query: { country },
            throwOnError: true,
          })
        ).data;
      else if (node.type === "officer" && node.meta.officer_id)
        graph = (
          await getOfficerGraph({
            path: { officer_id: node.meta.officer_id },
            query: { country },
            throwOnError: true,
          })
        ).data;
      else return;

      canvasRef.current?.addGraph(graph);
      canvasRef.current?.markExpanded(node.id);
      setAllNodes((prev) => {
        const existingById = new Map(prev.map((n) => [n.id, n]));
        const incoming = graph?.nodes ?? [];
        const updated = prev.map((n) => {
          const incomingNode = incoming.find((i) => i.id === n.id);
          if (!incomingNode) return n;
          return {
            ...incomingNode,
            meta: mergeMeta(n.meta, incomingNode.meta),
          };
        });
        const added = incoming.filter((n) => !existingById.has(n.id));
        return [...updated, ...added];
      });
      setSelected((prev) => {
        if (!prev) return prev;
        const updated = (graph?.nodes ?? []).find((n) => n.id === prev.id);
        if (!updated)
          return { ...prev, meta: { ...prev.meta, expanded: true } };
        return { ...updated, meta: mergeMeta(prev.meta, updated.meta) };
      });
    } finally {
      setExpanding(false);
    }
  }, []);

  const isLoading = isFetching && !rootGraph;

  // Resigned officers / ceased PSCs are hidden by default — mirrors the
  // GraphCanvas toggle so the roster and the board stay in sync.
  const rosterNodes = showResigned
    ? allNodes
    : allNodes.filter((n) => !isHistorical(n));

  return (
    <div
      className="flex h-full flex-col"
      style={{ fontFamily: "var(--font-body)" }}
    >
      {/* Header */}
      <header
        data-print-hide
        className="relative z-30 flex items-center gap-6 px-6 py-3 shadow-md"
        style={{
          background: "linear-gradient(180deg, #faf6f0 0%, #f2ebe0 100%)",
          borderBottom: "1px solid #d4c4a8",
        }}
      >
        <a href="/" className="shrink-0 no-underline">
          <div
            className="text-2xl font-black tracking-tighter text-[--color-string]"
            style={{
              fontFamily: "var(--font-display)",
              letterSpacing: "-0.03em",
            }}
          >
            DIKTYON
          </div>
          <div
            className="text-[10px] font-semibold tracking-widest text-stone-400 uppercase"
            style={{
              fontFamily: "var(--font-display)",
              letterSpacing: "0.12em",
            }}
          >
            Map the corporate network
          </div>
        </a>

        <div className="h-8 w-px bg-stone-300" />

        <SearchBox onSelect={handleSelect} />

        {rootGraph && (
          <div className="ml-auto flex items-center gap-2">
            <button
              type="button"
              onClick={() => setShowResigned((v) => !v)}
              title="Show or hide resigned officers and ceased PSCs"
              style={{
                display: "flex",
                alignItems: "center",
                gap: 6,
                padding: "6px 12px",
                background: showResigned ? "#ede3d6" : "transparent",
                border: "1px solid #d4c4a8",
                borderRadius: 5,
                cursor: "pointer",
                fontFamily: "var(--font-display)",
                fontSize: 11,
                fontWeight: 700,
                letterSpacing: "0.06em",
                color: "#5a4a38",
              }}
            >
              🕵️ {showResigned ? "FORMER: ON" : "FORMER: OFF"}
            </button>
            <button
              type="button"
              onClick={() => window.print()}
              title="Export as PDF"
              style={{
                display: "flex",
                alignItems: "center",
                gap: 6,
                padding: "6px 12px",
                background: "transparent",
                border: "1px solid #d4c4a8",
                borderRadius: 5,
                cursor: "pointer",
                fontFamily: "var(--font-display)",
                fontSize: 11,
                fontWeight: 700,
                letterSpacing: "0.06em",
                color: "#5a4a38",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = "#ede3d6";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = "transparent";
              }}
            >
              <svg
                aria-hidden="true"
                width={14}
                height={14}
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth={2}
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <polyline points="6 9 6 2 18 2 18 9" />
                <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2" />
                <rect x={6} y={14} width={12} height={8} />
              </svg>
              PRINT
            </button>
          </div>
        )}
      </header>

      {/* Canvas */}
      <main className="relative flex flex-1 overflow-hidden">
        {isLoading && (
          <div
            data-print-hide
            className="absolute inset-0 z-10 flex items-center justify-center cork-board"
          >
            <div
              className="flex flex-col items-center gap-4 rounded-sm bg-[--color-manila]/90 px-10 py-8 shadow-2xl"
              style={{ border: "1px solid #c9bfb2" }}
            >
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-stone-300 border-t-[--color-string]" />
              <p
                className="text-sm font-semibold text-[--color-navy]"
                style={{ fontFamily: "var(--font-display)" }}
              >
                {loadingCopy}
              </p>
            </div>
          </div>
        )}

        {isError && (
          <div className="absolute inset-0 z-20 flex items-center justify-center cork-board">
            <div
              className="relative w-80 rounded-sm px-7 py-6"
              style={{
                background: "#ffffff",
                border: "1px solid #c9bfb2",
                boxShadow:
                  "0 8px 32px rgba(0,0,0,0.22), 0 2px 8px rgba(0,0,0,0.12)",
              }}
            >
              <div className="absolute -top-[19px] left-1/2 -translate-x-1/2">
                <Pushpin color="var(--color-string)" size={18} />
              </div>
              <h2
                className="mb-1 text-sm font-black uppercase tracking-widest text-[--color-string]"
                style={{ fontFamily: "var(--font-display)" }}
              >
                Company not found
              </h2>
              <p className="text-xs leading-relaxed text-stone-500">
                No company matched that number. Use the search box in the header
                to look up a company by name.
              </p>
            </div>
          </div>
        )}

        {!rootCompany && (
          <div className="absolute inset-0 z-20 flex items-center justify-center cork-board">
            <div
              className="relative w-80 rounded-sm px-7 py-6"
              style={{
                background: "#ffffff",
                border: "1px solid #c9bfb2",
                boxShadow:
                  "0 8px 32px rgba(0,0,0,0.22), 0 2px 8px rgba(0,0,0,0.12)",
              }}
            >
              {/* pushpin */}
              <div className="absolute -top-[19px] left-1/2 -translate-x-1/2">
                <Pushpin color="var(--color-string)" size={18} />
              </div>

              <h2
                className="mb-1 text-sm font-black uppercase tracking-widest text-[--color-string]"
                style={{ fontFamily: "var(--font-display)" }}
              >
                Diktyon
              </h2>
              <p className="mb-4 text-xs leading-relaxed text-stone-500">
                Use the search box above to look up any UK company by name.
                Click a node to inspect it — then hit Expand to follow the
                network further. Toggle FORMER on the graph to reveal resigned
                officers and ceased PSCs.
              </p>

              <p
                className="mb-2 text-[10px] font-semibold uppercase tracking-widest text-stone-400"
                style={{ fontFamily: "var(--font-display)" }}
              >
                Try an example
              </p>
              <div className="flex flex-col gap-2">
                {EXAMPLES.map((ex) => (
                  <button
                    type="button"
                    key={ex.number}
                    onClick={() => router.push(`/uk/company/${ex.number}`)}
                    className="flex items-center justify-between rounded-sm px-3 py-2 text-left hover:bg-stone-100"
                    style={{
                      border: "1px solid #d4c4a8",
                      background: "transparent",
                      cursor: "pointer",
                    }}
                  >
                    <span className="text-xs font-semibold text-[--color-navy]">
                      {ex.name}
                    </span>
                    <span className="text-[10px] text-stone-400">
                      {ex.hint}
                    </span>
                  </button>
                ))}
              </div>
            </div>
          </div>
        )}

        <GraphCanvas
          ref={canvasRef}
          onNodeClick={handleNodeClick}
          onExpand={handleExpand}
          selectedNodeId={selected?.id}
          showResigned={showResigned}
        />

        <GraphLegend />

        <GraphRoster
          nodes={rosterNodes}
          onHoverNode={(id) => canvasRef.current?.setHoveredNode(id)}
          onClickNode={handleNodeClick}
        />

        {selected && (
          <div data-print-hide className="absolute right-6 top-6 z-20">
            <NodePanel
              node={selected}
              isExpanding={expanding}
              onExpand={handleExpand}
              onClose={() => setSelected(null)}
            />
          </div>
        )}
      </main>

      {/* Footer */}
      <footer
        data-print-hide
        className="shrink-0 flex items-center justify-center gap-3 px-6 py-2 text-xs"
        style={{
          background: "linear-gradient(180deg, #f2ebe0 0%, #ede3d6 100%)",
          borderTop: "1px solid #d4c4a8",
          color: "#6a5a4a",
          fontFamily: "var(--font-body)",
          letterSpacing: "0.01em",
        }}
      >
        <span>
          Built by{" "}
          <a
            href="https://github.com/angelospanag"
            target="_blank"
            rel="noopener noreferrer"
            className="font-semibold hover:underline underline-offset-2"
            style={{ color: "#5a4a38" }}
          >
            Angelos Panagiotopoulos
          </a>
        </span>
        <span style={{ opacity: 0.4 }}>·</span>
        <a
          href="https://github.com/angelospanag/diktyon"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:underline underline-offset-2"
          style={{
            color: "#cc2200",
            fontFamily: "var(--font-display)",
            fontSize: 10,
            letterSpacing: "0.04em",
          }}
        >
          github.com/angelospanag/diktyon
        </a>
        <span style={{ opacity: 0.4 }}>·</span>
        <a
          href="https://developer-specs.company-information.service.gov.uk/companies-house-public-data-api/reference"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:underline underline-offset-2"
        >
          Powered by Companies House
        </a>
        <span style={{ opacity: 0.4 }}>·</span>
        <a
          href="https://ko-fi.com/angelospanag"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:underline underline-offset-2"
          style={{ color: "#cc2200", fontWeight: 600 }}
        >
          ☕ Buy me a coffee
        </a>
        <span style={{ opacity: 0.4 }}>·</span>
        <span>No cookies. No tracking. No data collected.</span>
      </footer>
    </div>
  );
}
