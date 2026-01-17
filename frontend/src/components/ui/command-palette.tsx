import React from "react";
import { useNavigate } from "react-router-dom";
import { useList } from "@refinedev/core";
import {
    CommandDialog as Dialog,
    CommandInput,
    CommandList,
    CommandEmpty,
    CommandGroup,
    CommandItem,
    CommandSeparator,
    CommandShortcut,
} from "@/components/ui/command";
import {
    Box,
    Database,
    Rocket,
    FolderKanban,
    Key,
    Settings,
    FileText,
    BarChart3,
    Activity,
    Plus,
} from "lucide-react";

interface CommandDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export const CommandDialog: React.FC<CommandDialogProps> = ({
    open,
    onOpenChange,
}) => {
    const navigate = useNavigate();
    const [search, setSearch] = React.useState("");

    // Fetch recent data for search (mockable if providers not ready)
    const { data: applications } = useList({
        resource: "applications",
        pagination: { pageSize: 5 },
        filters: search
            ? [{ field: "name", operator: "contains", value: search }]
            : [],
        queryOptions: {
            enabled: !!search
        }
    }) as any;

    const { data: databases } = useList({
        resource: "databases",
        pagination: { pageSize: 5 },
        filters: search
            ? [{ field: "name", operator: "contains", value: search }]
            : [],
        queryOptions: {
            enabled: !!search
        }
    }) as any;

    const runCommand = (command: () => void) => {
        onOpenChange(false);
        command();
    };

    const navigationItems = [
        { name: "Dashboard", icon: Box, path: "/", shortcut: "⌘D" },
        { name: "Applications", icon: Box, path: "/applications", shortcut: "⌘A" },
        { name: "Databases", icon: Database, path: "/databases", shortcut: "⌘B" },
        { name: "Deployments", icon: Rocket, path: "/deployments" },
        { name: "Projects", icon: FolderKanban, path: "/projects" },
        { name: "Secrets", icon: Key, path: "/secrets" },
        { name: "Logs", icon: FileText, path: "/logs", shortcut: "⌘L" },
        { name: "Metrics", icon: BarChart3, path: "/metrics", shortcut: "⌘M" },
        { name: "Traces", icon: Activity, path: "/traces" },
        { name: "Settings", icon: Settings, path: "/settings", shortcut: "⌘," },
    ];

    const quickActions = [
        { name: "New Application", icon: Plus, path: "/applications/create" },
        { name: "New Database", icon: Plus, path: "/databases/create" },
        { name: "New Project", icon: Plus, path: "/projects/create" },
    ];

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <CommandInput
                placeholder="Search applications, databases, or type a command..."
                value={search}
                onValueChange={setSearch}
            />
            <CommandList>
                <CommandEmpty>No results found.</CommandEmpty>

                {/* Quick Actions */}
                {!search && (
                    <CommandGroup heading="Quick Actions">
                        {quickActions.map((action) => (
                            <CommandItem
                                key={action.path}
                                onSelect={() => runCommand(() => navigate(action.path))}
                            >
                                <action.icon className="mr-2 h-4 w-4" />
                                {action.name}
                            </CommandItem>
                        ))}
                    </CommandGroup>
                )}

                <CommandSeparator />

                {/* Search Results - Applications */}
                {applications?.data && applications.data.length > 0 && (
                    <CommandGroup heading="Applications">
                        {applications.data.map((app: any) => (
                            <CommandItem
                                key={app.id}
                                onSelect={() =>
                                    runCommand(() => navigate(`/applications/${app.id}`))
                                }
                            >
                                <Box className="mr-2 h-4 w-4" />
                                <span>{app.name}</span>
                                <span className="ml-2 text-xs text-muted-foreground">
                                    {app.status}
                                </span>
                            </CommandItem>
                        ))}
                    </CommandGroup>
                )}

                {/* Search Results - Databases */}
                {databases?.data && databases.data.length > 0 && (
                    <CommandGroup heading="Databases">
                        {databases.data.map((db: any) => (
                            <CommandItem
                                key={db.id}
                                onSelect={() =>
                                    runCommand(() => navigate(`/databases/${db.id}`))
                                }
                            >
                                <Database className="mr-2 h-4 w-4" />
                                <span>{db.name}</span>
                                <span className="ml-2 text-xs text-muted-foreground">
                                    {db.engine}
                                </span>
                            </CommandItem>
                        ))}
                    </CommandGroup>
                )}

                <CommandSeparator />

                {/* Navigation */}
                <CommandGroup heading="Navigation">
                    {navigationItems.map((item) => (
                        <CommandItem
                            key={item.path}
                            onSelect={() => runCommand(() => navigate(item.path))}
                        >
                            <item.icon className="mr-2 h-4 w-4" />
                            {item.name}
                            {item.shortcut && <CommandShortcut>{item.shortcut}</CommandShortcut>}
                        </CommandItem>
                    ))}
                </CommandGroup>
            </CommandList>
        </Dialog>
    );
};
