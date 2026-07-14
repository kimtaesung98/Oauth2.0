// js/chat.js
import { apiFetch, getToken, getUser } from './api.js';

let ws = null;
let currentPartnerId = null;

// ─────────────────────────────────────────
// WebSocket 연결
// ─────────────────────────────────────────
export function connectWS(onMessage, onOnlineUpdate) {
    const token = getToken();
    if (!token) return;

    ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

    ws.onopen = () => console.log('WebSocket 연결됨');

    ws.onmessage = (e) => {
        const data = JSON.parse(e.data);
        if (data.type === 'online_users') {
            onOnlineUpdate(data.onlineUsers);
            return;
        }
        onMessage(data);
    };

    ws.onclose = () => {
        console.log('WebSocket 연결 끊김 — 3초 후 재연결');
        setTimeout(() => connectWS(onMessage, onOnlineUpdate), 3000);
    };

    ws.onerror = (e) => console.error('WebSocket 에러:', e);
}

// ─────────────────────────────────────────
// 메시지 전송
// ─────────────────────────────────────────
export function sendMessage(receiverId, content) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        console.error('WebSocket 연결 안 됨');
        return false;
    }
    ws.send(JSON.stringify({ receiverId, content }));
    return true;
}

// ─────────────────────────────────────────
// API 호출
// ─────────────────────────────────────────
export const loadChatList  = () => apiFetch('/chats');
export const loadMessages  = (partnerId) => apiFetch(`/messages/${partnerId}`);
export const markAsRead    = (senderId) => apiFetch('/messages/read', {
    method: 'PATCH',
    body: JSON.stringify({ senderId }),
});

// ─────────────────────────────────────────
// 현재 열린 채팅 파트너 ID
// ─────────────────────────────────────────
export const getCurrentPartnerId = () => currentPartnerId;
export const setCurrentPartnerId = (id) => { currentPartnerId = id; };

// ─────────────────────────────────────────
// 시간 포맷
// ─────────────────────────────────────────
export function formatTime(dateStr) {
    const d = new Date(dateStr);
    const now = new Date();
    const isToday = d.toDateString() === now.toDateString();
    if (isToday) {
        return d.toLocaleTimeString('ko-KR', { hour: '2-digit', minute: '2-digit' });
    }
    return d.toLocaleDateString('ko-KR', { month: 'short', day: 'numeric' });
}
