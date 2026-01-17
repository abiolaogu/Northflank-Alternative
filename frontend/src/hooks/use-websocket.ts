import { useEffect, useRef, useState } from 'react';

export function useWebSocket(url: string) {
    const ws = useRef<WebSocket | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const [lastMessage, setLastMessage] = useState<any>(null);

    useEffect(() => {
        if (!url) return;

        ws.current = new WebSocket(url);

        ws.current.onopen = () => {
            setIsConnected(true);
            console.log('WS Connected');
        };

        ws.current.onclose = () => {
            setIsConnected(false);
            console.log('WS Disconnected');
        };

        ws.current.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                setLastMessage(data);
            } catch (e) {
                console.error('WS Parse Error', e);
            }
        };

        return () => {
            ws.current?.close();
        };
    }, [url]);

    const sendMessage = (msg: any) => {
        if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify(msg));
        }
    };

    return { isConnected, lastMessage, sendMessage };
}
