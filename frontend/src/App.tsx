// src/App.tsx
import { BrowserRouter, Route, Routes } from "react-router-dom";
import Layout from "./Layout";
import Dashboard from "./pages/Dashboard";
import AlchemistsPage from "./pages/AlchemistsPage";
import MaterialsPage from "./pages/MaterialsPage";
import MissionsPage from "./pages/MissionsPage";
import TransmutationsPage from "./pages/TransmutationsPage";
import AuditsPage from "./pages/AuditsPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/alchemists" element={<AlchemistsPage />} />
          <Route path="/materials" element={<MaterialsPage />} />
          <Route path="/missions" element={<MissionsPage />} />
          <Route path="/transmutations" element={<TransmutationsPage />} />
          <Route path="/audits" element={<AuditsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
