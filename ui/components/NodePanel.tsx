import type { Node as GraphNode } from "@/client/types.gen";
import { Pushpin } from "@/components/Pushpin";
import { chLink, distressFlags, PIN_COLOR, tiltAngle } from "@/lib/graph";

interface Props {
  node: GraphNode;
  isExpanding: boolean;
  onExpand: (node: GraphNode) => void;
  onClose: () => void;
}

const TYPE_LABEL: Record<GraphNode["type"], string> = {
  company: "Company",
  officer: "Officer",
  psc: "PSC",
};

const TYPE_COLOUR: Record<GraphNode["type"], string> = {
  company: "bg-amber-100 text-amber-800 border-amber-200",
  officer: "bg-sky-100   text-sky-800   border-sky-200",
  psc: "bg-rose-100  text-rose-800  border-rose-200",
};

export function NodePanel({ node, isExpanding, onExpand, onClose }: Props) {
  const link = chLink(node);
  const flags = distressFlags(node);
  const canExpand =
    (node.type === "company" && !!node.meta.company_number) ||
    (node.type === "officer" && !!node.meta.officer_id);

  const tilt = tiltAngle(node.id);

  return (
    <div
      className="relative w-60"
      style={{ transform: `rotate(${tilt}deg)`, transformOrigin: "top center" }}
    >
      {/* Pushpin */}
      <div className="absolute -top-[22px] left-1/2 z-10 -translate-x-1/2">
        <Pushpin color={PIN_COLOR[node.type]} size={20} />
      </div>

      {/* Card */}
      <div
        className="flex flex-col gap-2 rounded-sm pt-5 pb-4 px-4 shadow-2xl overflow-y-auto"
        style={{
          background: "linear-gradient(175deg, #ffffff 0%, #faf6f0 100%)",
          border: "1px solid #d4c4a8",
          boxShadow:
            "0 8px 32px rgba(40,20,0,0.35), 0 2px 6px rgba(40,20,0,0.15)",
          maxHeight: "calc(100vh - 120px)",
        }}
      >
        {/* Close */}
        <button
          type="button"
          onClick={onClose}
          className="absolute right-2.5 top-2.5 rounded p-0.5 text-stone-400 hover:text-stone-600"
          aria-label="Close"
        >
          ✕
        </button>

        {/* Type badge + label */}
        <div className="flex flex-col gap-1 border-b border-stone-200 pb-2">
          <span
            className={`self-start rounded border px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-widest ${TYPE_COLOUR[node.type]}`}
          >
            {TYPE_LABEL[node.type]}
          </span>
          <h2 className="text-sm font-semibold leading-snug text-[--color-navy]">
            {node.label}
          </h2>
        </div>

        {/* Metadata */}
        <dl className="grid grid-cols-[auto_1fr] gap-x-2 gap-y-1 text-[11px]">
          {node.meta.company_number && (
            <Row label="Number" value={node.meta.company_number} mono />
          )}
          {node.meta.status && <Row label="Status" value={node.meta.status} />}
          {node.meta.company_type && (
            <Row label="Type" value={node.meta.company_type} />
          )}
          {node.meta.incorporation_date && (
            <Row
              label="Incorporated"
              value={node.meta.incorporation_date}
              mono
            />
          )}
          {node.meta.address && (
            <Row label="Address" value={node.meta.address} />
          )}
          {node.meta.sic_codes && node.meta.sic_codes.length > 0 && (
            <Row label="SIC code" value={node.meta.sic_codes.join(", ")} mono />
          )}
          {node.meta.role && <Row label="Role" value={node.meta.role} />}
          {node.meta.occupation && (
            <Row label="Occupation" value={node.meta.occupation} />
          )}
          {node.meta.appointed_on && (
            <Row label="Appointed" value={node.meta.appointed_on} mono />
          )}
          {node.meta.resigned_on && (
            <Row label="Resigned" value={node.meta.resigned_on} mono />
          )}
          {node.meta.nationality && (
            <Row label="Nationality" value={node.meta.nationality} />
          )}
          {node.meta.notified_on && (
            <Row label="PSC since" value={node.meta.notified_on} mono />
          )}
          {node.meta.ceased_on && (
            <Row label="Ceased" value={node.meta.ceased_on} mono />
          )}
          {node.meta.natures_of_control?.map((n, i) => (
            <Row
              key={n}
              label={i === 0 ? "Control" : ""}
              value={n.replaceAll("-", " ")}
            />
          ))}
        </dl>

        {/* Distress flags — overdue filings / insolvency history / charges */}
        {flags.length > 0 && (
          <div className="flex flex-col gap-0.5 border-t border-stone-200 pt-2 text-[11px]">
            {flags.map((flag) => (
              <p key={flag} className="font-medium text-rose-700">
                ⚠ {flag}
              </p>
            ))}
          </div>
        )}

        {/* Actions */}
        <div className="flex flex-col gap-1.5 border-t border-stone-200 pt-2">
          {canExpand && (
            <button
              type="button"
              onClick={() => onExpand(node)}
              disabled={isExpanding || node.meta.expanded}
              className="rounded px-2.5 py-1.5 text-[11px] font-semibold shadow transition-opacity"
              style={
                node.meta.expanded
                  ? {
                      background: "#e7e5e4",
                      color: "#57534e",
                      border: "1px solid #d6d3d1",
                      cursor: "default",
                    }
                  : isExpanding
                    ? {
                        background: "#e7e5e4",
                        color: "#78716c",
                        border: "1px solid #d6d3d1",
                        cursor: "default",
                      }
                    : {
                        background: "#cc2200",
                        color: "#ffffff",
                        cursor: "pointer",
                      }
              }
            >
              {isExpanding
                ? "Expanding…"
                : node.meta.expanded
                  ? "✓ Already expanded"
                  : "Expand network →"}
            </button>
          )}
          {node.meta.address && (
            <a
              href={`https://www.openstreetmap.org/search?query=${encodeURIComponent(node.meta.address)}`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-center text-[11px] text-[--color-string] underline-offset-2 hover:underline"
            >
              View on map ↗
            </a>
          )}
          {node.meta.registry_url ? (
            <a
              href={node.meta.registry_url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-center text-[11px] text-[--color-string] underline-offset-2 hover:underline"
            >
              View on GEMI ↗
            </a>
          ) : link ? (
            <a
              href={link}
              target="_blank"
              rel="noopener noreferrer"
              className="text-center text-[11px] text-[--color-string] underline-offset-2 hover:underline"
            >
              View on Companies House ↗
            </a>
          ) : null}
        </div>
      </div>

      {/* Tape strip */}
      <div
        className="absolute -bottom-2 left-1/2 h-4 w-16 -translate-x-1/2 -rotate-1 rounded-sm opacity-70"
        style={{
          background: "rgba(255,240,150,0.85)",
          border: "1px solid rgba(200,185,80,0.5)",
          boxShadow: "inset 0 0 4px rgba(0,0,0,0.06)",
        }}
        aria-hidden
      />
    </div>
  );
}

function Row({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <>
      <dt className="font-medium text-stone-400">{label}</dt>
      <dd className={`text-[--color-navy] ${mono ? "font-mono" : ""}`}>
        {value}
      </dd>
    </>
  );
}
