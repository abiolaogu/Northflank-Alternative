import React, { Component, ErrorInfo, ReactNode } from "react";
import { AlertTriangle, RefreshCw, Home } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
}

interface State {
    hasError: boolean;
    error: Error | null;
    errorInfo: ErrorInfo | null;
}

/**
 * Error Boundary component for graceful error handling
 * Catches JavaScript errors anywhere in the child component tree
 */
export class ErrorBoundary extends Component<Props, State> {
    public state: State = {
        hasError: false,
        error: null,
        errorInfo: null,
    };

    public static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error, errorInfo: null };
    }

    public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error("Error Boundary caught an error:", error, errorInfo);
        this.setState({ errorInfo });

        // You could log to an error reporting service here
        // e.g., Sentry.captureException(error);
    }

    private handleReset = () => {
        this.setState({ hasError: false, error: null, errorInfo: null });
    };

    private handleGoHome = () => {
        window.location.href = "/";
    };

    private handleRefresh = () => {
        window.location.reload();
    };

    public render() {
        if (this.state.hasError) {
            if (this.props.fallback) {
                return this.props.fallback;
            }

            return (
                <div className="flex min-h-screen items-center justify-center bg-muted/20 p-4">
                    <Card className="w-full max-w-lg">
                        <CardHeader className="text-center">
                            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
                                <AlertTriangle className="h-8 w-8 text-destructive" />
                            </div>
                            <CardTitle className="text-2xl">Something went wrong</CardTitle>
                            <CardDescription>
                                An unexpected error occurred. We apologize for the inconvenience.
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {this.state.error && (
                                <div className="rounded-lg bg-muted p-4">
                                    <p className="text-sm font-medium text-destructive">
                                        {this.state.error.message}
                                    </p>
                                    {import.meta.env.DEV && this.state.errorInfo && (
                                        <pre className="mt-2 max-h-32 overflow-auto text-xs text-muted-foreground">
                                            {this.state.errorInfo.componentStack}
                                        </pre>
                                    )}
                                </div>
                            )}
                        </CardContent>
                        <CardFooter className="flex gap-2">
                            <Button
                                variant="outline"
                                className="flex-1"
                                onClick={this.handleGoHome}
                            >
                                <Home className="mr-2 h-4 w-4" />
                                Go Home
                            </Button>
                            <Button
                                className="flex-1"
                                onClick={this.handleRefresh}
                            >
                                <RefreshCw className="mr-2 h-4 w-4" />
                                Refresh
                            </Button>
                        </CardFooter>
                    </Card>
                </div>
            );
        }

        return this.props.children;
    }
}
