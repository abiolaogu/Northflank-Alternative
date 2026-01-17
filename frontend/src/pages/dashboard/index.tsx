import { useList, useNavigation } from "@refinedev/core";
import {
    Activity,
    ArrowUpRight,
    Box,
    Database,
    GitBranch,
    MoreHorizontal,
    Users
} from "lucide-react";
import { Link } from "react-router-dom";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
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
import {
    Area,
    AreaChart,
    Bar,
    BarChart,
    CartesianGrid,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis
} from "recharts";

// Mock data for charts
const resourceData = [
    { name: "00:00", cpu: 45, memory: 30 },
    { name: "04:00", cpu: 55, memory: 40 },
    { name: "08:00", cpu: 75, memory: 60 },
    { name: "12:00", cpu: 65, memory: 50 },
    { name: "16:00", cpu: 85, memory: 70 },
    { name: "20:00", cpu: 60, memory: 45 },
    { name: "23:59", cpu: 50, memory: 35 },
];

const deploymentData = [
    { name: "Mon", total: 12, success: 10, failed: 2 },
    { name: "Tue", total: 18, success: 17, failed: 1 },
    { name: "Wed", total: 15, success: 15, failed: 0 },
    { name: "Thu", total: 25, success: 23, failed: 2 },
    { name: "Fri", total: 20, success: 19, failed: 1 },
    { name: "Sat", total: 8, success: 8, failed: 0 },
    { name: "Sun", total: 5, success: 5, failed: 0 },
];

