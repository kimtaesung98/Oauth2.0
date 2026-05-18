// js/auth.js
import { apiFetch, setToken, setUser } from './api.js';

// ─────────────────────────────────────────
// 이메일 로그인
// ─────────────────────────────────────────
export async function emailLogin(email, password) {
    const data = await apiFetch('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
    });
    setToken(data.token);
    setUser(data.user);
    return data;
}
 
// ─────────────────────────────────────────
// Google 로그인
// Google SDK가 ID Token을 주면 → Go 서버로 전달
// ─────────────────────────────────────────
export async function googleLogin(idToken) {
    const data = await apiFetch('/auth/google', {
        method: 'POST',
        body: JSON.stringify({ idToken }),
    });
    setToken(data.token);
    setUser(data.user);
    return data;
}

// Google One Tap 초기화
// GOOGLE_CLIENT_ID를 여기에 직접 넣어주세요
export function initGoogle(clientId) {
    google.accounts.id.initialize({
        client_id: clientId,
        callback: async (response) => {
            // response.credential = Google ID Token
            console.log('ID Token:', response.credential); // 임시: aud 값 확인
            try {
                await googleLogin(response.credential);
                window.location.href = '/chat.html';
            } catch (e) {
                alert('Google 로그인 실패: ' + e.message);
            }
        },
    });

    // 버튼 렌더링
    google.accounts.id.renderButton(
        document.getElementById('google-btn'),
        { theme: 'outline', size: 'large', text: 'signin_with', width: 300 }
    );
}