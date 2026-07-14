// js/api.js
import { CONFIG } from './config.js';

const BASE_URL = CONFIG.API_BASE_URL;

// ─────────────────────────────────────────
// 토큰 · 유저 관리
// ─────────────────────────────────────────
export const getToken  = () => localStorage.getItem('token');
export const setToken  = (t) => localStorage.setItem('token', t);
export const getUser   = () => JSON.parse(localStorage.getItem('user') || 'null');
export const setUser   = (u) => localStorage.setItem('user', JSON.stringify(u));
export const clearAuth = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
};

// ─────────────────────────────────────────
// 공통 fetch — 인증 헤더 자동 포함
// ─────────────────────────────────────────
export async function apiFetch(path, options = {}) {
    const token = getToken();
    const res = await fetch(BASE_URL + path, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
            ...options.headers,
        },
    });

    if (!res.ok) {
        const err = await res.json().catch(() => ({ error: '요청 실패' }));
        throw new Error(err.error || '요청 실패');
    }
    return res.json();
}

// ─────────────────────────────────────────
// 유저 검색
// GET /practice/users — 전체 유저 목록에서 검색
// ─────────────────────────────────────────
export async function searchUsers(keyword) {
    const data = await apiFetch('/practice/users');
    const me = getUser();
    // 나 자신 제외 + 키워드 필터
    return data.users.filter(u =>
        u.id !== me?.id &&
        (u.name.includes(keyword) || u.email.includes(keyword))
    );
}
