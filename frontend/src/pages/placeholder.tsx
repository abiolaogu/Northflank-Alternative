import { Construction } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Link, useLocation } from "react-router-dom";

export const PlaceholderPage = () => {
    const location = useLocation();

    return (
        <div className="flex flex-col items-center justify-center h-[60vh] text-center space-y-6">
            <div className="p-6 bg-muted rounded-full">
                <Construction className="h-12 w-12 text-muted-foreground" />
            </div>
            <div className="space-y-2">
                <h1 className="text-3xl font-bold tracking-tight">Under Construction</h1>
                <p className="text-muted-foreground max-w-md mx-auto">
                    The <span className="font-semibold text-foreground capitalize">{location.pathname.split('/')[1]}</span> module is currently being implemented. Check back soon!
                </p>
            </div>
            <Button asChild variant="default">
                <Link to="/">Back to Dashboard</Link>
            </Button>
        </div>
    );
};
