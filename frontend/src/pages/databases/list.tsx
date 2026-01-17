import React from "react";
import { MoreHorizontal, Plus, Search, Database, Rocket } from "lucide-react";
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useNavigation } from "@refinedev/core";

// Mock data for development
const mockDatabases = [
    { id: "1", name: "production-db", type: "PostgreSQL", version: "16", status: "Running", storage: "50 GB", created_at: "2024-01-10" },
    { id: "2", name: "staging-db", type: "PostgreSQL", version: "15", status: "Running", storage: "25 GB", created_at: "2024-01-08" },
    { id: "3", name: "cache-redis", type: "Redis", version: "7.2", status: "Running", storage: "10 GB", created_at: "2024-01-05" },
    { id: "4", name: "analytics-ch", type: "ClickHouse", version: "24.1", status: "Stopped", storage: "100 GB", created_at: "2024-01-02" },
];

export const DatabaseList = () => {
    const { edit, show, create } = useNavigation();
    const [searchTerm, setSearchTerm] = React.useState("");

    const filteredDatabases = mockDatabases.filter(db =>
        db.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        db.type.toLowerCase().includes(searchTerm.toLowerCase())
    );

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Databases</h2>
                    <p className="text-muted-foreground">
                        Manage your managed database instances.
                    </p>
                </div>
                <Button asChild>
                    <Link to="/databases/create">
                        <Plus className="mr-2 h-4 w-4" />
                        Create Database
                    </Link>
                </Button>
            </div>

            <Card>
                <CardHeader>
                    <div className="flex items-center gap-2">
                        <div className="relative flex-1">
                            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search databases..."
                                className="pl-8"
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                            />
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Type</TableHead>
                                    <TableHead>Version</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead>Storage</TableHead>
                                    <TableHead>Created</TableHead>
                                    <TableHead className="w-[80px]">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filteredDatabases.length > 0 ? (
                                    filteredDatabases.map((db) => (
                                        <TableRow key={db.id}>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    <Database className="h-4 w-4 text-muted-foreground" />
                                                    <span className="font-medium">{db.name}</span>
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <Badge variant="outline">{db.type}</Badge>
                                            </TableCell>
                                            <TableCell className="font-mono text-xs">{db.version}</TableCell>
                                            <TableCell>
                                                <Badge variant={db.status === 'Running' ? 'default' : 'secondary'}>
                                                    {db.status}
                                                </Badge>
                                            </TableCell>
                                            <TableCell>{db.storage}</TableCell>
                                            <TableCell className="text-muted-foreground">
                                                {new Date(db.created_at).toLocaleDateString()}
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
                                                        <DropdownMenuItem>View Details</DropdownMenuItem>
                                                        <DropdownMenuItem>
                                                            <Rocket className="mr-2 h-4 w-4" />
                                                            Deploy Changes
                                                        </DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuItem>Edit Configuration</DropdownMenuItem>
                                                        <DropdownMenuItem className="text-red-600">
                                                            Delete Database
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </TableCell>
                                        </TableRow>
                                    ))
                                ) : (
                                    <TableRow>
                                        <TableCell colSpan={7} className="h-24 text-center">
                                            No databases found.
                                        </TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
};
