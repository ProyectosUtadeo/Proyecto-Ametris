// src/App.tsx
import { BrowserRouter, Route, Routes } from "react-router-dom";
import Layout from "./Layout";
import Dashboard from "./pages/Dashboard";
import AlchemistsPage from "./pages/AlchemistsPage";
import MaterialsPage from "./pages/MaterialsPage";
import MissionsPage from "./pages/MissionsPage";
import TransmutationsPage from "./pages/TransmutationsPage";
import AuditsPage from "./pages/AuditsPage";
import RegisterPage from "./pages/RegisterPage";


import LoginPage from "./pages/LoginPage";           // 👈 nuevo
import { PrivateRoute, RoleRoute } from "./routeGuards"; // 👈 nuevo
import { AuthProvider } from "./auth";               // 👈 nuevo

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Ruta pública para login */}
          
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          {/* Todo lo demás protegido por JWT */}
          <Route
            element={
              <PrivateRoute>
                <Layout />
              </PrivateRoute>
            }
          >
            <Route path="/" element={<Dashboard />} />
            <Route path="/alchemists" element={<AlchemistsPage />} />
            <Route path="/materials" element={<MaterialsPage />} />
            <Route path="/missions" element={<MissionsPage />} />
            <Route path="/transmutations" element={<TransmutationsPage />} />

            {/* Si quieres proteger /audits solo para SUPERVISOR, descomenta: */}
            {/* <Route
              path="/audits"
              element={
                <RoleRoute role="SUPERVISOR">
                  <AuditsPage />
                </RoleRoute>
              }
            /> */}

            {/* Por ahora lo dejamos abierto dentro del área autenticada */}
            <Route path="/audits" element={<AuditsPage />} />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
