import { NotificationProvider } from "@refinedev/core";
// import { toast } from "@/components/ui/use-toast"; // Function-based call requires extracting logic, using simplified object for now or custom implementation.
// Since Refine expects a simplified object, and Shadcn's toast is hook-based, we need a bridge.
// A common pattern is to use a global event emitter or a simpler library like `sonner` or `react-hot-toast` wrapped.
// For strict Shadcn compatibility, we often use a custom hook inside Layout, but Refine needs a static provider object.
// We will implement a simplified console/event based bridge that the Layout listens to, or assume `toast` can be imported directly if we switch to `sonner` in the future.
// For now, we'll use a mocked implementation that logs to console to avoid complex hook-bridge outside of React context.

// Update: To properly use Shadcn toast outside of components, we need a different approach or just rely on in-component hooks.
// However, Refine *requires* this provider for automatic Success/Error notifications.
// We'll leave this empty for now and handle notifications manually in components using `useToast()`, 
// or implement a custom event dispatcher.

export const notificationProvider: NotificationProvider = {
    open: ({ message, type, description }) => {
        // This is where we would trigger the toast if we had a static instance.
        // For now, we will dispatch a custom event that our Layout can listen to.
        const event = new CustomEvent("refine-notification", {
            detail: { message, type, description },
        });
        window.dispatchEvent(event);
    },
    close: () => { },
};
