// js/api.js
import { CONFIG } from './config.js';
const BASE_URL = CONFIG.API_BASE_URL; // 'http://localhost:3000'


// 토큰 저장/조회 — localStorage
export const getToken = () => localStorage.getItem('token');
export const setToken = (t) => localStorage.setItem('token', t);
export const getUser  = () => JSON.parse(localStorage.getItem('user') || 'null');
export const setUser  = (u) => localStorage.setItem('user', JSON.stringify(u));
export const clearAuth = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
};

// 인증 헤더 포함 fetch
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
        const err = await res.json();
        throw new Error(err.error || '요청 실패');
    }
    return res.json();
}