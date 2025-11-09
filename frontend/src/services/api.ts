// src/services/api.ts

// ======================
// 🔹 Tipos de entidades
// ======================

export interface Alchemist {
  id: number;
  name: string;
  specialty: string;
  rank: string;
  age?: number;
  email?: string | null;
  created_at?: string;
}

export interface Mission {
  id: number;
  title: string;
  description: string;
  assigned_to: number | null;
  status: string;
  created_at?: string;
}

export interface Material {
  id: number;
  name: string;
  stock: number;
  created_at?: string;
}

export interface Transmutation {
  id: number;
  alchemist_id: number;
  description: string;
  status: string;
  created_at?: string;
}

export interface Audit {
  id: number;
  action: string;
  entity: string;
  entity_id: number;
  created_at?: string;
}

// ======================
// 🌐 Configuración base
// ======================

const BASE = "http://localhost:8000";

async function http<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers || {}),
    },
  });

  // Manejo de errores
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`HTTP ${res.status}: ${text}`);
  }

  // Algunos endpoints DELETE o vacíos pueden no devolver JSON
  try {
    return (await res.json()) as T;
  } catch {
    return {} as T;
  }
}

// ======================
// ⚗️ Alchemists
// ======================

export const getAlchemists = () => http<Alchemist[]>(`${BASE}/alchemists`);

export const createAlchemist = (data: Partial<Omit<Alchemist, "id" | "created_at">>) => {
  // El backend convierte email vacío en NULL, así que mandamos null si está vacío
  const sanitized = {
    ...data,
    email: data.email && data.email.trim() !== "" ? data.email : null,
    age: data.age ?? 0,
  };
  return http<Alchemist>(`${BASE}/alchemists`, {
    method: "POST",
    body: JSON.stringify(sanitized),
  });
};

export const updateAlchemist = (id: number, data: Partial<Omit<Alchemist, "id">>) => {
  const sanitized = {
    ...data,
    email: data.email && data.email.trim() !== "" ? data.email : null,
    age: data.age ?? 0,
  };
  return http<Alchemist>(`${BASE}/alchemists/${id}`, {
    method: "PUT",
    body: JSON.stringify(sanitized),
  });
};

export const deleteAlchemist = (id: number) =>
  fetch(`${BASE}/alchemists/${id}`, { method: "DELETE" }).then(() => undefined);

// ======================
// 🧱 Materials
// ======================

export const getMaterials = () => http<Material[]>(`${BASE}/materials`);

export const createMaterial = (data: Omit<Material, "id" | "created_at">) =>
  http<Material>(`${BASE}/materials`, {
    method: "POST",
    body: JSON.stringify(data),
  });

export const updateMaterial = (id: number, data: Partial<Material>) =>
  http<Material>(`${BASE}/materials/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });

export const deleteMaterial = (id: number) =>
  fetch(`${BASE}/materials/${id}`, { method: "DELETE" }).then(() => undefined);

// ======================
// 🧭 Missions
// ======================

export const getMissions = () => http<Mission[]>(`${BASE}/missions`);

export const createMission = (data: Omit<Mission, "id" | "created_at" | "status">) =>
  http<Mission>(`${BASE}/missions`, {
    method: "POST",
    body: JSON.stringify(data),
  });

export const updateMission = (id: number, data: Partial<Mission>) =>
  http<Mission>(`${BASE}/missions/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });

export const deleteMission = (id: number) =>
  fetch(`${BASE}/missions/${id}`, { method: "DELETE" }).then(() => undefined);

// ======================
// 🔮 Transmutations
// ======================

export const getTransmutations = () => http<Transmutation[]>(`${BASE}/transmutations`);

// ⚠️ Importante: el backend espera POST /transmutations/:alchemistId
export const startTransmutation = (alchemistId: number, description: string) =>
  http<Transmutation>(`${BASE}/transmutations/${alchemistId}`, {
    method: "POST",
    body: JSON.stringify({ description }),
  });

// ======================
// 🧾 Audits
// ======================

export const getAudits = () => http<Audit[]>(`${BASE}/audits`);
