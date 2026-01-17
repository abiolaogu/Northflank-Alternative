import React from "react";
import { useForm } from "@refinedev/react-hook-form";
import { useNavigation } from "@refinedev/core";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { ArrowLeft, Database } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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

const databaseSchema = z.object({
    name: z.string().min(3, "Name must be at least 3 characters"),
    type: z.string().min(1, "Please select a database type"),
    version: z.string().min(1, "Please select a version"),
    storage: z.string().min(1, "Please select storage size"),
    region: z.string().min(1, "Please select a region"),
});

type DatabaseFormValues = z.infer<typeof databaseSchema>;

export const DatabaseCreate = () => {
    const { list } = useNavigation();

    const form = useForm<DatabaseFormValues>({
        resolver: zodResolver(databaseSchema),
        defaultValues: {
            name: "",
            type: "",
            version: "",
            storage: "",
            region: "",
        },
        refineCoreProps: {
            resource: "databases",
            action: "create",
        },
    });

    const databaseTypes = [
        { value: "postgresql", label: "PostgreSQL" },
        { value: "mysql", label: "MySQL" },
        { value: "mongodb", label: "MongoDB" },
        { value: "redis", label: "Redis" },
        { value: "clickhouse", label: "ClickHouse" },
    ];

    const versions: Record<string, string[]> = {
        postgresql: ["16", "15", "14", "13"],
        mysql: ["8.0", "5.7"],
        mongodb: ["7.0", "6.0", "5.0"],
        redis: ["7.2", "7.0", "6.2"],
        clickhouse: ["24.1", "23.8"],
    };

    const selectedType = form.watch("type");

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-4">
                <Button variant="ghost" size="icon" asChild>
                    <Link to="/databases">
                        <ArrowLeft className="h-4 w-4" />
                    </Link>
                </Button>
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Create Database</h2>
                    <p className="text-muted-foreground">
                        Deploy a new managed database instance.
                    </p>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Database className="h-5 w-5" />
                            Database Configuration
                        </CardTitle>
                        <CardDescription>
                            Configure your database instance settings.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Form {...form}>
                            <form onSubmit={form.handleSubmit(() => list("databases"))} className="space-y-4">
                                <FormField
                                    control={form.control}
                                    name="name"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Database Name</FormLabel>
                                            <FormControl>
                                                <Input placeholder="my-database" {...field} />
                                            </FormControl>
                                            <FormDescription>
                                                A unique name for your database instance.
                                            </FormDescription>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="type"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Database Type</FormLabel>
                                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder="Select database type" />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectContent>
                                                    {databaseTypes.map((db) => (
                                                        <SelectItem key={db.value} value={db.value}>
                                                            {db.label}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="version"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Version</FormLabel>
                                            <Select
                                                onValueChange={field.onChange}
                                                defaultValue={field.value}
                                                disabled={!selectedType}
                                            >
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder="Select version" />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectContent>
                                                    {(versions[selectedType] || []).map((v) => (
                                                        <SelectItem key={v} value={v}>
                                                            {v}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="storage"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Storage Size</FormLabel>
                                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder="Select storage" />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectContent>
                                                    <SelectItem value="10">10 GB</SelectItem>
                                                    <SelectItem value="25">25 GB</SelectItem>
                                                    <SelectItem value="50">50 GB</SelectItem>
                                                    <SelectItem value="100">100 GB</SelectItem>
                                                    <SelectItem value="250">250 GB</SelectItem>
                                                </SelectContent>
                                            </Select>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="region"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Region</FormLabel>
                                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder="Select region" />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectContent>
                                                    <SelectItem value="us-east-1">US East (N. Virginia)</SelectItem>
                                                    <SelectItem value="eu-west-1">EU West (Ireland)</SelectItem>
                                                    <SelectItem value="ap-southeast-1">Asia Pacific (Singapore)</SelectItem>
                                                </SelectContent>
                                            </Select>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <div className="flex gap-2 pt-4">
                                    <Button type="submit">Create Database</Button>
                                    <Button type="button" variant="outline" onClick={() => list("databases")}>
                                        Cancel
                                    </Button>
                                </div>
                            </form>
                        </Form>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>Pricing Estimate</CardTitle>
                        <CardDescription>
                            Estimated monthly cost based on your configuration.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            <div className="flex justify-between">
                                <span className="text-muted-foreground">Compute</span>
                                <span>$15.00/mo</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-muted-foreground">Storage</span>
                                <span>$2.50/mo</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-muted-foreground">Backups</span>
                                <span>$1.00/mo</span>
                            </div>
                            <div className="border-t pt-4 flex justify-between font-medium">
                                <span>Estimated Total</span>
                                <span>$18.50/mo</span>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
};
