import React from "react";
import { MoreHorizontal, Plus, Search, Key, Eye, EyeOff, Copy, Check } from "lucide-react";
import { Link } from "react-router-dom";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardDescription, CardTitle } from "@/components/ui/card";
import { useNavigation, useList } from "@refinedev/core";

export const SecretList = () => {
    const { create } = useNavigation();
    const [visibleSecrets, setVisibleSecrets] = React.useState<Record<string, boolean>>({});
    const [copiedId, setCopiedId] = React.useState<string | null>(null);

    // Mock secrets data
    const secrets = [
        { id: "1", name: "DATABASE_URL", value: "postgresql://user:pass@host:5432/db", scope: "application", updated_at: "2024-01-15" },
        { id: "2", name: "API_KEY", value: "sk-1234567890abcdef", scope: "global", updated_at: "2024-01-14" },
        { id: "3", name: "JWT_SECRET", value: "super-secret-jwt-key-here", scope: "application", updated_at: "2024-01-13" },
        { id: "4", name: "REDIS_URL", value: "redis://localhost:6379", scope: "environment", updated_at: "2024-01-12" },
    ];

    const toggleVisibility = (id: string) => {
        setVisibleSecrets(prev => ({ ...prev, [id]: !prev[id] }));
    };

    const copyToClipboard = (id: string, value: string) => {
        navigator.clipboard.writeText(value);
        setCopiedId(id);
        setTimeout(() => setCopiedId(null), 2000);
    };

    const maskValue = (value: string) => "•".repeat(Math.min(value.length, 24));

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Secrets</h2>
                    <p className="text-muted-foreground">
                        Manage environment variables and sensitive configuration.
                    </p>
                </div>
                <Button onClick={() => create("secrets")}>
                    <Plus className="mr-2 h-4 w-4" />
                    Add Secret
                </Button>
            </div>

            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle className="text-base">Secret Variables</CardTitle>
                            <CardDescription>
                                Encrypted at rest and in transit. Never logged or exposed.
                            </CardDescription>
                        </div>
                        <div className="relative w-64">
                            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input placeholder="Search secrets..." className="pl-8" />
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Value</TableHead>
                                    <TableHead>Scope</TableHead>
                                    <TableHead>Last Updated</TableHead>
                                    <TableHead className="w-[100px]">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {secrets.map((secret) => (
                                    <TableRow key={secret.id}>
                                        <TableCell>
                                            <div className="flex items-center gap-2">
                                                <Key className="h-4 w-4 text-muted-foreground" />
                                                <span className="font-mono font-medium">{secret.name}</span>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex items-center gap-2">
                                                <code className="font-mono text-xs bg-muted px-2 py-1 rounded">
                                                    {visibleSecrets[secret.id] ? secret.value : maskValue(secret.value)}
                                                </code>
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    className="h-6 w-6"
                                                    onClick={() => toggleVisibility(secret.id)}
                                                >
                                                    {visibleSecrets[secret.id] ? (
                                                        <EyeOff className="h-3 w-3" />
                                                    ) : (
                                                        <Eye className="h-3 w-3" />
                                                    )}
                                                </Button>
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    className="h-6 w-6"
                                                    onClick={() => copyToClipboard(secret.id, secret.value)}
                                                >
                                                    {copiedId === secret.id ? (
                                                        <Check className="h-3 w-3 text-green-500" />
                                                    ) : (
                                                        <Copy className="h-3 w-3" />
                                                    )}
                                                </Button>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant="outline">{secret.scope}</Badge>
                                        </TableCell>
                                        <TableCell className="text-muted-foreground">
                                            {new Date(secret.updated_at).toLocaleDateString()}
                                        </TableCell>
                                        <TableCell>
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button variant="ghost" className="h-8 w-8 p-0">
                                                        <MoreHorizontal className="h-4 w-4" />
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                    <DropdownMenuItem>Edit Value</DropdownMenuItem>
                                                    <DropdownMenuItem>View History</DropdownMenuItem>
                                                    <DropdownMenuSeparator />
                                                    <DropdownMenuItem className="text-red-600">
                                                        Delete Secret
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
};
