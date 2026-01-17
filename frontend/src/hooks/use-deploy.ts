import { useState, useCallback } from "react";
import { useToast } from "@/hooks/use-toast";

const API_URL = import.meta.env.VITE_API_URL || "https://api.antigravity.io";

interface DeployOptions {
    image?: string;
    tag?: string;
    replicas?: number;
    environment?: Record<string, string>;
}

interface DeploymentResult {
    id: string;
    status: string;
    message?: string;
}

interface UseDeployResult {
    deploy: (applicationId: string, options?: DeployOptions) => Promise<DeploymentResult>;
    scale: (applicationId: string, replicas: number) => Promise<DeploymentResult>;
    restart: (applicationId: string) => Promise<DeploymentResult>;
    rollback: (deploymentId: string, revision: number) => Promise<DeploymentResult>;
    isLoading: boolean;
    error: Error | null;
}

/**
 * Hook for deployment mutations (deploy, scale, restart, rollback)
 */
export function useDeploy(): UseDeployResult {
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);
    const { toast } = useToast();

    const getAuthHeaders = () => ({
        "Content-Type": "application/json",
        Authorization: `Bearer ${localStorage.getItem("auth_token") || ""}`,
    });

    const handleError = (err: Error, action: string) => {
        setError(err);
        toast({
            title: `${action} Failed`,
            description: err.message,
            variant: "destructive",
        });
    };

    const deploy = useCallback(async (applicationId: string, options?: DeployOptions): Promise<DeploymentResult> => {
        setIsLoading(true);
        setError(null);

        try {
            const response = await fetch(`${API_URL}/api/v1/applications/${applicationId}/deploy`, {
                method: "POST",
                headers: getAuthHeaders(),
                body: JSON.stringify(options || {}),
            });

            if (!response.ok) {
                throw new Error(`Deploy failed: ${response.statusText}`);
            }

            const result = await response.json();

            toast({
                title: "Deployment Started",
                description: `Deployment ${result.id} initiated successfully.`,
            });

            return result;
        } catch (err) {
            const error = err instanceof Error ? err : new Error(String(err));
            handleError(error, "Deployment");
            throw error;
        } finally {
            setIsLoading(false);
        }
    }, [toast]);

    const scale = useCallback(async (applicationId: string, replicas: number): Promise<DeploymentResult> => {
        setIsLoading(true);
        setError(null);

        try {
            const response = await fetch(`${API_URL}/api/v1/applications/${applicationId}/scale`, {
                method: "POST",
                headers: getAuthHeaders(),
                body: JSON.stringify({ replicas }),
            });

            if (!response.ok) {
                throw new Error(`Scale failed: ${response.statusText}`);
            }

            const result = await response.json();

            toast({
                title: "Scaling Initiated",
                description: `Application scaling to ${replicas} replicas.`,
            });

            return result;
        } catch (err) {
            const error = err instanceof Error ? err : new Error(String(err));
            handleError(error, "Scaling");
            throw error;
        } finally {
            setIsLoading(false);
        }
    }, [toast]);

    const restart = useCallback(async (applicationId: string): Promise<DeploymentResult> => {
        setIsLoading(true);
        setError(null);

        try {
            const response = await fetch(`${API_URL}/api/v1/applications/${applicationId}/restart`, {
                method: "POST",
                headers: getAuthHeaders(),
            });

            if (!response.ok) {
                throw new Error(`Restart failed: ${response.statusText}`);
            }

            const result = await response.json();

            toast({
                title: "Restart Initiated",
                description: "Application is restarting.",
            });

            return result;
        } catch (err) {
            const error = err instanceof Error ? err : new Error(String(err));
            handleError(error, "Restart");
            throw error;
        } finally {
            setIsLoading(false);
        }
    }, [toast]);

    const rollback = useCallback(async (deploymentId: string, revision: number): Promise<DeploymentResult> => {
        setIsLoading(true);
        setError(null);

        try {
            const response = await fetch(`${API_URL}/api/v1/deployments/${deploymentId}/rollback`, {
                method: "POST",
                headers: getAuthHeaders(),
                body: JSON.stringify({ targetRevision: revision }),
            });

            if (!response.ok) {
                throw new Error(`Rollback failed: ${response.statusText}`);
            }

            const result = await response.json();

            toast({
                title: "Rollback Initiated",
                description: `Rolling back to revision ${revision}.`,
            });

            return result;
        } catch (err) {
            const error = err instanceof Error ? err : new Error(String(err));
            handleError(error, "Rollback");
            throw error;
        } finally {
            setIsLoading(false);
        }
    }, [toast]);

    return { deploy, scale, restart, rollback, isLoading, error };
}

/**
 * Simulated deploy for development (when API is unavailable)
 */
export function useSimulatedDeploy(): UseDeployResult {
    const [isLoading, setIsLoading] = useState(false);
    const [error] = useState<Error | null>(null);
    const { toast } = useToast();

    const simulateDelay = () => new Promise(resolve => setTimeout(resolve, 1500));

    const deploy = useCallback(async (applicationId: string, options?: DeployOptions): Promise<DeploymentResult> => {
        setIsLoading(true);
        await simulateDelay();

        const result = {
            id: `deploy-${Date.now()}`,
            status: "deploying",
            message: "Deployment initiated (simulated)",
        };

        toast({
            title: "🚀 Deployment Started",
            description: `Deploying ${applicationId}...`,
        });

        setIsLoading(false);
        return result;
    }, [toast]);

    const scale = useCallback(async (applicationId: string, replicas: number): Promise<DeploymentResult> => {
        setIsLoading(true);
        await simulateDelay();

        const result = {
            id: `scale-${Date.now()}`,
            status: "scaling",
            message: `Scaling to ${replicas} replicas (simulated)`,
        };

        toast({
            title: "📈 Scaling Initiated",
            description: `Scaling ${applicationId} to ${replicas} replicas.`,
        });

        setIsLoading(false);
        return result;
    }, [toast]);

    const restart = useCallback(async (applicationId: string): Promise<DeploymentResult> => {
        setIsLoading(true);
        await simulateDelay();

        const result = {
            id: `restart-${Date.now()}`,
            status: "restarting",
            message: "Restart initiated (simulated)",
        };

        toast({
            title: "🔄 Restart Initiated",
            description: `Restarting ${applicationId}...`,
        });

        setIsLoading(false);
        return result;
    }, [toast]);

    const rollback = useCallback(async (deploymentId: string, revision: number): Promise<DeploymentResult> => {
        setIsLoading(true);
        await simulateDelay();

        const result = {
            id: `rollback-${Date.now()}`,
            status: "rolling_back",
            message: `Rolling back to revision ${revision} (simulated)`,
        };

        toast({
            title: "⏪ Rollback Initiated",
            description: `Rolling back to revision ${revision}.`,
        });

        setIsLoading(false);
        return result;
    }, [toast]);

    return { deploy, scale, restart, rollback, isLoading, error };
}
