import { AccessControlProvider } from "@refinedev/core";

export const accessControlProvider: AccessControlProvider = {
    can: async ({ resource, action, params }) => {
        const userString = localStorage.getItem("user");
        if (!userString) {
            return { can: false, reason: "Unauthorized" };
        }

        const user = JSON.parse(userString);

        // Admin has full access
        if (user.role === "admin") {
            return { can: true };
        }

        // Role-based logic
        if (user.role === "editor" && action === "delete") {
            return { can: false, reason: "Editors cannot delete resources" };
        }

        if (user.role === "viewer" && action !== "list" && action !== "show") {
            return { can: false, reason: "Viewers can only view resources" };
        }

        return { can: true };
    },
};
