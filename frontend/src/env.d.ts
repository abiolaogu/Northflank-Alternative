/// <reference types="vite/client" />

interface ImportMetaEnv {
    /** API base URL */
    readonly VITE_API_URL: string;
    /** Hasura GraphQL endpoint */
    readonly VITE_HASURA_URL: string;
    /** Hasura WebSocket endpoint for subscriptions */
    readonly VITE_HASURA_WS_URL: string;
}

interface ImportMeta {
    readonly env: ImportMetaEnv;
}
