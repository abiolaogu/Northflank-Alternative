/**
 * Mock Data Fixtures
 * Centralized mock data for development and testing
 */

// Application statuses
export type ApplicationStatus = "Running" | "Deploying" | "Stopped" | "Failed";

// Application mock data
export interface MockApplication {
    id: string;
    name: string;
    status: ApplicationStatus;
    image: string;
    replicas: number;
    created_at: string;
}

export const mockApplications: MockApplication[] = [
    { id: "1", name: "frontend-web", status: "Running", image: "nginx:1.25", replicas: 3, created_at: "2024-01-15" },
    { id: "2", name: "api-gateway", status: "Running", image: "envoyproxy/envoy:v1.28", replicas: 2, created_at: "2024-01-12" },
    { id: "3", name: "user-service", status: "Running", image: "app/user-svc:v2.1.0", replicas: 3, created_at: "2024-01-10" },
    { id: "4", name: "payment-service", status: "Deploying", image: "app/payment-svc:v1.5.2", replicas: 2, created_at: "2024-01-08" },
    { id: "5", name: "notification-worker", status: "Stopped", image: "app/notif-worker:v1.0.0", replicas: 0, created_at: "2024-01-05" },
];

// Deployment mock data
export type DeploymentStatus = "success" | "failed" | "deploying" | "rolled_back";

export interface MockDeployment {
    id: string;
    application: string;
    status: DeploymentStatus;
    revision: number;
    image: string;
    commit: string;
    branch: string;
    triggeredBy: string;
    duration: string;
    created_at: string;
}

export const mockDeployments: MockDeployment[] = [
    { id: "1", application: "frontend-web", status: "success", revision: 42, image: "nginx:1.25", commit: "a3f2c1d", branch: "main", triggeredBy: "Olivia Martin", duration: "2m 34s", created_at: "2024-01-17T10:30:00Z" },
    { id: "2", application: "api-gateway", status: "success", revision: 38, image: "envoyproxy/envoy:v1.28", commit: "b4e3d2f", branch: "main", triggeredBy: "Jackson Lee", duration: "1m 45s", created_at: "2024-01-17T09:15:00Z" },
    { id: "3", application: "user-service", status: "deploying", revision: 55, image: "app/user-svc:v2.2.0", commit: "c5f4e3g", branch: "feature/auth", triggeredBy: "System", duration: "0m 45s", created_at: "2024-01-17T11:00:00Z" },
    { id: "4", application: "payment-service", status: "failed", revision: 23, image: "app/payment-svc:v1.5.3", commit: "d6g5h4i", branch: "hotfix/payment", triggeredBy: "Isabella Nguyen", duration: "3m 12s", created_at: "2024-01-17T08:45:00Z" },
    { id: "5", application: "notification-worker", status: "success", revision: 15, image: "app/notif-worker:v1.1.0", commit: "e7h6i5j", branch: "main", triggeredBy: "Olivia Martin", duration: "1m 20s", created_at: "2024-01-16T16:30:00Z" },
    { id: "6", application: "frontend-web", status: "rolled_back", revision: 41, image: "nginx:1.24", commit: "f8i7j6k", branch: "main", triggeredBy: "Jackson Lee", duration: "2m 10s", created_at: "2024-01-16T14:00:00Z" },
];

// Dashboard stats
export interface DashboardStats {
    title: string;
    value: string;
    change: string;
    changeType: "positive" | "negative" | "neutral";
    icon: string;
}

export const dashboardStats: DashboardStats[] = [
    { title: "Total Applications", value: "24", change: "+12%", changeType: "positive", icon: "Box" },
    { title: "Active Databases", value: "12", change: "+4%", changeType: "positive", icon: "Database" },
    { title: "Active Users", value: "573", change: "+201 since last hour", changeType: "positive", icon: "Users" },
    { title: "Success Rate", value: "98.5%", change: "-0.5%", changeType: "negative", icon: "TrendingUp" },
];

// Recent activity
export interface ActivityItem {
    id: string;
    user: string;
    action: string;
    target: string;
    time: string;
    status: "Success" | "Failed" | "Pending" | "Scaled";
}

export const recentActivity: ActivityItem[] = [
    { id: "1", user: "Olivia Martin", action: "Deployed", target: "frontend-app to production", time: "2 mins ago", status: "Success" },
    { id: "2", user: "Jackson Lee", action: "Scaled", target: "database-main to 3 replicas", time: "5 mins ago", status: "Scaled" },
    { id: "3", user: "Isabella Nguyen", action: "Failed deployment", target: "for backend-api", time: "12 mins ago", status: "Failed" },
    { id: "4", user: "William Kim", action: "Created", target: "new project alpha-staging", time: "1 hour ago", status: "Success" },
];

// Database mock data
export type DatabaseType = "PostgreSQL" | "MySQL" | "MongoDB" | "Redis";
export type DatabaseStatus = "Running" | "Provisioning" | "Stopped";

export interface MockDatabase {
    id: string;
    name: string;
    type: DatabaseType;
    status: DatabaseStatus;
    size: string;
    connections: number;
    created_at: string;
}

export const mockDatabases: MockDatabase[] = [
    { id: "1", name: "production-db", type: "PostgreSQL", status: "Running", size: "50GB", connections: 45, created_at: "2024-01-10" },
    { id: "2", name: "analytics-db", type: "PostgreSQL", status: "Running", size: "120GB", connections: 12, created_at: "2024-01-08" },
    { id: "3", name: "cache-store", type: "Redis", status: "Running", size: "8GB", connections: 156, created_at: "2024-01-12" },
    { id: "4", name: "staging-db", type: "PostgreSQL", status: "Provisioning", size: "20GB", connections: 0, created_at: "2024-01-17" },
];

// Project mock data
export interface MockProject {
    id: string;
    name: string;
    description: string;
    services: number;
    databases: number;
    created_at: string;
}

export const mockProjects: MockProject[] = [
    { id: "1", name: "E-Commerce Platform", description: "Main production platform", services: 8, databases: 3, created_at: "2024-01-01" },
    { id: "2", name: "Analytics Service", description: "Data analytics and reporting", services: 4, databases: 2, created_at: "2024-01-05" },
    { id: "3", name: "Mobile Backend", description: "API for mobile applications", services: 6, databases: 2, created_at: "2024-01-08" },
];

// Secret mock data
export interface MockSecret {
    id: string;
    name: string;
    type: "env" | "file" | "docker";
    linkedServices: number;
    lastUpdated: string;
}

export const mockSecrets: MockSecret[] = [
    { id: "1", name: "DATABASE_URL", type: "env", linkedServices: 5, lastUpdated: "2024-01-15" },
    { id: "2", name: "API_KEYS", type: "env", linkedServices: 3, lastUpdated: "2024-01-14" },
    { id: "3", name: "SSL_CERT", type: "file", linkedServices: 8, lastUpdated: "2024-01-10" },
    { id: "4", name: "DOCKER_CONFIG", type: "docker", linkedServices: 12, lastUpdated: "2024-01-12" },
];
