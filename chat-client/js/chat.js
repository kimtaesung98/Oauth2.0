// js/chat.js
import { apiFetch, getToken, getUser } from './api.js';

let ws = null;
let currentPartner = null;

// ─────────────────────────────────────────
// WebSocket 연결
// ─────────────────────────────────────────
export function connectWS(onMessage) {
    const token = getToken();
    ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

    ws.onopen = () => console.log('WebSocket 연결됨');

    ws.onmessage = (e) => {
        const data = JSON.parse(e.data);

        // 온라인 상태 업데이트
        if (data.type === 'online_users') {
            updateOnlineStatus(data.onlineUsers);
            return;
        }
        // 메시지 수신
        onMessage(data);
    };

    ws.onclose = () => console.log('WebSocket 연결 끊김');
}

// ─────────────────────────────────────────
// 메시지 전송
// ─────────────────────────────────────────
export function sendMessage(receiverId, content) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ receiverId, content }));
}

// ─────────────────────────────────────────
// 채팅 목록 불러오기
// ─────────────────────────────────────────
export async function loadChatList() {
    return await apiFetch('/chats');
}

// ─────────────────────────────────────────
// 메시지 내역 불러오기
// ─────────────────────────────────────────
export async function loadMessages(partnerId) {
    return await apiFetch(`/messages/${partnerId}`);
}

// 온라인 상태 UI 업데이트
function updateOnlineStatus(onlineIds) {
    document.querySelectorAll('.user-item').forEach(el => {
        const uid = Number(el.dataset.uid);
        const dot = el.querySelector('.dot');
        if (dot) dot.className = `dot ${onlineIds.includes(uid) ? 'online' : 'offline'}`;
    });
}