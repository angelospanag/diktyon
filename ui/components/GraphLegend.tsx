import { Building2, ShieldAlert, User } from "lucide-react";
import type React from "react";

export function GraphLegend() {
  return (
    <div
      className="absolute bottom-5 left-5 z-20 select-none"
      style={{
        background: "linear-gradient(175deg,#ffffff 0%,#faf6f0 100%)",
        border: "1px solid #d4c4a8",
        borderRadius: 4,
        boxShadow: "2px 4px 16px rgba(40,20,0,0.22)",
        padding: "10px 14px",
        fontFamily: "Inter,system-ui,sans-serif",
        transform: "rotate(-0.8deg)",
        width: 180,
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
        Legend
      </p>

      <Section label="Nodes" />
      <NodeRow
        bg="linear-gradient(155deg,#ffffff 0%,#f5ede0 100%)"
        border="1.5px solid #c8b898"
        icon={<Building2 size={11} strokeWidth={1.8} color="#1a1a2e" />}
        label="Company"
      />
      <NodeRow
        bg="linear-gradient(135deg,#f4f7ff 0%,#dce8f5 100%)"
        border="1.5px solid #8fa8cc"
        icon={<User size={11} strokeWidth={1.8} color="#2a4a7f" />}
        label="Officer"
      />
      <NodeRow
        bg="linear-gradient(155deg,#fff8f8 0%,#fde8e8 100%)"
        border="1.5px dashed #d88888"
        icon={<ShieldAlert size={11} strokeWidth={1.8} color="#8a1a1a" />}
        label="PSC / Beneficial owner"
      />

      <div style={{ height: 1, background: "#e8dece", margin: "9px 0" }} />

      <Section label="Edges" />
      <EdgeRow dashed={false} label="Officer of company" />
      <EdgeRow dashed label="Controls (PSC)" />

      <div style={{ height: 1, background: "#e8dece", margin: "9px 0" }} />

      <Section label="Flags" />
      <p
        style={{
          fontSize: 10.5,
          color: "#1a1a2e",
          opacity: 0.7,
          lineHeight: 1.4,
          marginBottom: 0,
        }}
      >
        ⚠ on a company = overdue filing, insolvency history, or registered
        charges.
      </p>

      <div style={{ height: 1, background: "#e8dece", margin: "9px 0" }} />

      <Section label="History" />
      <p
        style={{
          fontSize: 10.5,
          color: "#1a1a2e",
          opacity: 0.7,
          lineHeight: 1.4,
          marginBottom: 0,
        }}
      >
        Greyed/stamped = resigned or ceased.
        <span data-print-hide>
          {" "}
          Toggle <strong>FORMER</strong> in the header to show them.
        </span>
      </p>

      <div
        data-print-hide
        style={{ height: 1, background: "#e8dece", margin: "9px 0" }}
      />

      <div data-print-hide>
        <Section label="Controls" />
        <ControlRow hint="scroll" label="zoom" />
        <ControlRow hint="drag" label="pan board" />
        <ControlRow hint="drag card" label="repin" />
      </div>
    </div>
  );
}

function Section({ label }: { label: string }) {
  return (
    <p
      style={{
        fontSize: 9,
        fontWeight: 700,
        letterSpacing: "0.1em",
        textTransform: "uppercase",
        color: "#1a1a2e",
        opacity: 0.5,
        marginBottom: 5,
      }}
    >
      {label}
    </p>
  );
}

function NodeRow({
  bg,
  border,
  icon,
  label,
}: {
  bg: string;
  border: string;
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <div
      style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 5 }}
    >
      <div
        style={{
          width: 30,
          height: 17,
          borderRadius: 3,
          flexShrink: 0,
          background: bg,
          border,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          opacity: 0.75,
        }}
      >
        {icon}
      </div>
      <span style={{ fontSize: 10.5, color: "#1a1a2e", fontWeight: 500 }}>
        {label}
      </span>
    </div>
  );
}

function EdgeRow({ dashed, label }: { dashed?: boolean; label: string }) {
  return (
    <div
      style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 5 }}
    >
      <div
        style={{
          width: 30,
          height: 0,
          borderTop: `2px ${dashed ? "dashed" : "solid"} #cc2200`,
          opacity: 0.82,
          flexShrink: 0,
        }}
      />
      <span style={{ fontSize: 10.5, color: "#1a1a2e", fontWeight: 500 }}>
        {label}
      </span>
    </div>
  );
}

function ControlRow({ hint, label }: { hint: string; label: string }) {
  return (
    <div
      style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 5 }}
    >
      <span
        style={{
          fontSize: 8.5,
          fontWeight: 700,
          color: "#5a4a38",
          fontFamily: "JetBrains Mono, monospace",
          letterSpacing: "0.04em",
          background: "#ede3d6",
          borderRadius: 3,
          padding: "2px 5px",
          flexShrink: 0,
          whiteSpace: "nowrap",
        }}
      >
        {hint}
      </span>
      <span style={{ fontSize: 11, color: "#1a1a2e", fontWeight: 500 }}>
        {label}
      </span>
    </div>
  );
}
