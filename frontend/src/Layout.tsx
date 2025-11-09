// src/Layout.tsx
import { Link, Outlet, useLocation } from "react-router-dom";

export default function Layout() {
  const loc = useLocation();
  const Tab = ({ to, label }: { to: string; label: string }) => (
    <Link
      to={to}
      style={{
        padding: "6px 10px",
        borderRadius: 6,
        border: "1px solid #ddd",
        background: loc.pathname === to ? "#222" : "#f6f6f6",
        color: loc.pathname === to ? "#fff" : "#222",
        textDecoration: "none",
        fontSize: 14,
      }}
    >
      {label}
    </Link>
  );
  return (
    <div>
      <div style={{ display: "flex", gap: 12, padding: 12, borderBottom: "1px solid #eee" }}>
        <Tab to="/" label="Dashboard" />
        <Tab to="/alchemists" label="Alchemists" />
        <Tab to="/materials" label="Materials" />
        <Tab to="/missions" label="Missions" />
        <Tab to="/transmutations" label="Transmutations" />
        <Tab to="/audits" label="Audits" />
      </div>
      <Outlet />
    </div>
  );
}
