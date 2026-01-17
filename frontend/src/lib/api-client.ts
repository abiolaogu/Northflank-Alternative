/**
 * Centralized API Client
 * Provides type-safe HTTP methods with automatic auth handling
 */

const API_URL = import.meta.env.VITE_API_URL || "https://api.antigravity.io";

interface RequestOptions {
    headers?: Record<string, string>;
    signal?: AbortSignal;
}

interface ApiResponse<T = unknown> {
    data: T;
    status: number;
    ok: boolean;
}

class ApiError extends Error {
    constructor(
        message: string,
        public status: number,
        public data?: unknown
    ) {
        super(message);
        this.name = "ApiError";
    }
}

const getAuthHeaders = (): Record<string, string> => {
    const token = localStorage.getItem("auth_token");
    return token ? { Authorization: `Bearer ${token}` } : {};
};

const handleResponse = async <T>(response: Response): Promise<ApiResponse<T>> => {
    const data = await response.json().catch(() => null);

    if (!response.ok) {
        throw new ApiError(
            data?.message || `HTTP error ${response.status}`,
            response.status,
            data
        );
    }

    return {
        data: data as T,
        status: response.status,
        ok: response.ok,
    };
};

export const apiClient = {
    /**
     * GET request
     */
    get: async <T = unknown>(
        endpoint: string,
        options?: RequestOptions
    ): Promise<ApiResponse<T>> => {
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: "GET",
            headers: {
                ...getAuthHeaders(),
                ...options?.headers,
            },
            signal: options?.signal,
        });
        return handleResponse<T>(response);
    },

    /**
     * POST request
     */
    post: async <T = unknown>(
        endpoint: string,
        data?: unknown,
        options?: RequestOptions
    ): Promise<ApiResponse<T>> => {
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                ...getAuthHeaders(),
                ...options?.headers,
            },
            body: data ? JSON.stringify(data) : undefined,
            signal: options?.signal,
        });
        return handleResponse<T>(response);
    },

    /**
     * PUT request
     */
    put: async <T = unknown>(
        endpoint: string,
        data?: unknown,
        options?: RequestOptions
    ): Promise<ApiResponse<T>> => {
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json",
                ...getAuthHeaders(),
                ...options?.headers,
            },
            body: data ? JSON.stringify(data) : undefined,
            signal: options?.signal,
        });
        return handleResponse<T>(response);
    },

    /**
     * DELETE request
     */
    delete: async <T = unknown>(
        endpoint: string,
        options?: RequestOptions
    ): Promise<ApiResponse<T>> => {
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: "DELETE",
            headers: {
                ...getAuthHeaders(),
                ...options?.headers,
            },
            signal: options?.signal,
        });
        return handleResponse<T>(response);
    },

    /**
     * PATCH request
     */
    patch: async <T = unknown>(
        endpoint: string,
        data?: unknown,
        options?: RequestOptions
    ): Promise<ApiResponse<T>> => {
        const response = await fetch(`${API_URL}${endpoint}`, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
                ...getAuthHeaders(),
                ...options?.headers,
            },
            body: data ? JSON.stringify(data) : undefined,
            signal: options?.signal,
        });
        return handleResponse<T>(response);
    },
};

export { ApiError };
export type { ApiResponse, RequestOptions };
