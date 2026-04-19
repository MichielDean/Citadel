import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './components/Sidebar';
import { Header } from './components/Header';
import { useDashboardEvents } from './hooks/useDashboardEvents';

export function AppLayout() {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const { data, connected } = useDashboardEvents();

  return (
    <div className="h-screen flex overflow-hidden bg-cistern-bg">
      <Sidebar open={sidebarOpen} onToggle={() => setSidebarOpen(!sidebarOpen)} />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <Header data={data} connected={connected} onMenuClick={() => setSidebarOpen(!sidebarOpen)} />
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}