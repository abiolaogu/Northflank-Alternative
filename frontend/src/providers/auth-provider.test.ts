import { describe, it, expect, vi, beforeEach } from 'vitest';
import { authProvider } from '@/providers/auth-provider';

describe('authProvider', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        (window.localStorage.setItem as any).mockClear();
        (window.localStorage.removeItem as any).mockClear();
    });

    describe('login', () => {
        it('should login with demo credentials', async () => {
            const result = await authProvider.login({
                email: 'demo@antigravity.io',
                password: 'demo',
            });

            expect(result.success).toBe(true);
            expect(result.redirectTo).toBe('/');
            expect(window.localStorage.setItem).toHaveBeenCalledWith(
                'auth_token',
                'demo-token'
            );
        });

        it('should fail with invalid credentials', async () => {
            (globalThis as any).fetch = vi.fn().mockResolvedValueOnce({
                ok: false,
            });

            const result = await authProvider.login({
                email: 'wrong@email.com',
                password: 'wrong',
            });

            expect(result.success).toBe(false);
        });
    });

    describe('logout', () => {
        it('should clear auth data and redirect to login', async () => {
            const result = await authProvider.logout({});

            expect(result.success).toBe(true);
            expect(result.redirectTo).toBe('/login');
            expect(window.localStorage.removeItem).toHaveBeenCalledWith('auth_token');
            expect(window.localStorage.removeItem).toHaveBeenCalledWith('user');
        });
    });

    describe('check', () => {
        it('should return authenticated when token exists', async () => {
            (window.localStorage.getItem as any).mockReturnValue('valid-token');

            const result = await authProvider.check();

            expect(result.authenticated).toBe(true);
        });

        it('should return not authenticated when no token', async () => {
            (window.localStorage.getItem as any).mockReturnValue(null);

            const result = await authProvider.check();

            expect(result.authenticated).toBe(false);
            expect(result.redirectTo).toBe('/login');
        });
    });
});