export const Dashboard = () => {
    const { show } = useNavigation();

    // Fetch real data (mocked by provider for now)
    const { data: applications } = useList({
        resource: "applications",
        pagination: { pageSize: 5 },
        sorters: [{ field: "created_at", order: "desc" }],
    }) as any;

    const { data: stats } = useList({
        resource: "stats", // Custom resource for dashboard stats
        queryOptions: {
            enabled: false // Disable auto fetch as we are mocking for now
        }
    }) as any;

    // Mock stats
    const dashboardStats = [
        {
            title: "Total Applications",
            value: "24",
            change: "+12%",
            trend: "up",
            icon: Box,
        },
        {
            title: "Active Databases",
            value: "12",
            change: "+4%",
            trend: "up",
            icon: Database,
        },
        {
            title: "Active Users",
            value: "573",
            change: "+201 since last hour",
            trend: "up",
            icon: Users,
        },
        {
            title: "Success Rate",
            value: "98.5%",
            change: "-0.5%",
            trend: "down",
            icon: Activity,
        },
    ];

    return (
        <div className="flex flex-col gap-6">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
                <div className="flex items-center gap-2">
                    <Button variant="outline" asChild>
                        <Link to="/logs">View Logs</Link>
                    </Button>
                    <Button asChild>
                        <Link to="/applications/create">New Application</Link>
                    </Button>
                </div>
            </div>

            {/* Stats Grid */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {dashboardStats.map((stat) => (
                    <Card key={stat.title}>
                        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                            <CardTitle className="text-sm font-medium">
                                {stat.title}
                            </CardTitle>
                            <stat.icon className="h-4 w-4 text-muted-foreground" />
                        </CardHeader>
                        <CardContent>
                            <div className="text-2xl font-bold">{stat.value}</div>
                            <p className="text-xs text-muted-foreground">
                                {stat.change}
                            </p>
                        </CardContent>
                    </Card>
                ))}
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
                {/* Charts - Resource Usage */}
                <Card className="col-span-4">
                    <CardHeader>
                        <CardTitle>Resource Usage Overview</CardTitle>
                        <CardDescription>
                            CPU and Memory usage across all clusters for today.
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="pl-2">
                        <ResponsiveContainer width="100%" height={350}>
                            <AreaChart data={resourceData}>
                                <defs>
                                    <linearGradient id="colorCpu" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8} />
                                        <stop offset="95%" stopColor="#8884d8" stopOpacity={0} />
                                    </linearGradient>
                                    <linearGradient id="colorMem" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#82ca9d" stopOpacity={0.8} />
                                        <stop offset="95%" stopColor="#82ca9d" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <XAxis dataKey="name" stroke="#888888" fontSize={12} tickLine={false} axisLine={false} />
                                <YAxis stroke="#888888" fontSize={12} tickLine={false} axisLine={false} tickFormatter={(value) => `${value}%`} />
                                <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                <Tooltip />
                                <Area type="monotone" dataKey="cpu" stroke="#8884d8" fillOpacity={1} fill="url(#colorCpu)" />
                                <Area type="monotone" dataKey="memory" stroke="#82ca9d" fillOpacity={1} fill="url(#colorMem)" />
                            </AreaChart>
                        </ResponsiveContainer>
                    </CardContent>
                </Card>

                {/* Recent Deployments Activity */}
                <Card className="col-span-3">
                    <CardHeader>
                        <CardTitle>Recent Activity</CardTitle>
                        <CardDescription>
                            Recent deployments and system events.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-8">
                            <div className="flex items-center">
                                <Avatar className="h-9 w-9">
                                    <AvatarImage src="/avatars/01.png" alt="Avatar" />
                                    <AvatarFallback>OM</AvatarFallback>
                                </Avatar>
                                <div className="ml-4 space-y-1">
                                    <p className="text-sm font-medium leading-none">Olivia Martin</p>
                                    <p className="text-sm text-muted-foreground">
                                        Deployed frontend-app to production
                                    </p>
                                </div>
                                <div className="ml-auto font-medium text-green-500">Success</div>
                            </div>
                            <div className="flex items-center">
                                <Avatar className="h-9 w-9">
                                    <AvatarImage src="/avatars/02.png" alt="Avatar" />
                                    <AvatarFallback>JL</AvatarFallback>
                                </Avatar>
                                <div className="ml-4 space-y-1">
                                    <p className="text-sm font-medium leading-none">Jackson Lee</p>
                                    <p className="text-sm text-muted-foreground">
                                        Scaled database-main to 3 replicas
                                    </p>
                                </div>
                                <div className="ml-auto font-medium text-blue-500">Scaled</div>
                            </div>
                            <div className="flex items-center">
                                <Avatar className="h-9 w-9">
                                    <AvatarImage src="/avatars/03.png" alt="Avatar" />
                                    <AvatarFallback>IN</AvatarFallback>
                                </Avatar>
                                <div className="ml-4 space-y-1">
                                    <p className="text-sm font-medium leading-none">Isabella Nguyen</p>
                                    <p className="text-sm text-muted-foreground">
                                        Failed deployment for backend-api
                                    </p>
                                </div>
                                <div className="ml-auto font-medium text-red-500">Failed</div>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Recent Applications Table */}
            <Card>
                <CardHeader>
                    <CardTitle>Recent Applications</CardTitle>
                    <CardDescription>A list of recently created or updated applications.</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Name</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Region</TableHead>
                                <TableHead>Created At</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {applications?.data?.map((app: any) => (
                                <TableRow key={app.id}>
                                    <TableCell className="font-medium">
                                        <div className="flex items-center gap-2">
                                            <Box className="h-4 w-4 text-muted-foreground" />
                                            {app.name}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant={app.status === 'Running' ? 'default' : 'secondary'}>
                                            {app.status}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>{app.region || 'us-east-1'}</TableCell>
                                    <TableCell>{new Date(app.created_at).toLocaleDateString()}</TableCell>
                                    <TableCell className="text-right">
                                        <DropdownMenu>
                                            <DropdownMenuTrigger asChild>
                                                <Button variant="ghost" className="h-8 w-8 p-0">
                                                    <MoreHorizontal className="h-4 w-4" />
                                                </Button>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent align="end">
                                                <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                <DropdownMenuItem onClick={() => show('applications', app.id)}>
                                                    View Details
                                                </DropdownMenuItem>
                                                <DropdownMenuSeparator />
                                                <DropdownMenuItem>View Logs</DropdownMenuItem>
                                                <DropdownMenuItem className="text-red-600">Delete</DropdownMenuItem>
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                    </TableCell>
                                </TableRow>
                            ))}
                            {!applications?.data && (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center py-4 text-muted-foreground">
                                        No applications found.
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
};
