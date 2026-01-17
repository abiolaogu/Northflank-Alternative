import { AuthProvider } from "@refinedev/core";

const API_URL = import.meta.env.VITE_API_URL || "https://api.antigravity.io";

export const authProvider: AuthProvider = {
    login: async ({ email, password }) => {
        try {
            // Mock login for demo purposes if backend not available
            if (email === "demo@antigravity.io" && password === "demo") {
                localStorage.setItem("auth_token", "demo-token");
                localStorage.setItem("user", JSON.stringify({
                    name: "Demo User",
                    email: "demo@antigravity.io",
                    role: "admin",
                    avatar: "https://github.com/shadcn.png"
                }));
                return {
                    success: true,
                    redirectTo: "/",
                };
            }

            const response = await fetch(`${API_URL}/api/v1/auth/login`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ email, password }),
            });

            if (!response.ok) {
                return {
                    success: false,
                    error: {
                        name: "Login Failed",
                        message: "Invalid credentials",
                    },
                };
            }

            const data = await response.json();
            localStorage.setItem("auth_token", data.token);
            localStorage.setItem("user", JSON.stringify(data.user));

            return {
                success: true,
                redirectTo: "/",
            };
        } catch (error) {
            // Fallback for demo
            return {
                success: false,
                error: {
                    name: "Network Error",
                    message: "Unable to connect to server",
                },
            };
        }
    },

    logout: async () => {
        localStorage.removeItem("auth_token");
        localStorage.removeItem("user");
        return {
            success: true,
            redirectTo: "/login",
        };
    },

    check: async () => {
        const token = localStorage.getItem("auth_token");
        if (!token) {
            return {
                authenticated: false,
                redirectTo: "/login",
            };
        }
        return {
            authenticated: true,
        };
    },

    getPermissions: async () => {
        const user = localStorage.getItem("user");
        if (!user) return null;
        const { role } = JSON.parse(user);
        return role;
    },

    getIdentity: async () => {
        const user = localStorage.getItem("user");
        if (!user) return null;
        return JSON.parse(user);
    },

    onError: async (error) => {
        if (error.status === 401) {
            return {
                logout: true,
                redirectTo: "/login",
            };
        }
        return { error };
    },
};
