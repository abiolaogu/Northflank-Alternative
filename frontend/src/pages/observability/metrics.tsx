import React, { useState, useEffect } from "react";
import {
    Area,
    AreaChart,
    CartesianGrid,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
    Line,
    LineChart,
} from "recharts";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Wifi, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useSimulatedMetricsSubscription } from "@/hooks/use-hasura-subscription";

// Generate time-series data for charts
function generateMetricsData(count = 12) {
    const now = Date.now();
    return Array.from({ length: count }, (_, i) => {
        const time = new Date(now - (count - 1 - i) * 5 * 60000);
        return {
            time: time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
            cpu: Math.round(30 + Math.random() * 40),
            memory: Math.round(40 + Math.random() * 30),
            networkIn: Math.round(Math.random() * 50),
            networkOut: Math.round(Math.random() * 30),
        };
    });
}

export const MetricsPage = () => {
    const [timeRange, setTimeRange] = useState("1h");
    const [metricsData, setMetricsData] = useState(generateMetricsData(12));
    const { metrics: streamMetrics } = useSimulatedMetricsSubscription();

    // Update metrics every 30 seconds
    useEffect(() => {
        const interval = setInterval(() => {
            setMetricsData(prev => {
                const newPoint = {
                    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
                    cpu: Math.round(30 + Math.random() * 40),
                    memory: Math.round(40 + Math.random() * 30),
                    networkIn: Math.round(Math.random() * 50),
                    networkOut: Math.round(Math.random() * 30),
                };
                return [...prev.slice(1), newPoint];
            });
        }, 30000);
        return () => clearInterval(interval);
    }, []);

    // Summary stats
    const currentCpu = metricsData[metricsData.length - 1]?.cpu || 0;
    const currentMemory = metricsData[metricsData.length - 1]?.memory || 0;
    const avgCpu = Math.round(metricsData.reduce((acc, d) => acc + d.cpu, 0) / metricsData.length);
    const avgMemory = Math.round(metricsData.reduce((acc, d) => acc + d.memory, 0) / metricsData.length);

    return (
        <div className="space-y-6">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">System Metrics</h2>
                    <p className="text-muted-foreground">
                        Monitor the performance and health of your infrastructure.
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    <Badge variant="default" className="gap-1">
                        <Wifi className="h-3 w-3" />
                        Live
                    </Badge>
                    <Select value={timeRange} onValueChange={setTimeRange}>
                        <SelectTrigger className="w-[120px]">
                            <SelectValue placeholder="Range" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="1h">Last 1h</SelectItem>
                            <SelectItem value="6h">Last 6h</SelectItem>
                            <SelectItem value="24h">Last 24h</SelectItem>
                        </SelectContent>
                    </Select>
                    <Button variant="outline" size="icon" onClick={() => setMetricsData(generateMetricsData(12))}>
                        <RefreshCw className="h-4 w-4" />
                    </Button>
                </div>
            </div>

            {/* Summary Cards */}
            <div className="grid gap-4 md:grid-cols-4">
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Current CPU</CardDescription>
                        <CardTitle className="text-3xl">{currentCpu}%</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-xs text-muted-foreground">Avg: {avgCpu}%</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Current Memory</CardDescription>
                        <CardTitle className="text-3xl">{currentMemory}%</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-xs text-muted-foreground">Avg: {avgMemory}%</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Network In</CardDescription>
                        <CardTitle className="text-3xl">{metricsData[metricsData.length - 1]?.networkIn || 0} MB/s</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-xs text-muted-foreground">Inbound traffic</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="pb-2">
                        <CardDescription>Network Out</CardDescription>
                        <CardTitle className="text-3xl">{metricsData[metricsData.length - 1]?.networkOut || 0} MB/s</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-xs text-muted-foreground">Outbound traffic</p>
                    </CardContent>
                </Card>
            </div>

            <Tabs defaultValue="overview" className="space-y-4">
                <TabsList>
                    <TabsTrigger value="overview">Overview</TabsTrigger>
                    <TabsTrigger value="nodes">Nodes</TabsTrigger>
                    <TabsTrigger value="pods">Pods</TabsTrigger>
                </TabsList>
                <TabsContent value="overview" className="space-y-4">
                    <div className="grid gap-4 md:grid-cols-2">
                        <Card>
                            <CardHeader>
                                <CardTitle>CPU Usage</CardTitle>
                                <CardDescription>Aggregate CPU utilization across all nodes.</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="h-[300px]">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <AreaChart data={metricsData}>
                                            <defs>
                                                <linearGradient id="colorCpu" x1="0" y1="0" x2="0" y2="1">
                                                    <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8} />
                                                    <stop offset="95%" stopColor="#8884d8" stopOpacity={0} />
                                                </linearGradient>
                                            </defs>
                                            <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                            <XAxis dataKey="time" stroke="#888888" fontSize={12} tickLine={false} axisLine={false} />
                                            <YAxis stroke="#888888" fontSize={12} tickLine={false} axisLine={false} unit="%" domain={[0, 100]} />
                                            <Tooltip />
                                            <Area type="monotone" dataKey="cpu" stroke="#8884d8" fillOpacity={1} fill="url(#colorCpu)" />
                                        </AreaChart>
                                    </ResponsiveContainer>
                                </div>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardHeader>
                                <CardTitle>Memory Usage</CardTitle>
                                <CardDescription>Aggregate memory consumption.</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="h-[300px]">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <AreaChart data={metricsData}>
                                            <defs>
                                                <linearGradient id="colorMem" x1="0" y1="0" x2="0" y2="1">
                                                    <stop offset="5%" stopColor="#82ca9d" stopOpacity={0.8} />
                                                    <stop offset="95%" stopColor="#82ca9d" stopOpacity={0} />
                                                </linearGradient>
                                            </defs>
                                            <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                            <XAxis dataKey="time" stroke="#888888" fontSize={12} tickLine={false} axisLine={false} />
                                            <YAxis stroke="#888888" fontSize={12} tickLine={false} axisLine={false} unit="%" domain={[0, 100]} />
                                            <Tooltip />
                                            <Area type="monotone" dataKey="memory" stroke="#82ca9d" fillOpacity={1} fill="url(#colorMem)" />
                                        </AreaChart>
                                    </ResponsiveContainer>
                                </div>
                            </CardContent>
                        </Card>

                        <Card className="md:col-span-2">
                            <CardHeader>
                                <CardTitle>Network Traffic</CardTitle>
                                <CardDescription>Inbound and outbound throughput.</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <div className="h-[300px]">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <LineChart data={metricsData}>
                                            <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                            <XAxis dataKey="time" stroke="#888888" fontSize={12} tickLine={false} axisLine={false} />
                                            <YAxis stroke="#888888" fontSize={12} tickLine={false} axisLine={false} unit=" MB/s" />
                                            <Tooltip />
                                            <Line type="monotone" dataKey="networkIn" stroke="#3b82f6" strokeWidth={2} dot={false} name="Inbound" />
                                            <Line type="monotone" dataKey="networkOut" stroke="#f97316" strokeWidth={2} dot={false} name="Outbound" />
                                        </LineChart>
                                    </ResponsiveContainer>
                                </div>
                            </CardContent>
                        </Card>
                    </div>
                </TabsContent>
                <TabsContent value="nodes">
                    <Card>
                        <CardHeader>
                            <CardTitle>Node Metrics</CardTitle>
                            <CardDescription>Per-node resource usage coming soon.</CardDescription>
                        </CardHeader>
                        <CardContent className="h-[400px] flex items-center justify-center text-muted-foreground">
                            Node-level metrics will be available when connected to Hasura.
                        </CardContent>
                    </Card>
                </TabsContent>
                <TabsContent value="pods">
                    <Card>
                        <CardHeader>
                            <CardTitle>Pod Metrics</CardTitle>
                            <CardDescription>Per-pod resource usage coming soon.</CardDescription>
                        </CardHeader>
                        <CardContent className="h-[400px] flex items-center justify-center text-muted-foreground">
                            Pod-level metrics will be available when connected to Hasura.
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
};
