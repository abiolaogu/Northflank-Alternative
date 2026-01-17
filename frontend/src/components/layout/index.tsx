import React from "react";
import { useLogout, useGetIdentity } from "@refinedev/core";
import { Link, useLocation, Outlet } from "react-router-dom";
import {
    LayoutDashboard,
    Box,
    Database,
    Rocket,
    FolderKanban,
    Key,
    FileText,
    BarChart3,
    Activity,
    Settings,
    LogOut,
    ChevronDown,
    Moon,
    Sun,
    Bell,
    Search,
    Menu,
    X,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useTheme } from "@/hooks/use-theme";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { CommandDialog } from "@/components/ui/command-palette";
import { Toaster } from "@/components/ui/toaster";
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from "@/components/ui/sheet";

interface LayoutProps {
    children?: React.ReactNode;
}

const navigation = [
    { name: "Dashboard", href: "/", icon: LayoutDashboard },
    { name: "Applications", href: "/applications", icon: Box },
    { name: "Databases", href: "/databases", icon: Database },
    { name: "Deployments", href: "/deployments", icon: Rocket },
    { name: "Projects", href: "/projects", icon: FolderKanban },
    { name: "Secrets", href: "/secrets", icon: Key },
];

const observability = [
    { name: "Logs", href: "/logs", icon: FileText },
    { name: "Metrics", href: "/metrics", icon: BarChart3 },
    { name: "Traces", href: "/traces", icon: Activity },
];

// Reusable navigation component
const NavigationLinks: React.FC<{ onNavigate?: () => void }> = ({ onNavigate }) => {
    const location = useLocation();

    return (
        <>
            <div className="space-y-1">
                {navigation.map((item) => {
                    const isActive = location.pathname === item.href ||
                        (item.href !== "/" && location.pathname.startsWith(item.href));
                    return (
                        <Link
                            key={item.name}
                            to={item.href}
                            onClick={onNavigate}
                            className={cn(
                                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                                isActive
                                    ? "bg-primary text-primary-foreground"
                                    : "text-muted-foreground hover:bg-muted hover:text-foreground"
                            )}
                        >
                            <item.icon className="h-4 w-4" />
                            {item.name}
                        </Link>
                    );
                })}
            </div>

            <div className="mt-8">
                <h3 className="px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Observability
                </h3>
                <div className="mt-2 space-y-1">
                    {observability.map((item) => {
                        const isActive = location.pathname === item.href;
                        return (
                            <Link
                                key={item.name}
                                to={item.href}
                                onClick={onNavigate}
                                className={cn(
                                    "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                                    isActive
                                        ? "bg-primary text-primary-foreground"
                                        : "text-muted-foreground hover:bg-muted hover:text-foreground"
                                )}
                            >
                                <item.icon className="h-4 w-4" />
                                {item.name}
                            </Link>
                        );
                    })}
                </div>
            </div>

            <div className="mt-8">
                <h3 className="px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Settings
                </h3>
                <div className="mt-2 space-y-1">
                    <Link
                        to="/settings"
                        onClick={onNavigate}
                        className={cn(
                            "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                            location.pathname === "/settings"
                                ? "bg-primary text-primary-foreground"
                                : "text-muted-foreground hover:bg-muted hover:text-foreground"
                        )}
                    >
                        <Settings className="h-4 w-4" />
                        API Explorer
                    </Link>
                </div>
            </div>
        </>
    );
};

