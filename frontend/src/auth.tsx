import React, { createContext, useContext, useEffect, useState } from "react";
import { clearToken, decodeRole, getToken } from "./services/session";

type User = { role: "ALCHEMIST" | "SUPERVISOR" } | null;

const Ctx = createContext<{ user: User; logout: () => void }>({ user: null, logout: () => {} } as any);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User>(null);

  useEffect(() => {
    if (getToken()) {
      const role = decodeRole();
      if (role) setUser({ role });
    }
  }, []);

  const logout = () => {
    clearToken();
    setUser(null);
    window.location.href = "/login";
  };

  return <Ctx.Provider value={{ user, logout }}>{children}</Ctx.Provider>;
}

export const useAuth = () => useContext(Ctx);
