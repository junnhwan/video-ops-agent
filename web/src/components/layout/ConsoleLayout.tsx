import { Outlet } from "react-router-dom";
import { Sidebar } from "./Sidebar";

export function ConsoleLayout() {
  return (
    <div className="flex h-screen overflow-hidden bg-[var(--color-surface)]">
      <Sidebar />
      <main className="flex-1 overflow-hidden flex flex-col">
        <div className="flex-1 overflow-auto scroll-panel">
          <div className="page-enter p-6 max-w-[1600px] mx-auto w-full">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}
