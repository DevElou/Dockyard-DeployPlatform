import { PageHeader } from "@/components/layout/page-header";
import { SettingsNav } from "@/components/settings/settings-nav";

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="Settings"
        description="Manage your Dockyard platform configuration"
      />
      <div className="flex flex-1 gap-8 p-6 min-h-0">
        <SettingsNav />
        <div className="flex-1 min-w-0 space-y-4">{children}</div>
      </div>
    </div>
  );
}
