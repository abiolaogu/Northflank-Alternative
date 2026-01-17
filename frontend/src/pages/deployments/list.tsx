import React from "react";
import {
    MoreHorizontal,
    Search,
    Rocket,
    CheckCircle2,
    XCircle,
    Clock,
    RotateCcw,
    ExternalLink,
    GitBranch,
    RefreshCw
} from "lucide-react";
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { useSimulatedDeploy } from "@/hooks/use-deploy";

// Mock deployment data
const mockDeployments = [
    {
        id: "1",
        application: "frontend-web",
        status: "success",
        revision: 42,
        image: "nginx:1.25",
        commit: "a3f2c1d",
        branch: "main",
        triggeredBy: "Olivia Martin",
        duration: "2m 34s",
        created_at: "2024-01-17T10:30:00Z"
    },
    {
        id: "2",
        application: "api-gateway",
        status: "success",
        revision: 38,
        image: "envoyproxy/envoy:v1.28",
        commit: "b4e3d2f",
        branch: "main",
        triggeredBy: "Jackson Lee",
        duration: "1m 45s",
        created_at: "2024-01-17T09:15:00Z"
    },
    {
        id: "3",
        application: "user-service",
        status: "deploying",
        revision: 55,
        image: "app/user-svc:v2.2.0",
        commit: "c5f4e3g",
        branch: "feature/auth",
        triggeredBy: "System",
        duration: "0m 45s",
        created_at: "2024-01-17T11:00:00Z"
    },
    {
        id: "4",
        application: "payment-service",
        status: "failed",
        revision: 23,
        image: "app/payment-svc:v1.5.3",
        commit: "d6g5h4i",
        branch: "hotfix/payment",
        triggeredBy: "Isabella Nguyen",
        duration: "3m 12s",
        created_at: "2024-01-17T08:45:00Z"
    },
    {
        id: "5",
        application: "notification-worker",
        status: "success",
        revision: 15,
        image: "app/notif-worker:v1.1.0",
        commit: "e7h6i5j",
        branch: "main",
        triggeredBy: "Olivia Martin",
        duration: "1m 20s",
        created_at: "2024-01-16T16:30:00Z"
    },
    {
        id: "6",
        application: "frontend-web",
        status: "rolled_back",
        revision: 41,
        image: "nginx:1.24",
        commit: "f8i7j6k",
        branch: "main",
        triggeredBy: "Jackson Lee",
        duration: "2m 10s",
        created_at: "2024-01-16T14:00:00Z"
    },
];

// Summary stats
const deploymentStats = {
    total: mockDeployments.length,
    success: mockDeployments.filter(d => d.status === 'success').length,
    failed: mockDeployments.filter(d => d.status === 'failed').length,
    deploying: mockDeployments.filter(d => d.status === 'deploying').length,
};

