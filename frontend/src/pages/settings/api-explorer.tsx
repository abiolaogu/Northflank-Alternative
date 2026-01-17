import React, { useState } from "react";
import {
    Search,
    Shield,
    ShieldCheck,
    ShieldX,
    Database,
    Code,
    Zap,
    Globe,
    Lock,
    Unlock,
    ChevronDown,
    ChevronRight,
    Copy,
    Check,
    Eye,
    EyeOff,
    Filter,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";

// Hasura API Catalog based on metadata.yaml
const hasuraAPIs = {
    tables: [
        {
            name: "projects",
            schema: "public",
            description: "User projects and workspaces",
            relationships: { object: ["owner"], array: ["services"] },
            permissions: {
                select: ["user", "admin"],
                insert: ["user", "admin"],
                update: ["user", "admin"],
                delete: ["admin"],
            },
            columns: ["id", "name", "slug", "description", "labels", "owner_id", "status", "created_at", "updated_at"],
            publicExposed: true,
        },
        {
            name: "services",
            schema: "public",
            description: "Application services and containers",
            relationships: { object: ["project"], array: ["builds", "deployments"] },
            permissions: {
                select: ["user", "admin"],
                insert: ["admin"],
                update: ["admin"],
                delete: ["admin"],
            },
            columns: ["id", "name", "type", "status", "project_id", "image", "replicas", "cpu", "memory", "created_at"],
            publicExposed: true,
        },
        {
            name: "builds",
            schema: "public",
            description: "CI/CD build records",
            relationships: { object: ["service"], array: [] },
            permissions: {
                select: ["user", "admin"],
                insert: ["admin"],
                update: ["admin"],
                delete: ["admin"],
            },
            columns: ["id", "service_id", "status", "commit_sha", "branch", "started_at", "completed_at", "logs"],
            publicExposed: true,
        },
        {
            name: "deployments",
            schema: "public",
            description: "Deployment history and status",
            relationships: { object: ["service", "build"], array: [] },
            permissions: {
                select: ["user", "admin"],
                insert: ["admin"],
                update: ["admin"],
                delete: ["admin"],
            },
            columns: ["id", "service_id", "build_id", "status", "revision", "replicas", "created_at", "completed_at"],
            publicExposed: true,
        },
        {
            name: "clusters",
            schema: "public",
            description: "Kubernetes cluster management",
            relationships: { object: [], array: [] },
            permissions: {
                select: ["admin"],
                insert: ["admin"],
                update: ["admin"],
                delete: ["admin"],
            },
            columns: ["id", "name", "provider", "region", "status", "version", "node_count", "created_at"],
            publicExposed: false,
        },
        {
            name: "databases",
            schema: "public",
            description: "Managed database instances",
            relationships: { object: ["project"], array: [] },
            permissions: {
                select: ["user", "admin"],
                insert: ["admin"],
                update: ["admin"],
                delete: ["admin"],
            },
            columns: ["id", "name", "type", "version", "status", "size", "project_id", "connection_string", "created_at"],
            publicExposed: true,
        },
        {
            name: "users",
            schema: "public",
            description: "User accounts and profiles",
            relationships: { object: [], array: [] },
            permissions: {
                select: ["user", "admin"],
                insert: ["admin"],
                update: ["admin"],
                delete: ["admin"],
            },
            columns: ["id", "email", "name", "avatar_url", "role", "created_at", "last_login"],
            publicExposed: false,
        },
        {
            name: "audit_logs",
            schema: "public",
            description: "System audit trail",
            relationships: { object: [], array: [] },
            permissions: {
                select: ["admin"],
                insert: ["admin"],
                update: [],
                delete: [],
            },
            columns: ["id", "user_id", "action", "resource", "resource_id", "metadata", "ip_address", "created_at"],
            publicExposed: false,
        },
    ],
    actions: [
        {
            name: "triggerBuild",
            type: "mutation",
            description: "Trigger a new CI/CD build for a service",
            arguments: [
                { name: "serviceId", type: "uuid!", required: true },
                { name: "branch", type: "String", required: false },
            ],
            returnType: "Build",
            permissions: ["user", "admin"],
            publicExposed: true,
        },
        {
            name: "scaleService",
            type: "mutation",
            description: "Scale service replica count",
            arguments: [
                { name: "serviceId", type: "uuid!", required: true },
                { name: "replicas", type: "Int!", required: true },
            ],
            returnType: "Service",
            permissions: ["user", "admin"],
            publicExposed: true,
        },
        {
            name: "createDatabase",
            type: "mutation",
            description: "Provision a new managed database",
            arguments: [
                { name: "projectId", type: "uuid!", required: true },
                { name: "name", type: "String!", required: true },
                { name: "size", type: "String!", required: true },
                { name: "highAvailability", type: "Boolean", required: false },
            ],
            returnType: "Database",
            permissions: ["admin"],
            publicExposed: false,
        },
    ],
    subscriptions: [
        {
            name: "projects",
            description: "Real-time project updates",
            permissions: ["user", "admin"],
            publicExposed: true,
        },
        {
            name: "services",
            description: "Real-time service status updates",
            permissions: ["user", "admin"],
            publicExposed: true,
        },
        {
            name: "builds",
            description: "Real-time build progress",
            permissions: ["user", "admin"],
            publicExposed: true,
        },
        {
            name: "deployments",
            description: "Real-time deployment status",
            permissions: ["user", "admin"],
            publicExposed: true,
        },
    ],
    restEndpoints: [
        {
            name: "GetProjects",
            method: "GET",
            url: "/api/rest/projects",
            description: "Fetch user projects via REST",
            permissions: ["user", "admin"],
            publicExposed: true,
        },
    ],
    eventTriggers: [
        {
            name: "service_status_changed",
            table: "services",
            events: ["INSERT", "UPDATE"],
            description: "Webhook when service status changes",
            publicExposed: false,
        },
        {
            name: "build_completed",
            table: "builds",
            events: ["UPDATE"],
            description: "Webhook when build completes",
            publicExposed: false,
        },
        {
            name: "deployment_completed",
            table: "deployments",
            events: ["UPDATE"],
            description: "Webhook when deployment completes",
            publicExposed: false,
        },
    ],
};

// Summary stats
const apiStats = {
    totalQueries: hasuraAPIs.tables.length,
    totalMutations: hasuraAPIs.tables.length + hasuraAPIs.actions.length,
    totalSubscriptions: hasuraAPIs.subscriptions.length,
    publicAPIs: hasuraAPIs.tables.filter(t => t.publicExposed).length +
        hasuraAPIs.actions.filter(a => a.publicExposed).length,
    adminOnly: hasuraAPIs.tables.filter(t => !t.publicExposed).length +
        hasuraAPIs.actions.filter(a => !a.publicExposed).length,
};

export const APIExplorerPage = () => {
    const [searchTerm, setSearchTerm] = useState("");
    const [roleFilter, setRoleFilter] = useState<string>("all");
    const [expandedTables, setExpandedTables] = useState<Set<string>>(new Set());
    const [copiedItem, setCopiedItem] = useState<string | null>(null);

    const toggleTable = (tableName: string) => {
        setExpandedTables(prev => {
            const next = new Set(prev);
            if (next.has(tableName)) {
                next.delete(tableName);
            } else {
                next.add(tableName);
            }
            return next;
        });
    };

    const copyToClipboard = (text: string, id: string) => {
        navigator.clipboard.writeText(text);
        setCopiedItem(id);
        setTimeout(() => setCopiedItem(null), 2000);
    };

    const filteredTables = hasuraAPIs.tables.filter(table => {
        const matchesSearch = !searchTerm ||
            table.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            table.description.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesRole = roleFilter === "all" ||
            (roleFilter === "public" && table.publicExposed) ||
            (roleFilter === "admin" && !table.publicExposed);
        return matchesSearch && matchesRole;
    });

    const filteredActions = hasuraAPIs.actions.filter(action => {
        const matchesSearch = !searchTerm ||
            action.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            action.description.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesRole = roleFilter === "all" ||
            (roleFilter === "public" && action.publicExposed) ||
            (roleFilter === "admin" && !action.publicExposed);
        return matchesSearch && matchesRole;
    });

    const getPermissionBadge = (permissions: string[]) => {
        if (permissions.includes("user")) {
            return <Badge className="bg-green-500/10 text-green-500 border-green-500/20 gap-1"><Unlock className="h-3 w-3" /> User</Badge>;
        }
        if (permissions.includes("admin")) {
            return <Badge className="bg-orange-500/10 text-orange-500 border-orange-500/20 gap-1"><Lock className="h-3 w-3" /> Admin</Badge>;
        }
        return <Badge variant="secondary">None</Badge>;
    };

    return (
        <div className="space-y-6">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Hasura API Explorer</h2>
                    <p className="text-muted-foreground">
                        Review and configure API exposure settings.
                    </p>
                </div>
                <Button>
                    <ShieldCheck className="mr-2 h-4 w-4" />
                    Save Permissions
                </Button>
            </div>

            {/* Summary Cards */}
            <div className="grid gap-4 md:grid-cols-5">
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Queries</CardDescription>
                        <CardTitle className="text-2xl">{apiStats.totalQueries}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Mutations</CardDescription>
                        <CardTitle className="text-2xl">{apiStats.totalMutations}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Subscriptions</CardDescription>
                        <CardTitle className="text-2xl">{apiStats.totalSubscriptions}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Public APIs</CardDescription>
                        <CardTitle className="text-2xl text-green-500">{apiStats.publicAPIs}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Admin Only</CardDescription>
                        <CardTitle className="text-2xl text-orange-500">{apiStats.adminOnly}</CardTitle>
                    </CardHeader>
                </Card>
            </div>

            {/* Filters */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2">
                        <div className="relative flex-1">
                            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search APIs..."
                                className="pl-8"
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                            />
                        </div>
                        <Select value={roleFilter} onValueChange={setRoleFilter}>
                            <SelectTrigger className="w-full sm:w-[150px]">
                                <Filter className="mr-2 h-4 w-4" />
                                <SelectValue placeholder="Filter" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">All APIs</SelectItem>
                                <SelectItem value="public">Public Only</SelectItem>
                                <SelectItem value="admin">Admin Only</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </CardHeader>
            </Card>

            {/* API Tabs */}
            <Tabs defaultValue="tables" className="space-y-4">
                <TabsList className="grid w-full grid-cols-4 lg:w-auto lg:inline-flex">
                    <TabsTrigger value="tables" className="gap-2">
                        <Database className="h-4 w-4" />
                        Tables ({hasuraAPIs.tables.length})
                    </TabsTrigger>
                    <TabsTrigger value="actions" className="gap-2">
                        <Zap className="h-4 w-4" />
                        Actions ({hasuraAPIs.actions.length})
                    </TabsTrigger>
                    <TabsTrigger value="subscriptions" className="gap-2">
                        <Globe className="h-4 w-4" />
                        Subscriptions ({hasuraAPIs.subscriptions.length})
                    </TabsTrigger>
                    <TabsTrigger value="events" className="gap-2">
                        <Code className="h-4 w-4" />
                        Events ({hasuraAPIs.eventTriggers.length})
                    </TabsTrigger>
                </TabsList>

                {/* Tables Tab */}
                <TabsContent value="tables" className="space-y-4">
                    {filteredTables.map((table) => (
                        <Collapsible
                            key={table.name}
                            open={expandedTables.has(table.name)}
                            onOpenChange={() => toggleTable(table.name)}
                        >
                            <Card>
                                <CollapsibleTrigger asChild>
                                    <CardHeader className="cursor-pointer hover:bg-muted/50 transition-colors">
                                        <div className="flex items-center justify-between">
                                            <div className="flex items-center gap-3">
                                                {expandedTables.has(table.name) ?
                                                    <ChevronDown className="h-4 w-4" /> :
                                                    <ChevronRight className="h-4 w-4" />
                                                }
                                                <Database className="h-5 w-5 text-primary" />
                                                <div>
                                                    <CardTitle className="text-lg">{table.name}</CardTitle>
                                                    <CardDescription>{table.description}</CardDescription>
                                                </div>
                                            </div>
                                            <div className="flex items-center gap-3">
                                                {table.publicExposed ? (
                                                    <Badge className="bg-green-500/10 text-green-500 border-green-500/20">
                                                        <Eye className="mr-1 h-3 w-3" /> Public
                                                    </Badge>
                                                ) : (
                                                    <Badge variant="secondary">
                                                        <EyeOff className="mr-1 h-3 w-3" /> Internal
                                                    </Badge>
                                                )}
                                                <Switch checked={table.publicExposed} />
                                            </div>
                                        </div>
                                    </CardHeader>
                                </CollapsibleTrigger>
                                <CollapsibleContent>
                                    <CardContent className="pt-0">
                                        <div className="space-y-4">
                                            {/* Permissions */}
                                            <div>
                                                <h4 className="text-sm font-medium mb-2">Permissions by Operation</h4>
                                                <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
                                                    <div className="p-3 rounded-lg border bg-muted/30">
                                                        <div className="text-xs text-muted-foreground mb-1">SELECT</div>
                                                        <div className="flex flex-wrap gap-1">
                                                            {table.permissions.select.map(role => (
                                                                <Badge key={role} variant="outline" className="text-xs">{role}</Badge>
                                                            ))}
                                                        </div>
                                                    </div>
                                                    <div className="p-3 rounded-lg border bg-muted/30">
                                                        <div className="text-xs text-muted-foreground mb-1">INSERT</div>
                                                        <div className="flex flex-wrap gap-1">
                                                            {table.permissions.insert.map(role => (
                                                                <Badge key={role} variant="outline" className="text-xs">{role}</Badge>
                                                            ))}
                                                        </div>
                                                    </div>
                                                    <div className="p-3 rounded-lg border bg-muted/30">
                                                        <div className="text-xs text-muted-foreground mb-1">UPDATE</div>
                                                        <div className="flex flex-wrap gap-1">
                                                            {table.permissions.update.map(role => (
                                                                <Badge key={role} variant="outline" className="text-xs">{role}</Badge>
                                                            ))}
                                                        </div>
                                                    </div>
                                                    <div className="p-3 rounded-lg border bg-muted/30">
                                                        <div className="text-xs text-muted-foreground mb-1">DELETE</div>
                                                        <div className="flex flex-wrap gap-1">
                                                            {table.permissions.delete.length > 0 ?
                                                                table.permissions.delete.map(role => (
                                                                    <Badge key={role} variant="outline" className="text-xs">{role}</Badge>
                                                                )) :
                                                                <span className="text-xs text-muted-foreground">None</span>
                                                            }
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>

                                            {/* Columns */}
                                            <div>
                                                <h4 className="text-sm font-medium mb-2">Columns ({table.columns.length})</h4>
                                                <div className="flex flex-wrap gap-1">
                                                    {table.columns.map(col => (
                                                        <Badge key={col} variant="secondary" className="font-mono text-xs">
                                                            {col}
                                                        </Badge>
                                                    ))}
                                                </div>
                                            </div>

                                            {/* Example Query */}
                                            <div>
                                                <h4 className="text-sm font-medium mb-2">Example Query</h4>
                                                <div className="relative">
                                                    <pre className="bg-zinc-950 text-zinc-300 p-4 rounded-lg text-xs overflow-x-auto font-mono">
                                                        {`query Get${table.name.charAt(0).toUpperCase() + table.name.slice(1)} {
  ${table.name} {
    ${table.columns.slice(0, 5).join('\n    ')}
  }
}`}
                                                    </pre>
                                                    <Button
                                                        size="icon"
                                                        variant="ghost"
                                                        className="absolute top-2 right-2 h-6 w-6"
                                                        onClick={() => copyToClipboard(`query { ${table.name} { id } }`, table.name)}
                                                    >
                                                        {copiedItem === table.name ?
                                                            <Check className="h-3 w-3 text-green-500" /> :
                                                            <Copy className="h-3 w-3" />
                                                        }
                                                    </Button>
                                                </div>
                                            </div>
                                        </div>
                                    </CardContent>
                                </CollapsibleContent>
                            </Card>
                        </Collapsible>
                    ))}
                </TabsContent>

                {/* Actions Tab */}
                <TabsContent value="actions" className="space-y-4">
                    <Card>
                        <CardContent className="pt-6">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Action Name</TableHead>
                                        <TableHead>Type</TableHead>
                                        <TableHead className="hidden md:table-cell">Arguments</TableHead>
                                        <TableHead>Permissions</TableHead>
                                        <TableHead>Public</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {filteredActions.map((action) => (
                                        <TableRow key={action.name}>
                                            <TableCell>
                                                <div>
                                                    <div className="font-medium font-mono">{action.name}</div>
                                                    <div className="text-xs text-muted-foreground">{action.description}</div>
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <Badge variant="outline">{action.type}</Badge>
                                            </TableCell>
                                            <TableCell className="hidden md:table-cell">
                                                <div className="flex flex-wrap gap-1">
                                                    {action.arguments.map(arg => (
                                                        <Badge key={arg.name} variant="secondary" className="font-mono text-xs">
                                                            {arg.name}: {arg.type}
                                                        </Badge>
                                                    ))}
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                {getPermissionBadge(action.permissions)}
                                            </TableCell>
                                            <TableCell>
                                                <Switch checked={action.publicExposed} />
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* Subscriptions Tab */}
                <TabsContent value="subscriptions" className="space-y-4">
                    <Card>
                        <CardContent className="pt-6">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Subscription</TableHead>
                                        <TableHead>Description</TableHead>
                                        <TableHead>Permissions</TableHead>
                                        <TableHead>Public</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {hasuraAPIs.subscriptions.map((sub) => (
                                        <TableRow key={sub.name}>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    <Globe className="h-4 w-4 text-primary" />
                                                    <span className="font-mono">{sub.name}</span>
                                                </div>
                                            </TableCell>
                                            <TableCell className="text-muted-foreground">{sub.description}</TableCell>
                                            <TableCell>
                                                {getPermissionBadge(sub.permissions)}
                                            </TableCell>
                                            <TableCell>
                                                <Switch checked={sub.publicExposed} />
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </CardContent>
                    </Card>
                </TabsContent>

                {/* Event Triggers Tab */}
                <TabsContent value="events" className="space-y-4">
                    <Card>
                        <CardContent className="pt-6">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Trigger Name</TableHead>
                                        <TableHead>Table</TableHead>
                                        <TableHead>Events</TableHead>
                                        <TableHead>Description</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {hasuraAPIs.eventTriggers.map((trigger) => (
                                        <TableRow key={trigger.name}>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    <Zap className="h-4 w-4 text-yellow-500" />
                                                    <span className="font-mono">{trigger.name}</span>
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <Badge variant="outline">{trigger.table}</Badge>
                                            </TableCell>
                                            <TableCell>
                                                <div className="flex gap-1">
                                                    {trigger.events.map(e => (
                                                        <Badge key={e} variant="secondary" className="text-xs">{e}</Badge>
                                                    ))}
                                                </div>
                                            </TableCell>
                                            <TableCell className="text-muted-foreground">{trigger.description}</TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
};
