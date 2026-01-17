import React from "react";
import { useForm } from "@refinedev/react-hook-form";
import { useNavigation } from "@refinedev/core";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { ArrowLeft, Key, Eye, EyeOff } from "lucide-react";
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

const secretSchema = z.object({
    name: z.string()
        .min(1, "Name is required")
        .regex(/^[A-Z][A-Z0-9_]*$/, "Must be uppercase with underscores (e.g., MY_SECRET)"),
    value: z.string().min(1, "Value is required"),
    scope: z.string().min(1, "Please select a scope"),
    description: z.string().optional(),
});

type SecretFormValues = z.infer<typeof secretSchema>;

export const SecretCreate = () => {
    const { list } = useNavigation();
    const [showValue, setShowValue] = React.useState(false);

    const form = useForm<SecretFormValues>({
        resolver: zodResolver(secretSchema),
        defaultValues: {
            name: "",
            value: "",
            scope: "",
            description: "",
        },
        refineCoreProps: {
            resource: "secrets",
            action: "create",
        },
    });

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-4">
                <Button variant="ghost" size="icon" asChild>
                    <Link to="/secrets">
                        <ArrowLeft className="h-4 w-4" />
                    </Link>
                </Button>
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Add Secret</h2>
                    <p className="text-muted-foreground">
                        Create a new encrypted secret variable.
                    </p>
                </div>
            </div>

            <Card className="max-w-2xl">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Key className="h-5 w-5" />
                        Secret Configuration
                    </CardTitle>
                    <CardDescription>
                        Secrets are encrypted at rest and never exposed in logs.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(() => list("secrets"))} className="space-y-4">
                            <FormField
                                control={form.control}
                                name="name"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Secret Name</FormLabel>
                                        <FormControl>
                                            <Input
                                                placeholder="DATABASE_URL"
                                                className="font-mono"
                                                {...field}
                                                onChange={(e) => field.onChange(e.target.value.toUpperCase())}
                                            />
                                        </FormControl>
                                        <FormDescription>
                                            Use SCREAMING_SNAKE_CASE (e.g., API_KEY, DATABASE_URL).
                                        </FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="value"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Secret Value</FormLabel>
                                        <div className="relative">
                                            <FormControl>
                                                <Input
                                                    type={showValue ? "text" : "password"}
                                                    placeholder="Enter secret value..."
                                                    className="font-mono pr-10"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="icon"
                                                className="absolute right-0 top-0 h-full px-3"
                                                onClick={() => setShowValue(!showValue)}
                                            >
                                                {showValue ? (
                                                    <EyeOff className="h-4 w-4" />
                                                ) : (
                                                    <Eye className="h-4 w-4" />
                                                )}
                                            </Button>
                                        </div>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="scope"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Scope</FormLabel>
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select scope" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="global">Global (All environments)</SelectItem>
                                                <SelectItem value="environment">Environment-specific</SelectItem>
                                                <SelectItem value="application">Application-specific</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormDescription>
                                            Determines where this secret is available.
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
                                        <FormLabel>Description (Optional)</FormLabel>
                                        <FormControl>
                                            <Textarea
                                                placeholder="What is this secret used for?"
                                                className="resize-none"
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <div className="flex gap-2 pt-4">
                                <Button type="submit">Create Secret</Button>
                                <Button type="button" variant="outline" onClick={() => list("secrets")}>
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
