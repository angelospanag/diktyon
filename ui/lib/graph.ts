import type { Node as GraphNode, NodeMeta } from "@/client/types.gen";

// Deterministic per-node Polaroid tilt, derived purely from the node id (a
// djb2-style string hash) so the same node always tilts the same way across
// sessions and clients without the backend needing to store or send state.
export function tiltAngle(id: string): number {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = (hash * 33 + id.charCodeAt(i)) | 0;
  }
  const seed = Math.abs(hash) % 1000;
  return ((seed - 500) / 500) * 7.5; // ±7.5 degrees
}

// Pin colour by node type — shared between the graph cards and the
// selected-node panel so the same entity always gets the same pin colour.
export const PIN_COLOR: Record<GraphNode["type"], string> = {
  company: "#d48b0a",
  officer: "#1a66cc",
  psc: "#cc2200",
};

export function chLink(node: GraphNode): string | null {
  if (node.type === "company" && node.meta.company_number)
    return `https://find-and-update.company-information.service.gov.uk/company/${node.meta.company_number}`;
  if (node.type === "officer" && node.meta.officer_id)
    return `https://find-and-update.company-information.service.gov.uk/officers/${node.meta.officer_id}/appointments`;
  return null;
}

const COPY = [
  "Following the money…",
  "Cross-referencing the filings…",
  "Reviewing the appointments…",
  "Pulling on the threads…",
  "Connecting the dots…",
  "Checking the register…",
  "Tracing the paper trail…",
];

export function randomLoadingCopy(): string {
  return COPY[Math.floor(Math.random() * COPY.length)];
}

// ── Metadata merge ───────────────────────────────────────────────────────────
// Expansion endpoints (e.g. ForOfficer) return sparse nodes — never overwrite
// a non-empty existing field with an empty incoming one.

export function mergeMeta(existing: NodeMeta, incoming: NodeMeta): NodeMeta {
  const merged = { ...existing };
  for (const [k, v] of Object.entries(incoming)) {
    if (k === "expanded") {
      if (v === true) merged.expanded = true;
    } else if (
      v !== undefined &&
      v !== null &&
      v !== "" &&
      !(Array.isArray(v) && (v as unknown[]).length === 0)
    ) {
      (merged as Record<string, unknown>)[k] = v;
    }
  }
  return merged;
}

// ── Resigned officers / ceased PSCs ─────────────────────────────────────────
// Hidden by default (cleaner "who controls this now" view); a toggle reveals
// them for investigators digging into history.

export function isHistorical(node: GraphNode): boolean {
  return !!(node.meta.resigned_on || node.meta.ceased_on);
}

export function historicalLabel(node: GraphNode): string | null {
  if (node.meta.resigned_on) return `Resigned ${node.meta.resigned_on}`;
  if (node.meta.ceased_on) return `Ceased ${node.meta.ceased_on}`;
  return null;
}

// ── Distress flags ──────────────────────────────────────────────────────────
// Straight from the company profile we already fetch for the root node — no
// extra API calls. Only populated for nodes that came from a direct profile
// fetch (the root company), not for companies discovered via officer/PSC
// fan-out, since those never get their own profile request.

export function distressFlags(node: GraphNode): string[] {
  if (node.type !== "company") return [];
  const flags: string[] = [];
  if (node.meta.accounts_overdue) flags.push("Accounts overdue");
  if (node.meta.confirmation_statement_overdue)
    flags.push("Confirmation statement overdue");
  if (node.meta.has_insolvency_history) flags.push("Insolvency history");
  if (node.meta.has_charges) flags.push("Registered charges");
  return flags;
}
