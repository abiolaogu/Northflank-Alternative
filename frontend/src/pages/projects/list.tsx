import React from "react";
import { MoreHorizontal, Plus, Search, FolderKanban, ExternalLink, Rocket } from "lucide-react";
import { Link } from "react-router-dom";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { useNavigation } from "@refinedev/core";

// Mock data for development
const mockProjects = [
    { id: "1", name: "Frontend App", slug: "frontend-app", environment: "Production", services_count: 5, team: "Engineering", updated_at: "2024-01-15" },
    { id: "2", name: "Backend Services", slug: "backend-services", environment: "Production", services_count: 12, team: "Platform", updated_at: "2024-01-14" },
    { id: "3", name: "Data Pipeline", slug: "data-pipeline", environment: "Staging", services_count: 8, team: "Data", updated_at: "2024-01-12" },
    { id: "4", name: "Mobile API", slug: "mobile-api", environment: "Development", services_count: 3, team: "Mobile", updated_at: "2024-01-10" },
];

export const ProjectList = () => {
    const { create } = useNavigation();
    const [searchTerm, setSearchTerm] = React.useState("");

    const filteredProjects = mockProjects.filter(proj =>
        proj.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        proj.slug.toLowerCase().includes(searchTerm.toLowerCase())
    );

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Projects</h2>
                    <p className="text-muted-foreground">
                        Organize your applications and resources into projects.
                    </p>
                </div>
                <Button asChild>
                    <Link to="/projects/create">
                        <Plus className="mr-2 h-4 w-4" />
                        New Project
                    </Link>
                </Button>
            </div>

            <Card>
                <CardHeader>
                    <div className="flex items-center gap-2">
                        <div className="relative flex-1">
                            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search projects..."
                                className="pl-8"
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                            />
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="rounded-md border">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Project Name</TableHead>
                                    <TableHead>Environment</TableHead>
                                    <TableHead>Services</TableHead>
                                    <TableHead>Team</TableHead>
                                    <TableHead>Last Updated</TableHead>
                                    <TableHead className="w-[80px]">Actions</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filteredProjects.length > 0 ? (
                                    filteredProjects.map((project) => (
                                        <TableRow key={project.id}>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    <div className="h-8 w-8 rounded-lg bg-primary/10 flex items-center justify-center">
                                                        <FolderKanban className="h-4 w-4 text-primary" />
                                                    </div>
                                                    <div>
                                                        <div className="font-medium">{project.name}</div>
                                                        <div className="text-xs text-muted-foreground">{project.slug}</div>
                                                    </div>
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                <Badge variant="outline">{project.environment}</Badge>
                                            </TableCell>
                                            <TableCell>{project.services_count}</TableCell>
                                            <TableCell>{project.team}</TableCell>
                                            <TableCell className="text-muted-foreground">
                                                {new Date(project.updated_at).toLocaleDateString()}
                                            </TableCell>
                                            <TableCell>
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <Button variant="ghost" className="h-8 w-8 p-0">
                                                            <MoreHorizontal className="h-4 w-4" />
                                                        </Button>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent align="end">
                                                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                        <DropdownMenuItem>
                                                            <ExternalLink className="mr-2 h-4 w-4" />
                                                            Open Project
                                                        </DropdownMenuItem>
                                                        <DropdownMenuItem>
                                                            <Rocket className="mr-2 h-4 w-4" />
                                                            Deploy All
                                                        </DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuItem>Edit Settings</DropdownMenuItem>
                                                        <DropdownMenuItem className="text-red-600">
                                                            Delete Project
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </TableCell>
                                        </TableRow>
                                    ))
                                ) : (
                                    <TableRow>
                                        <TableCell colSpan={6} className="h-24 text-center">
                                            No projects found.
                                        </TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
};
