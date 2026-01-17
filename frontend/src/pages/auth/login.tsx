import { useLogin } from "@refinedev/core";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Rocket } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";

const formSchema = z.object({
    email: z.string().email({
        message: "Please enter a valid email address.",
    }),
    password: z.string().min(1, {
        message: "Password is required.",
    }),
});

export const LoginPage = () => {
    const { mutate: login, isLoading } = useLogin() as any;
    const { toast } = useToast();

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            email: "demo@antigravity.io",
            password: "demo",
        },
    });

    const onSubmit = (values: z.infer<typeof formSchema>) => {
        login(values, {
            onError: (error) => {
                toast({
                    variant: "destructive",
                    title: "Login Failed",
                    description: error?.message || "Invalid credentials. Please try again.",
                });
            }
        });
    };

    return (
        <div className="flex h-screen items-center justify-center bg-muted/20">
            <Card className="w-[380px]">
                <CardHeader className="text-center">
                    <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-lg bg-primary">
                        <Rocket className="h-6 w-6 text-primary-foreground" />
                    </div>
                    <CardTitle className="text-2xl">Antigravity Portal</CardTitle>
                    <CardDescription>
                        Enter your credentials to access the platform.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                            <FormField
                                control={form.control}
                                name="email"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Email</FormLabel>
                                        <FormControl>
                                            <Input placeholder="name@company.com" {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="password"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Password</FormLabel>
                                        <FormControl>
                                            <Input type="password" {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <Button type="submit" className="w-full" disabled={isLoading}>
                                {isLoading ? "Signing in..." : "Sign In"}
                            </Button>
                        </form>
                    </Form>
                </CardContent>
                <CardFooter className="flex flex-col gap-2">
                    <div className="text-xs text-center text-muted-foreground">
                        Demo Credentials: <br />
                        <span className="font-mono">demo@antigravity.io</span> / <span className="font-mono">demo</span>
                    </div>
                </CardFooter>
            </Card>
        </div>
    );
};