export const DeploymentList = () => {
    const [searchTerm, setSearchTerm] = React.useState("");
    const [statusFilter, setStatusFilter] = React.useState<string>("all");
    const { rollback, isLoading } = useSimulatedDeploy();

    const filteredDeployments = mockDeployments.filter(dep => {
        const matchesSearch = !searchTerm ||
            dep.application.toLowerCase().includes(searchTerm.toLowerCase()) ||
            dep.commit.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === "all" || dep.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const getStatusIcon = (status: string) => {
        switch (status) {
            case 'success': return <CheckCircle2 className="h-4 w-4 text-green-500" />;
            case 'failed': return <XCircle className="h-4 w-4 text-red-500" />;
            case 'deploying': return <RefreshCw className="h-4 w-4 text-blue-500 animate-spin" />;
            case 'rolled_back': return <RotateCcw className="h-4 w-4 text-yellow-500" />;
            default: return <Clock className="h-4 w-4 text-gray-500" />;
        }
    };

    const getStatusBadge = (status: string) => {
        switch (status) {
            case 'success': return <Badge className="bg-green-500/10 text-green-500 border-green-500/20">Success</Badge>;
            case 'failed': return <Badge variant="destructive">Failed</Badge>;
            case 'deploying': return <Badge className="bg-blue-500/10 text-blue-500 border-blue-500/20">Deploying</Badge>;
            case 'rolled_back': return <Badge className="bg-yellow-500/10 text-yellow-500 border-yellow-500/20">Rolled Back</Badge>;
            default: return <Badge variant="secondary">Unknown</Badge>;
        }
    };

    const handleRollback = async (deploymentId: string, revision: number) => {
        await rollback(deploymentId, revision - 1);
    };

    const formatTime = (isoString: string) => {
        const date = new Date(isoString);
        return date.toLocaleString([], {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    return (
        <div className="space-y-6">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Deployments</h2>
                    <p className="text-muted-foreground">
                        Track deployment history and manage rollbacks.
                    </p>
                </div>
            </div>

            {/* Stats Cards */}
            <div className="grid gap-4 md:grid-cols-4">
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Total Deployments</CardDescription>
                        <CardTitle className="text-3xl">{deploymentStats.total}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Successful</CardDescription>
                        <CardTitle className="text-3xl text-green-500">{deploymentStats.success}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Failed</CardDescription>
                        <CardTitle className="text-3xl text-red-500">{deploymentStats.failed}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>In Progress</CardDescription>
                        <CardTitle className="text-3xl text-blue-500">{deploymentStats.deploying}</CardTitle>
                    </CardHeader>
                </Card>
            </div>

            <Card>
                <CardHeader>
                    <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2">
                        <div className="relative flex-1">
                            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search by application or commit..."
                                className="pl-8"
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                            />
                        </div>
                        <Select value={statusFilter} onValueChange={setStatusFilter}>
                            <SelectTrigger className="w-full sm:w-[150px]">
                                <SelectValue placeholder="Status" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">All Statuses</SelectItem>
                                <SelectItem value="success">Success</SelectItem>
                                <SelectItem value="failed">Failed</SelectItem>
                                <SelectItem value="deploying">Deploying</SelectItem>
                                <SelectItem value="rolled_back">Rolled Back</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Application</TableHead>
                                    <TableHead>Status</TableHead>
                                    <TableHead className="hidden md:table-cell">Revision</TableHead>
                                    <TableHead className="hidden lg:table-cell">Commit</TableHead>
                                    <TableHead className="hidden xl:table-cell">Triggered By</TableHead>
                                    <TableHead>Time</TableHead>
                                    <TableHead className="w-[80px]">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filteredDeployments.length > 0 ? (
                                    filteredDeployments.map((deployment) => (
                                        <TableRow key={deployment.id}>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    <Rocket className="h-4 w-4 text-muted-foreground" />
                                                    <div>
                                                        <div className="font-medium">{deployment.application}</div>
                                                        <div className="text-xs text-muted-foreground font-mono hidden sm:block">
                                                            {deployment.image}
                                                        </div>
                                                    </div>
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    {getStatusIcon(deployment.status)}
                                                    {getStatusBadge(deployment.status)}
                                                </div>
                                            </TableCell>
                                            <TableCell className="hidden md:table-cell">
                                                <span className="font-mono">v{deployment.revision}</span>
                                            </TableCell>
                                            <TableCell className="hidden lg:table-cell">
                                                <div className="flex items-center gap-1">
                                                    <GitBranch className="h-3 w-3 text-muted-foreground" />
                                                    <span className="font-mono text-xs">{deployment.commit}</span>
                                                    <span className="text-xs text-muted-foreground">({deployment.branch})</span>
                                                </div>
                                            </TableCell>
                                            <TableCell className="hidden xl:table-cell">
                                                {deployment.triggeredBy}
                                            </TableCell>
                                            <TableCell>
                                                <div className="text-sm">{formatTime(deployment.created_at)}</div>
                                                <div className="text-xs text-muted-foreground">{deployment.duration}</div>
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
                                                        <DropdownMenuItem>
                                                            <ExternalLink className="mr-2 h-4 w-4" />
                                                            View Details
                                                        </DropdownMenuItem>
                                                        <DropdownMenuItem>View Logs</DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        {deployment.status === 'success' && (
                                                            <DropdownMenuItem
                                                                onClick={() => handleRollback(deployment.id, deployment.revision)}
                                                            >
                                                                <RotateCcw className="mr-2 h-4 w-4" />
                                                                Rollback to v{deployment.revision - 1}
                                                            </DropdownMenuItem>
                                                        )}
                                                        {deployment.status === 'failed' && (
                                                            <DropdownMenuItem>
                                                                <RefreshCw className="mr-2 h-4 w-4" />
                                                                Retry Deployment
                                                            </DropdownMenuItem>
                                                        )}
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </TableCell>
                                        </TableRow>
                                    ))
                                ) : (
                                    <TableRow>
                                        <TableCell colSpan={7} className="h-24 text-center">
                                            No deployments found.
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
