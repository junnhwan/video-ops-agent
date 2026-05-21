import { Routes, Route } from "react-router-dom";
import { ConsoleLayout } from "./components/layout/ConsoleLayout";
import { Dashboard } from "./pages/Dashboard";
import { SessionList } from "./pages/sessions/SessionList";
import { SessionDetail } from "./pages/sessions/SessionDetail";
import { ToolCatalog } from "./pages/tools/ToolCatalog";
import { ToolDetail } from "./pages/tools/ToolDetail";
import { InvocationTrace } from "./pages/invocations/InvocationTrace";
import { SkillList } from "./pages/skills/SkillList";
import { SkillDetail } from "./pages/skills/SkillDetail";
import { SkillCreate } from "./pages/skills/SkillCreate";
import { EvalDashboard } from "./pages/eval/EvalDashboard";

function App() {
  return (
    <Routes>
      <Route element={<ConsoleLayout />}>
        <Route index element={<Dashboard />} />
        <Route path="/sessions" element={<SessionList />} />
        <Route path="/sessions/:id" element={<SessionDetail />} />
        <Route path="/tools" element={<ToolCatalog />} />
        <Route path="/tools/:name" element={<ToolDetail />} />
        <Route path="/invocations" element={<InvocationTrace />} />
        <Route path="/skills" element={<SkillList />} />
        <Route path="/skills/new" element={<SkillCreate />} />
        <Route path="/skills/:id" element={<SkillDetail />} />
        <Route path="/eval" element={<EvalDashboard />} />
      </Route>
    </Routes>
  );
}

export default App;
