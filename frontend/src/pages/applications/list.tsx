import React from "react";
import { MoreHorizontal, Plus, Search, Box, Rocket, RotateCw, Scale } from "lucide-react";
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
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { useNavigation } from "@refinedev/core";
import { useSimulatedDeploy } from "@/hooks/use-deploy";

// Mock data for development
const mockApplications = [
    { id: "1", name: "frontend-web", status: "Running", image: "nginx:1.25", replicas: 3, created_at: "2024-01-15" },
    { id: "2", name: "api-gateway", status: "Running", image: "envoyproxy/envoy:v1.28", replicas: 2, created_at: "2024-01-14" },
    { id: "3", name: "user-service", status: "Running", image: "app/user-svc:v2.1.0", replicas: 3, created_at: "2024-01-12" },
    { id: "4", name: "payment-service", status: "Deploying", image: "app/payment-svc:v1.5.2", replicas: 2, created_at: "2024-01-10" },
    { id: "5", name: "notification-worker", status: "Stopped", image: "app/notif-worker:v1.0.0", replicas: 0, created_at: "2024-01-08" },
];

export const ApplicationList = () => {
    const { create } = useNavigation();
    const [searchTerm, setSearchTerm] = React.useState("");
    const { deploy, restart, scale, isLoading } = useSimulatedDeploy();

    const filteredApps = mockApplications.filter(app =>
        app.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        app.image.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const getStatusVariant = (status: string) => {
        switch (status) {
            case 'Running': return 'default';
            case 'Deploying': return 'secondary';
            default: return 'outline';
        }
    };

    const handleDeploy = async (appId: string, appName: string) => {
        await deploy(appId);
    };

    const handleRestart = async (appId: string) => {
        await restart(appId);
    };

    const handleScale = async (appId: string, replicas: number) => {
        await scale(appId, replicas + 1);
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Applications</h2>
                    <p className="text-muted-foreground">
                        Manage your cloud native applications and microservices.
                    </p>
                </div>
                <Button asChild>
                    <Link to="/applications/create">
                        <Plus className="mr-2 h-4 w-4" />
                        Create Application
                    </Link>
                </Button>
            </div>

            <Card>
                <CardHeader>
                    <div className="flex items-center gap-2">
                        <div className="relative flex-1">
                            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search applications..."
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
                                    <TableHead>Status</TableHead>
                                    <TableHead>Image</TableHead>
                                    <TableHead>Replicas</TableHead>
                                    <TableHead>Created</TableHead>
                                    <TableHead className="w-[80px]">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filteredApps.length > 0 ? (
                                    filteredApps.map((app) => (
                                        <TableRow key={app.id}>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    <Box className="h-4 w-4 text-muted-foreground" />
                                                    <span className="font-medium">{app.name}</span>
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <Badge variant={getStatusVariant(app.status)}>
                                                    {app.status}
                                                </Badge>
                                            </TableCell>
                                            <TableCell className="font-mono text-xs">{app.image}</TableCell>
                                            <TableCell>{app.replicas}</TableCell>
                                            <TableCell className="text-muted-foreground">
                                                {new Date(app.created_at).toLocaleDateString()}
                                            </TableCell>
                                            <TableCell>
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <Button variant="ghost" className="h-8 w-8 p-0" disabled={isLoading}>
                                                            <MoreHorizontal className="h-4 w-4" />
                                                        </Button>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent align="end">
                                                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                        <DropdownMenuItem>View Details</DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuItem onClick={() => handleDeploy(app.id, app.name)}>
                                                            <Rocket className="mr-2 h-4 w-4" />
                                                            Deploy
                                                        </DropdownMenuItem>
                                                        <DropdownMenuItem onClick={() => handleRestart(app.id)}>
                                                            <RotateCw className="mr-2 h-4 w-4" />
                                                            Restart
                                                        </DropdownMenuItem>
                                                        <DropdownMenuItem onClick={() => handleScale(app.id, app.replicas)}>
                                                            <Scale className="mr-2 h-4 w-4" />
                                                            Scale Up
                                                        </DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuItem>Edit Configuration</DropdownMenuItem>
                                                        <DropdownMenuItem>View Logs</DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuItem className="text-red-600">
                                                            Delete Application
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </TableCell>
                                        </TableRow>
                                    ))
                                ) : (
                                    <TableRow>
                                        <TableCell colSpan={6} className="h-24 text-center">
                                            No applications found.
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
