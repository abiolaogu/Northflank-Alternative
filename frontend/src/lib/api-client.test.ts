import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient, ApiError } from '@/lib/api-client';

describe('apiClient', () => {
    beforeEach(() => {
        vi.resetAllMocks();
        (globalThis as any).fetch = vi.fn();
    });

    describe('get', () => {
        it('should make GET request with auth header', async () => {
            const mockResponse = { data: { id: 1, name: 'Test' } };
            ((globalThis as any).fetch as any).mockResolvedValueOnce({
                ok: true,
                status: 200,
                json: () => Promise.resolve(mockResponse.data),
            });

            (window.localStorage.getItem as any).mockReturnValue('test-token');

            const result = await apiClient.get('/api/test');

            expect((globalThis as any).fetch).toHaveBeenCalledWith(
                expect.stringContaining('/api/test'),
                expect.objectContaining({
                    method: 'GET',
                    headers: expect.objectContaining({
                        Authorization: 'Bearer test-token',
                    }),
                })
            );
            expect(result.data).toEqual(mockResponse.data);
            expect(result.ok).toBe(true);
        });

        it('should throw ApiError on non-ok response', async () => {
            ((globalThis as any).fetch as any).mockResolvedValueOnce({
                ok: false,
                status: 401,
                json: () => Promise.resolve({ message: 'Unauthorized' }),
            });

            await expect(apiClient.get('/api/test')).rejects.toThrow(ApiError);
        });
    });

    describe('post', () => {
        it('should make POST request with JSON body', async () => {
            const payload = { name: 'New Item' };
            ((globalThis as any).fetch as any).mockResolvedValueOnce({
                ok: true,
                status: 201,
                json: () => Promise.resolve({ id: 1, ...payload }),
            });

            await apiClient.post('/api/items', payload);

            expect((globalThis as any).fetch).toHaveBeenCalledWith(
                expect.stringContaining('/api/items'),
                expect.objectContaining({
                    method: 'POST',
                    headers: expect.objectContaining({
                        'Content-Type': 'application/json',
                    }),
                    body: JSON.stringify(payload),
                })
            );
        });
    });
});
