// src/pages/TransmutationsPage.tsx
import { useEffect, useState } from "react";
import type { Alchemist, Transmutation } from "../services/api";
import { getAlchemists, getTransmutations, startTransmutation } from "../services/api";

export default function TransmutationsPage() {
  const [list, setList] = useState<Transmutation[]>([]);
  const [alchs, setAlchs] = useState<Alchemist[]>([]);
  const [alchemistId, setAlchemistId] = useState<number | "">("");
  const [desc, setDesc] = useState("");

  const load = async () => {
    const [t, a] = await Promise.all([getTransmutations(), getAlchemists()]);
    setList(t);
    setAlchs(a);
  };

  useEffect(() => {
    load();
    const t = setInterval(load, 3000);
    return () => clearInterval(t);
  }, []);

  const onStart = async () => {
    if (!alchemistId) return;
    await startTransmutation(Number(alchemistId), desc || "Generic transmutation");
    setDesc("");
    await load();
  };

  const getAlchName = (id: number) => {
    const found = alchs.find(a => a.id === id);
    return found ? found.name : `#${id}`;
    // si quieres: `${found?.name ?? `#${id}`}`
  };

  return (
    <div style={{ padding: 20 }}>
      <h2>⚗️ Transmutations</h2>

      <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <select
          value={alchemistId}
          onChange={e => setAlchemistId(e.target.value ? Number(e.target.value) : "")}
        >
          <option value="">Select Alchemist</option>
          {alchs
            .filter(a => typeof a.id === "number")
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
          onChange={e => setDesc(e.target.value)}
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
            list.map(t => (
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
