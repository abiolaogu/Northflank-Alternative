import React from "react";
import { useForm } from "@refinedev/react-hook-form";
import { useNavigation } from "@refinedev/core";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { ArrowLeft, FolderKanban } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
    Form,
    FormControl,
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";

const projectSchema = z.object({
    name: z.string().min(3, "Name must be at least 3 characters"),
    slug: z.string().min(3, "Slug must be at least 3 characters").regex(/^[a-z0-9-]+$/, "Slug must be lowercase alphanumeric with dashes"),
    description: z.string().optional(),
    environment: z.string().min(1, "Please select an environment"),
    team: z.string().min(1, "Please select a team"),
});

type ProjectFormValues = z.infer<typeof projectSchema>;

export const ProjectCreate = () => {
    const { list } = useNavigation();

    const form = useForm<ProjectFormValues>({
        resolver: zodResolver(projectSchema),
        defaultValues: {
            name: "",
            slug: "",
            description: "",
            environment: "",
            team: "",
        },
        refineCoreProps: {
            resource: "projects",
            action: "create",
        },
    });

    // Auto-generate slug from name
    const watchName = form.watch("name");
    React.useEffect(() => {
        if (watchName) {
            const slug = watchName.toLowerCase().replace(/\s+/g, "-").replace(/[^a-z0-9-]/g, "");
            form.setValue("slug", slug);
        }
    }, [watchName, form]);

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-4">
                <Button variant="ghost" size="icon" asChild>
                    <Link to="/projects">
                        <ArrowLeft className="h-4 w-4" />
                    </Link>
                </Button>
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Create Project</h2>
                    <p className="text-muted-foreground">
                        Create a new project to organize your resources.
                    </p>
                </div>
            </div>

            <Card className="max-w-2xl">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <FolderKanban className="h-5 w-5" />
                        Project Details
                    </CardTitle>
                    <CardDescription>
                        Configure your new project settings.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(() => list("projects"))} className="space-y-4">
                            <FormField
                                control={form.control}
                                name="name"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Project Name</FormLabel>
                                        <FormControl>
                                            <Input placeholder="My Project" {...field} />
                                        </FormControl>
                                        <FormDescription>
                                            A descriptive name for your project.
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="slug"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Project Slug</FormLabel>
                                        <FormControl>
                                            <Input placeholder="my-project" {...field} />
                                        </FormControl>
                                        <FormDescription>
                                            Used in URLs and API references.
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="description"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Description</FormLabel>
                                        <FormControl>
                                            <Textarea
                                                placeholder="Describe your project..."
                                                className="resize-none"
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="environment"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Default Environment</FormLabel>
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select environment" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="development">Development</SelectItem>
                                                <SelectItem value="staging">Staging</SelectItem>
                                                <SelectItem value="production">Production</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="team"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Team</FormLabel>
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select team" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="engineering">Engineering</SelectItem>
                                                <SelectItem value="platform">Platform</SelectItem>
                                                <SelectItem value="devops">DevOps</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <div className="flex gap-2 pt-4">
                                <Button type="submit">Create Project</Button>
                                <Button type="button" variant="outline" onClick={() => list("projects")}>
                                    Cancel
                                </Button>
                            </div>
                        </form>
                    </Form>
                </CardContent>
            </Card>
        </div>
    );
};
