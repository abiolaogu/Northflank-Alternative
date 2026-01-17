import { useState, useEffect, useCallback, useRef } from "react";
import { createClient, Client } from "graphql-ws";

// WebSocket configuration
const HASURA_WS_URL = import.meta.env.VITE_HASURA_WS_URL || "wss://hasura.antigravity.io/v1/graphql";

interface SubscriptionOptions<T> {
    query: string;
    variables?: Record<string, any>;
    onData?: (data: T) => void;
    enabled?: boolean;
}

interface UseSubscriptionResult<T> {
    data: T | null;
    loading: boolean;
    error: Error | null;
    connected: boolean;
}

// Global WebSocket client (singleton)
let wsClient: Client | null = null;

function getWsClient(): Client {
    if (!wsClient) {
        wsClient = createClient({
            url: HASURA_WS_URL,
            connectionParams: () => {
                const token = localStorage.getItem("auth_token");
                return {
                    headers: {
                        Authorization: token ? `Bearer ${token}` : "",
                        "x-hasura-admin-secret": import.meta.env.VITE_HASURA_SECRET || "",
                    },
                };
            },
            retryAttempts: 5,
            shouldRetry: () => true,
            on: {
                connected: () => console.log("[WS] Connected to Hasura"),
                error: (err) => console.error("[WS] Connection error:", err),
                closed: () => console.log("[WS] Connection closed"),
            },
        });
    }
    return wsClient;
}

/**
 * Hook for GraphQL subscriptions with Hasura
 */
export function useHasuraSubscription<T = any>(
    options: SubscriptionOptions<T>
): UseSubscriptionResult<T> {
    const { query, variables, onData, enabled = true } = options;
    const [data, setData] = useState<T | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<Error | null>(null);
    const [connected, setConnected] = useState(false);
    const unsubscribeRef = useRef<(() => void) | null>(null);

    useEffect(() => {
        if (!enabled) {
            setLoading(false);
            return;
        }

        const client = getWsClient();
        setLoading(true);

        unsubscribeRef.current = client.subscribe(
            { query, variables },
            {
                next: (result: any) => {
                    setConnected(true);
                    setLoading(false);
                    if (result.data) {
                        setData(result.data);
                        onData?.(result.data);
                    }
                    if (result.errors) {
                        setError(new Error(result.errors[0]?.message || "Subscription error"));
                    }
                },
                error: (err: any) => {
                    setConnected(false);
                    setLoading(false);
                    setError(err instanceof Error ? err : new Error(String(err)));
                },
                complete: () => {
                    setConnected(false);
                    setLoading(false);
                },
            }
        );

        return () => {
            if (unsubscribeRef.current) {
                unsubscribeRef.current();
                unsubscribeRef.current = null;
            }
        };
    }, [query, JSON.stringify(variables), enabled]);

    return { data, loading, error, connected };
}

// Pre-built subscription queries
export const LOGS_SUBSCRIPTION = `
    subscription LogsStream($applicationId: uuid, $level: String, $limit: Int = 100) {
        logs(
            where: {
                _and: [
                    { application_id: { _eq: $applicationId } }
                    { level: { _eq: $level } }
                ]
            }
            order_by: { timestamp: desc }
            limit: $limit
        ) {
            id
            timestamp
            level
            message
            pod_name
            container
            metadata
        }
    }
`;

export const METRICS_SUBSCRIPTION = `
    subscription MetricsStream($applicationId: uuid) {
        metrics(
            where: { application_id: { _eq: $applicationId } }
            order_by: { timestamp: desc }
            limit: 60
        ) {
            id
            timestamp
            cpu_percent
            memory_mb
            network_rx_bytes
            network_tx_bytes
            disk_read_bytes
            disk_write_bytes
        }
    }
`;

export const DEPLOYMENTS_SUBSCRIPTION = `
    subscription DeploymentsStream($applicationId: uuid) {
        deployments(
            where: { application_id: { _eq: $applicationId } }
            order_by: { created_at: desc }
            limit: 10
        ) {
            id
            status
            revision
            image
            replicas
            created_at
            completed_at
        }
    }
`;

// Simulated subscription for development (fallback when Hasura unavailable)
export function useSimulatedLogsSubscription() {
    const [logs, setLogs] = useState<any[]>([]);
    const [isStreaming, setIsStreaming] = useState(true);

    useEffect(() => {
        if (!isStreaming) return;

        const levels = ["INFO", "INFO", "INFO", "WARN", "DEBUG", "ERROR"];
        const apps = ["frontend-app", "backend-api", "worker-service", "db-proxy"];
        const messages = [
            "Processed request took %dms",
            "Database query completed in %dms",
            "Cache hit for key: session_%s",
            "HTTP 200 OK - GET /api/health",
            "Connection pool size: %d active",
            "Memory usage: %dMB / 512MB",
            "Graceful shutdown initiated",
            "New WebSocket connection from %s",
        ];

        const interval = setInterval(() => {
            const level = levels[Math.floor(Math.random() * levels.length)];
            const app = apps[Math.floor(Math.random() * apps.length)];
            const msgTemplate = messages[Math.floor(Math.random() * messages.length)];
            const msg = msgTemplate
                .replace("%d", String(Math.floor(Math.random() * 200)))
                .replace("%s", Math.random().toString(36).substring(7));

            const newLog = {
                id: crypto.randomUUID(),
                timestamp: new Date().toISOString(),
                level,
                message: msg,
                pod_name: `${app}-${Math.random().toString(36).substring(0, 5)}`,
                container: "main",
            };

            setLogs((prev) => [newLog, ...prev].slice(0, 500));
        }, 800);

        return () => clearInterval(interval);
    }, [isStreaming]);

    return { logs, isStreaming, setIsStreaming };
}

// Simulated metrics for development
export function useSimulatedMetricsSubscription() {
    const [metrics, setMetrics] = useState<any[]>([]);

    useEffect(() => {
        // Generate initial metrics history
        const now = Date.now();
        const initial = Array.from({ length: 60 }, (_, i) => ({
            timestamp: new Date(now - (59 - i) * 60000).toISOString(),
            cpu_percent: 30 + Math.random() * 40,
            memory_mb: 200 + Math.random() * 150,
            network_rx_bytes: Math.floor(Math.random() * 1000000),
            network_tx_bytes: Math.floor(Math.random() * 500000),
        }));
        setMetrics(initial);

        const interval = setInterval(() => {
            setMetrics((prev) => {
                const newMetric = {
                    timestamp: new Date().toISOString(),
                    cpu_percent: 30 + Math.random() * 40,
                    memory_mb: 200 + Math.random() * 150,
                    network_rx_bytes: Math.floor(Math.random() * 1000000),
                    network_tx_bytes: Math.floor(Math.random() * 500000),
                };
                return [...prev.slice(1), newMetric];
            });
        }, 60000); // Update every minute

        return () => clearInterval(interval);
    }, []);

    return { metrics };
}