export const Layout: React.FC<LayoutProps> = ({ children }) => {
    const { mutate: logout } = useLogout();
    const { data: user } = useGetIdentity<{ name: string; email: string; avatar?: string }>();
    const location = useLocation();
    const { theme, setTheme } = useTheme();
    const [commandOpen, setCommandOpen] = React.useState(false);
    const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);

    // Close mobile menu on route change
    React.useEffect(() => {
        setMobileMenuOpen(false);
    }, [location.pathname]);

    // Keyboard shortcut for command palette
    React.useEffect(() => {
        const down = (e: KeyboardEvent) => {
            if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
                e.preventDefault();
                setCommandOpen((open) => !open);
            }
        };
        document.addEventListener("keydown", down);
        return () => document.removeEventListener("keydown", down);
    }, []);

    return (
        <div className="flex h-screen bg-background">
            {/* Desktop Sidebar */}
            <aside className="hidden lg:flex lg:flex-col lg:w-64 lg:border-r border-border bg-card/50">
                {/* Logo */}
                <div className="flex h-16 items-center gap-2 px-6 border-b border-border">
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
                        <span className="text-lg font-bold text-primary-foreground">A</span>
                    </div>
                    <span className="text-xl font-bold">Antigravity</span>
                </div>

                {/* Navigation */}
                <nav className="flex-1 overflow-y-auto p-4">
                    <NavigationLinks />
                </nav>

                {/* User Section */}
                <div className="border-t border-border p-4">
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="w-full justify-start gap-2 h-auto py-2">
                                <Avatar className="h-8 w-8">
                                    <AvatarImage src={user?.avatar} />
                                    <AvatarFallback>
                                        {user?.name?.charAt(0).toUpperCase() || "U"}
                                    </AvatarFallback>
                                </Avatar>
                                <div className="flex flex-1 flex-col items-start text-sm overflow-hidden">
                                    <span className="font-medium truncate w-full text-left">{user?.name}</span>
                                    <span className="text-xs text-muted-foreground truncate w-full text-left">{user?.email}</span>
                                </div>
                                <ChevronDown className="h-4 w-4 text-muted-foreground shrink-0" />
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-56">
                            <DropdownMenuLabel>My Account</DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem asChild>
                                <Link to="/settings">
                                    <Settings className="mr-2 h-4 w-4" />
                                    Settings
                                </Link>
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={(e) => {
                                e.preventDefault();
                                setTheme(theme === "dark" ? "light" : "dark");
                            }}>
                                {theme === "dark" ? (
                                    <>
                                        <Sun className="mr-2 h-4 w-4" />
                                        Light Mode
                                    </>
                                ) : (
                                    <>
                                        <Moon className="mr-2 h-4 w-4" />
                                        Dark Mode
                                    </>
                                )}
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem onClick={() => logout()}>
                                <LogOut className="mr-2 h-4 w-4" />
                                Logout
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>
            </aside>

            {/* Main Content */}
            <div className="flex flex-1 flex-col overflow-hidden">
                {/* Header */}
                <header className="flex h-16 items-center gap-4 border-b border-border bg-background px-4 lg:px-6">
                    {/* Mobile Menu Toggle */}
                    <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
                        <SheetTrigger asChild>
                            <Button variant="ghost" size="icon" className="lg:hidden">
                                <Menu className="h-5 w-5" />
                                <span className="sr-only">Toggle menu</span>
                            </Button>
                        </SheetTrigger>
                        <SheetContent side="left" className="w-64 p-0">
                            <SheetHeader className="flex h-16 items-center gap-2 px-6 border-b border-border">
                                <div className="flex items-center gap-2">
                                    <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
                                        <span className="text-lg font-bold text-primary-foreground">A</span>
                                    </div>
                                    <SheetTitle className="text-xl font-bold">Antigravity</SheetTitle>
                                </div>
                            </SheetHeader>
                            <nav className="flex-1 overflow-y-auto p-4">
                                <NavigationLinks onNavigate={() => setMobileMenuOpen(false)} />
                            </nav>
                            {/* Mobile user section */}
                            <div className="absolute bottom-0 left-0 right-0 border-t border-border p-4 bg-background">
                                <div className="flex items-center gap-2">
                                    <Avatar className="h-8 w-8">
                                        <AvatarImage src={user?.avatar} />
                                        <AvatarFallback>
                                            {user?.name?.charAt(0).toUpperCase() || "U"}
                                        </AvatarFallback>
                                    </Avatar>
                                    <div className="flex-1 text-sm overflow-hidden">
                                        <p className="font-medium truncate">{user?.name}</p>
                                        <p className="text-xs text-muted-foreground truncate">{user?.email}</p>
                                    </div>
                                    <Button variant="ghost" size="icon" onClick={() => logout()}>
                                        <LogOut className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                        </SheetContent>
                    </Sheet>

                    {/* Search */}
                    <Button
                        variant="outline"
                        className="w-full max-w-64 justify-start text-muted-foreground hidden sm:flex"
                        onClick={() => setCommandOpen(true)}
                    >
                        <Search className="mr-2 h-4 w-4" />
                        Search...
                        <kbd className="ml-auto pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
                            <span className="text-xs">⌘</span>K
                        </kbd>
                    </Button>

                    {/* Mobile search icon */}
                    <Button
                        variant="ghost"
                        size="icon"
                        className="sm:hidden"
                        onClick={() => setCommandOpen(true)}
                    >
                        <Search className="h-5 w-5" />
                    </Button>

                    <div className="flex-1" />

                    {/* Settings Button */}
                    <Button variant="ghost" size="icon" asChild>
                        <Link to="/settings">
                            <Settings className="h-5 w-5" />
                        </Link>
                    </Button>

                    {/* Notifications */}
                    <Button variant="ghost" size="icon">
                        <Bell className="h-5 w-5" />
                    </Button>

                    {/* User Menu */}
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                                <Avatar className="h-8 w-8">
                                    <AvatarImage src={user?.avatar} alt={user?.name} />
                                    <AvatarFallback>
                                        {user?.name?.charAt(0).toUpperCase() || "U"}
                                    </AvatarFallback>
                                </Avatar>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent className="w-56" align="end" forceMount>
                            <DropdownMenuLabel className="font-normal">
                                <div className="flex flex-col space-y-1">
                                    <p className="text-sm font-medium leading-none">{user?.name || "User"}</p>
                                    <p className="text-xs leading-none text-muted-foreground">
                                        {user?.email || "user@example.com"}
                                    </p>
                                </div>
                            </DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem asChild>
                                <Link to="/settings">
                                    <Settings className="mr-2 h-4 w-4" />
                                    Settings
                                </Link>
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={(e) => {
                                e.preventDefault();
                                setTheme(theme === "dark" ? "light" : "dark");
                            }}>
                                {theme === "dark" ? (
                                    <>
                                        <Sun className="mr-2 h-4 w-4" />
                                        Light Mode
                                    </>
                                ) : (
                                    <>
                                        <Moon className="mr-2 h-4 w-4" />
                                        Dark Mode
                                    </>
                                )}
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem asChild>
                                <Link to="/login">
                                    <LogOut className="mr-2 h-4 w-4" />
                                    Login / Sign Up
                                </Link>
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => logout()} className="text-red-500">
                                <LogOut className="mr-2 h-4 w-4" />
                                Logout
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </header>

                {/* Page Content */}
                <main className="flex-1 overflow-y-auto p-4 lg:p-6 bg-muted/20">
                    {children || <Outlet />}
                </main>
            </div>

            {/* Command Palette */}
            <CommandDialog open={commandOpen} onOpenChange={setCommandOpen} />
            <Toaster />
        </div>
    );
};
