import { DataProvider, LiveProvider } from "@refinedev/core";
import { GraphQLClient } from "graphql-request";
import { createClient } from "graphql-ws";
import dataProviderHasura from "@refinedev/hasura";

// Configuration
const HASURA_URL = import.meta.env.VITE_HASURA_URL || "https://hasura.antigravity.io/v1/graphql";
const HASURA_WS_URL = import.meta.env.VITE_HASURA_WS_URL || "wss://hasura.antigravity.io/v1/graphql";
const API_URL = import.meta.env.VITE_API_URL || "https://api.antigravity.io";

// GraphQL Client for Hasura
// NOTE: Authentication uses JWT tokens only. Admin secret should never be sent from client.
const gqlClient = new GraphQLClient(HASURA_URL, {
    headers: () => {
        const token = localStorage.getItem("auth_token");
        return {
            Authorization: token ? `Bearer ${token}` : "",
        };
    },
});

// WebSocket Client for Live Updates
const wsClient = createClient({
    url: HASURA_WS_URL,
    connectionParams: () => {
        const token = localStorage.getItem("auth_token");
        return {
            headers: {
                Authorization: token ? `Bearer ${token}` : "",
            },
        };
    },
});

// Hasura Data Provider
const hasuraDataProvider = dataProviderHasura(gqlClient as any, {
    namingConvention: "hasura-default",
    idType: "uuid",
});

// Custom REST endpoints for non-Hasura operations
const customEndpoints = {
    // Deploy application
    deploy: async (applicationId: string, options: any) => {
        const response = await fetch(`${API_URL}/api/v1/applications/${applicationId}/deploy`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${localStorage.getItem("auth_token")}`,
            },
            body: JSON.stringify(options),
        });
        return response.json();
    },

    // Scale application
    scale: async (applicationId: string, replicas: number) => {
        const response = await fetch(`${API_URL}/api/v1/applications/${applicationId}/scale`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${localStorage.getItem("auth_token")}`,
            },
            body: JSON.stringify({ replicas }),
        });
        return response.json();
    },

    // Restart application
    restart: async (applicationId: string) => {
        const response = await fetch(`${API_URL}/api/v1/applications/${applicationId}/restart`, {
            method: "POST",
            headers: {
                Authorization: `Bearer ${localStorage.getItem("auth_token")}`,
            },
        });
        return response.json();
    },

    // Rollback deployment
    rollback: async (deploymentId: string, targetRevision: number) => {
        const response = await fetch(`${API_URL}/api/v1/deployments/${deploymentId}/rollback`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${localStorage.getItem("auth_token")}`,
            },
            body: JSON.stringify({ targetRevision }),
        });
        return response.json();
    },

    // Get logs
    getLogs: async (applicationId: string, options: any) => {
        const params = new URLSearchParams({
            since: options.since || "1h",
            limit: String(options.limit || 100),
            ...(options.container && { container: options.container }),
        });
        const response = await fetch(
            `${API_URL}/api/v1/applications/${applicationId}/logs?${params}`,
            {
                headers: {
                    Authorization: `Bearer ${localStorage.getItem("auth_token")}`,
                },
            }
        );
        return response.json();
    },

    // Get metrics
    getMetrics: async (applicationId: string, options: any) => {
        const params = new URLSearchParams({
            start: options.start,
            end: options.end,
            step: options.step || "1m",
        });
        const response = await fetch(
            `${API_URL}/api/v1/applications/${applicationId}/metrics?${params}`,
            {
                headers: {
                    Authorization: `Bearer ${localStorage.getItem("auth_token")}`,
                },
            }
        );
        return response.json();
    },
};

// Combined Data Provider
export const dataProvider: DataProvider = {
    ...hasuraDataProvider,

    // Override create for special handling
    create: async ({ resource, variables, meta }) => {
        // Use custom endpoint for certain resources
        if (meta?.customEndpoint) {
            const response = await fetch(`${API_URL}${meta.customEndpoint}`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${localStorage.getItem("auth_token")}`,
                },
                body: JSON.stringify(variables),
            });
            const data = await response.json();
            return { data };
        }

        // Default to Hasura
        return hasuraDataProvider.create({ resource, variables, meta });
    },

    // Custom methods accessible via useCustom
    custom: async ({ url, method, payload, headers }) => {
        const response = await fetch(`${API_URL}${url}`, {
            method: method || "GET",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${localStorage.getItem("auth_token")}`,
                ...headers,
            },
            ...(payload && { body: JSON.stringify(payload) }),
        });
        const data = await response.json();
        return { data };
    },
};

// Live Provider for real-time updates
export const liveProvider: LiveProvider = {
    subscribe: ({ channel, types, params, callback }) => {
        // GraphQL subscription query based on channel (resource name)
        const subscriptionQuery = `
      subscription ${channel}Subscription {
        ${channel}(
          ${params?.ids ? `where: { id: { _in: ${JSON.stringify(params.ids)} } }` : ""}
          ${params?.id ? `where: { id: { _eq: "${params.id}" } }` : ""}
          order_by: { updated_at: desc }
          limit: 1
        ) {
          id
          updated_at
        }
      }
    `;

        const unsubscribe = wsClient.subscribe(
            { query: subscriptionQuery },
            {
                next: (data) => {
                    callback({
                        channel,
                        type: "updated",
                        payload: data.data,
                        date: new Date(),
                    });
                },
                error: (error) => {
                    console.error("Subscription error:", error);
                },
                complete: () => {
                    console.log("Subscription complete");
                },
            }
        );

        return unsubscribe;
    },

    unsubscribe: (unsubscribe) => {
        unsubscribe();
    },
};
