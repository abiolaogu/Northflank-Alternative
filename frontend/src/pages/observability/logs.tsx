import React, { useState, useRef } from "react";
import {
    Play,
    Pause,
    Download,
    Search,
    Clock,
    RefreshCw,
    Wifi,
    WifiOff,
} from "lucide-react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { useSimulatedLogsSubscription } from "@/hooks/use-hasura-subscription";

interface LogEntry {
    id: string;
    timestamp: string;
    level: "INFO" | "WARN" | "ERROR" | "DEBUG";
    message: string;
    pod_name: string;
    container: string;
}

export const LogsPage = () => {
    const [search, setSearch] = useState("");
    const [selectedApp, setSelectedApp] = useState<string>("all");
    const [selectedLevel, setSelectedLevel] = useState<string>("all");
    const [timeRange, setTimeRange] = useState("1h");
    const scrollRef = useRef<HTMLDivElement>(null);

    // Use simulated subscription (real Hasura would use useHasuraSubscription)
    const { logs, isStreaming, setIsStreaming } = useSimulatedLogsSubscription();

    // Filter logs based on search and filters
    const filteredLogs = logs.filter((log: LogEntry) => {
        const matchesSearch = !search ||
            log.message.toLowerCase().includes(search.toLowerCase()) ||
            log.pod_name.toLowerCase().includes(search.toLowerCase());
        const matchesLevel = selectedLevel === "all" || log.level === selectedLevel;
        const matchesApp = selectedApp === "all" || log.pod_name.startsWith(selectedApp);
        return matchesSearch && matchesLevel && matchesApp;
    });

    // Auto-scroll when streaming
    React.useEffect(() => {
        if (isStreaming && scrollRef.current) {
            const scrollable = scrollRef.current.querySelector('[data-radix-scroll-area-viewport]');
            if (scrollable) {
                scrollable.scrollTop = 0; // Scroll to top since logs are prepended
            }
        }
    }, [logs, isStreaming]);

    const exportLogs = () => {
        const content = filteredLogs
            .map(l => `${l.timestamp} [${l.level}] [${l.pod_name}] ${l.message}`)
            .join('\n');
        const blob = new Blob([content], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `logs-${new Date().toISOString().split('T')[0]}.txt`;
        a.click();
        URL.revokeObjectURL(url);
    };

    const getLevelColor = (level: string) => {
        switch (level) {
            case 'ERROR': return 'text-red-500';
            case 'WARN': return 'text-yellow-500';
            case 'DEBUG': return 'text-blue-500';
            default: return 'text-green-500';
        }
    };

    return (
        <div className="flex flex-col h-[calc(100vh-8rem)] gap-4">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">System Logs</h2>
                    <p className="text-muted-foreground">
                        Real-time stream of application and infrastructure logs.
                    </p>
                </div>
                <div className="flex gap-2">
                    <Badge variant={isStreaming ? "default" : "secondary"} className="gap-1">
                        {isStreaming ? <Wifi className="h-3 w-3" /> : <WifiOff className="h-3 w-3" />}
                        {isStreaming ? "Live" : "Paused"}
                    </Badge>
                    <Button variant="outline" size="sm" onClick={() => setIsStreaming(false)}>
                        <RefreshCw className="mr-2 h-4 w-4" />
                        Clear
                    </Button>
                    <Button variant="outline" size="sm" onClick={exportLogs}>
                        <Download className="mr-2 h-4 w-4" />
                        Export
                    </Button>
                </div>
            </div>

            <Card className="flex flex-col flex-1 overflow-hidden">
                <CardHeader className="py-4 border-b">
                    <div className="flex flex-col lg:flex-row gap-4 justify-between items-stretch lg:items-center">
                        <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 flex-1">
                            <div className="relative flex-1 max-w-md">
                                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                                <Input
                                    placeholder="Search logs..."
                                    className="pl-8 h-9"
                                    value={search}
                                    onChange={(e) => setSearch(e.target.value)}
                                />
                            </div>
                            <Select value={selectedApp} onValueChange={setSelectedApp}>
                                <SelectTrigger className="w-full sm:w-[160px] h-9">
                                    <SelectValue placeholder="Application" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Applications</SelectItem>
                                    <SelectItem value="frontend-app">frontend-app</SelectItem>
                                    <SelectItem value="backend-api">backend-api</SelectItem>
                                    <SelectItem value="worker-service">worker-service</SelectItem>
                                    <SelectItem value="db-proxy">db-proxy</SelectItem>
                                </SelectContent>
                            </Select>
                            <Select value={selectedLevel} onValueChange={setSelectedLevel}>
                                <SelectTrigger className="w-full sm:w-[120px] h-9">
                                    <SelectValue placeholder="Level" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">All Levels</SelectItem>
                                    <SelectItem value="ERROR">Error</SelectItem>
                                    <SelectItem value="WARN">Warning</SelectItem>
                                    <SelectItem value="INFO">Info</SelectItem>
                                    <SelectItem value="DEBUG">Debug</SelectItem>
                                </SelectContent>
                            </Select>
                            <Select value={timeRange} onValueChange={setTimeRange}>
                                <SelectTrigger className="w-full sm:w-[120px] h-9">
                                    <Clock className="mr-2 h-3 w-3" />
                                    <SelectValue placeholder="Range" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="5m">Last 5m</SelectItem>
                                    <SelectItem value="15m">Last 15m</SelectItem>
                                    <SelectItem value="1h">Last 1h</SelectItem>
                                    <SelectItem value="6h">Last 6h</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        <div className="flex items-center gap-2">
                            <Button
                                variant={isStreaming ? "destructive" : "default"}
                                size="sm"
                                onClick={() => setIsStreaming(!isStreaming)}
                            >
                                {isStreaming ? (
                                    <>
                                        <Pause className="mr-2 h-3 w-3" />
                                        Pause
                                    </>
                                ) : (
                                    <>
                                        <Play className="mr-2 h-3 w-3" />
                                        Resume
                                    </>
                                )}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent className="flex-1 p-0 overflow-hidden bg-zinc-950 text-zinc-300 font-mono text-xs">
                    <ScrollArea className="h-full" ref={scrollRef}>
                        <div className="p-4 space-y-1">
                            {filteredLogs.map((log: LogEntry) => (
                                <div key={log.id} className="flex gap-2 md:gap-4 hover:bg-zinc-900/50 px-2 py-0.5 rounded-sm group">
                                    <span className="text-zinc-500 w-[80px] md:w-[170px] shrink-0 select-none truncate">
                                        {new Date(log.timestamp).toLocaleTimeString()}
                                    </span>
                                    <span className={`w-[50px] md:w-[60px] shrink-0 font-bold ${getLevelColor(log.level)}`}>
                                        {log.level}
                                    </span>
                                    <span className="text-zinc-500 w-[100px] md:w-[140px] shrink-0 truncate hidden sm:block" title={log.pod_name}>
                                        [{log.pod_name}]
                                    </span>
                                    <span className="text-zinc-300 break-all group-hover:text-white transition-colors">
                                        {log.message}
                                    </span>
                                </div>
                            ))}
                            {filteredLogs.length === 0 && (
                                <div className="flex items-center justify-center h-full text-zinc-600 italic py-8">
                                    {isStreaming ? "Waiting for logs..." : "No logs match your filters"}
                                </div>
                            )}
                        </div>
                    </ScrollArea>
                </CardContent>
            </Card>
        </div>
    );
};
