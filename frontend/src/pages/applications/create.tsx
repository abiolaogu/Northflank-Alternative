import { useForm } from "@refinedev/react-hook-form";
import { useNavigation } from "@refinedev/core";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import {
    Form,
    FormControl,
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArrowLeft } from "lucide-react";

// Schema definition
const formSchema = z.object({
    name: z.string().min(2, {
        message: "Name must be at least 2 characters.",
    }),
    description: z.string().optional(),
    sourceType: z.enum(["git", "docker", "template"]),
    // Git Source
    repository: z.string().optional(),
    branch: z.string().optional(),
    // Docker Source
    image: z.string().optional(),
    // Resources
    cpu: z.string().default("0.5"),
    memory: z.string().default("512Mi"),
    replicas: z.coerce.number().min(1).default(1),
    port: z.coerce.number().min(1).default(8080),
});

export const ApplicationCreate = () => {
    const { list } = useNavigation();

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            sourceType: "git",
            cpu: "0.5",
            memory: "512Mi",
            replicas: 1,
            port: 8080,
        },
        refineCoreProps: {
            resource: "applications",
            redirect: "list",
            action: "create",
        }
    });

    const sourceType = form.watch("sourceType");

    return (
        <div className="max-w-3xl mx-auto py-6 space-y-6">
            <div className="flex items-center gap-4">
                <Button variant="ghost" size="icon" onClick={() => list("applications")}>
                    <ArrowLeft className="h-4 w-4" />
                </Button>
                <div>
                    <h1 className="text-2xl font-bold tracking-tight">Create Application</h1>
                    <p className="text-muted-foreground">Deploy a new service to your cluster.</p>
                </div>
            </div>

            <Form {...form}>
                <form onSubmit={form.handleSubmit(form.saveButtonProps.onClick)} className="space-y-8">

                    <Card>
                        <CardHeader>
                            <CardTitle>General Information</CardTitle>
                            <CardDescription>Basic details about your application.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <FormField
                                control={form.control}
                                name="name"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Application Name</FormLabel>
                                        <FormControl>
                                            <Input placeholder="my-awesome-app" {...field} />
                                        </FormControl>
                                        <FormDescription>
                                            Unique identifier for your application within the project.
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
                                            <Input placeholder="A brief description..." {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Source Configuration</CardTitle>
                            <CardDescription>Where should we fetch your code from?</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <FormField
                                control={form.control}
                                name="sourceType"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Source Type</FormLabel>
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select a source type" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="git">Git Repository</SelectItem>
                                                <SelectItem value="docker">Docker Image</SelectItem>
                                                <SelectItem value="template">Template</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {sourceType === "git" && (
                                <div className="grid grid-cols-2 gap-4">
                                    <FormField
                                        control={form.control}
                                        name="repository"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Repository URL</FormLabel>
                                                <FormControl>
                                                    <Input placeholder="https://github.com/username/repo" {...field} />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={form.control}
                                        name="branch"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>Branch</FormLabel>
                                                <FormControl>
                                                    <Input placeholder="main" {...field} />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                </div>
                            )}

                            {sourceType === "docker" && (
                                <FormField
                                    control={form.control}
                                    name="image"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Image Name</FormLabel>
                                            <FormControl>
                                                <Input placeholder="nginx:latest" {...field} />
                                            </FormControl>
                                            <FormDescription>The full Docker image name and tag.</FormDescription>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            )}
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Resources & Networking</CardTitle>
                            <CardDescription>Configure compute resources and network exposure.</CardDescription>
                        </CardHeader>
                        <CardContent className="grid grid-cols-2 gap-6">
                            <FormField
                                control={form.control}
                                name="cpu"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>CPU Limit</FormLabel>
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select CPU" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="0.1">0.1 vCPU</SelectItem>
                                                <SelectItem value="0.5">0.5 vCPU</SelectItem>
                                                <SelectItem value="1">1 vCPU</SelectItem>
                                                <SelectItem value="2">2 vCPU</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="memory"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Memory Limit</FormLabel>
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select Memory" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="128Mi">128 MiB</SelectItem>
                                                <SelectItem value="256Mi">256 MiB</SelectItem>
                                                <SelectItem value="512Mi">512 MiB</SelectItem>
                                                <SelectItem value="1Gi">1 GiB</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="replicas"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Replicas</FormLabel>
                                        <FormControl>
                                            <Input type="number" {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="port"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Container Port</FormLabel>
                                        <FormControl>
                                            <Input type="number" {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </CardContent>
                    </Card>

                    <div className="flex justify-end gap-4">
                        <Button variant="outline" type="button" onClick={() => list("applications")}>Cancel</Button>
                        <Button type="submit" disabled={form.formState.isSubmitting}>
                            {form.formState.isSubmitting ? "Deploying..." : "Deploy Application"}
                        </Button>
                    </div>

                </form>
            </Form>
        </div>
    );
};
