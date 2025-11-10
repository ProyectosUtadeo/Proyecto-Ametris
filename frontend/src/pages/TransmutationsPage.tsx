// src/pages/TransmutationsPage.tsx
import { useEffect, useMemo, useRef, useState } from "react";
import type { Alchemist, Transmutation } from "../services/api";
import { getAlchemists, getTransmutations, startTransmutation } from "../services/api";

type WsEnvelope =
  | { type: "transmutation:started" | "transmutation:updated" | "transmutation:completed" | "transmutation:cancelled"; data: Transmutation }
  | { type: string; data: any };

export default function TransmutationsPage() {
  const [list, setList] = useState<Transmutation[]>([]);
  const [alchs, setAlchs] = useState<Alchemist[]>([]);
  const [alchemistId, setAlchemistId] = useState<number | "">("");
  const [desc, setDesc] = useState("");

  // URL del WS derivada del host actual (no uses React.default)
  const wsUrl = useMemo(() => {
    const proto = window.location.protocol === "https:" ? "wss" : "ws";
    const host = window.location.hostname;
    const port = "8000"; // cambia si tu backend expone otro
    return `${proto}://${host}:${port}/ws`;
  }, []);

  const wsRef = useRef<WebSocket | null>(null);

  const load = async () => {
    const [t, a] = await Promise.all([getTransmutations(), getAlchemists()]);
    setList(t);
    setAlchs(a);
  };

  useEffect(() => {
    load();
  }, []);

  // Conexión WebSocket (solo dentro del componente)
  useEffect(() => {
    let ws: WebSocket | null = null;
    let reconnectTimer: number | undefined;

    const connect = () => {
      ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        // opcional: console.log("WS connected");
      };

      ws.onmessage = (evt) => {
        try {
          const msg: WsEnvelope = JSON.parse(evt.data);
          if (
            msg.type === "transmutation:started" ||
            msg.type === "transmutation:updated" ||
            msg.type === "transmutation:completed" ||
            msg.type === "transmutation:cancelled"
          ) {
            setList((prev) => {
              const t = msg.data;
              const idx = prev.findIndex((x) => x.id === t.id);
              if (idx === -1) return [t, ...prev];
              const copy = prev.slice();
              copy[idx] = t;
              return copy;
            });
          }
        } catch {
          // ignora mensajes no JSON
        }
      };

      ws.onclose = () => {
        // Reintento simple
        reconnectTimer = window.setTimeout(connect, 1500);
      };
    };

    connect();

    return () => {
      if (reconnectTimer) clearTimeout(reconnectTimer);
      if (wsRef.current?.readyState === WebSocket.OPEN) wsRef.current.close();
      wsRef.current = null;
    };
  }, [wsUrl]);

  const onStart = async () => {
    if (!alchemistId) return;
    await startTransmutation(Number(alchemistId), desc || "Generic transmutation");
    setDesc("");
    // no hace falta recargar, el WS llegará; pero mantenemos por si el WS cae
    await load();
  };

  const getAlchName = (id: number) => {
    const found = alchs.find((a) => a.id === id);
    return found ? found.name : `#${id}`;
  };

  return (
    <div style={{ padding: 20 }}>
      <h2>⚗️ Transmutations</h2>

      <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <select
          value={alchemistId}
          onChange={(e) => setAlchemistId(e.target.value ? Number(e.target.value) : "")}
        >
          <option value="">Select Alchemist</option>
          {alchs
            .filter((a) => typeof a.id === "number")
            .map((a, i) => (
              <option key={`alch-${a.id ?? "noid"}-${i}`} value={a.id}>
                {a.name}
              </option>
            ))}
        </select>

        <input
          style={{ flex: 1 }}
          placeholder="Description"
          value={desc}
          onChange={(e) => setDesc(e.target.value)}
        />
        <button onClick={onStart}>Start</button>
      </div>

      <table width="100%" cellPadding={6} style={{ borderCollapse: "collapse" }}>
        <thead style={{ background: "#eee" }}>
          <tr>
            <th align="left">ID</th>
            <th align="left">Alchemist</th>
            <th align="left">Description</th>
            <th align="left">Status</th>
            <th align="left">Created</th>
          </tr>
        </thead>
        <tbody>
          {list.length ? (
            list.map((t) => (
              <tr key={`trm-${t.id}`}>
                <td>{t.id}</td>
                <td>{t.alchemist?.name ?? getAlchName(t.alchemist_id)}</td>
                <td>{t.description}</td>
                <td>{t.status}</td>
                <td>{t.created_at ? new Date(t.created_at).toLocaleString() : "-"}</td>
              </tr>
            ))
          ) : (
            <tr key="empty">
              <td colSpan={5} align="center">
                No data
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
