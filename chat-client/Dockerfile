# chat-client/Dockerfile

# nginx로 정적 파일 서빙
# npx serve 대신 nginx 사용 — 실제 서비스에서 쓰는 방식
FROM nginx:alpine

# nginx 기본 설정 제거 후 우리 설정으로 교체
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 정적 파일 복사
COPY . /usr/share/nginx/html

EXPOSE 80
