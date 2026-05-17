# Contributing — 에이전트 행동 강령

> 이 파일은 **Claude Code를 포함한 모든 AI 에이전트**가 이 레포지토리에서 작업할 때 반드시 따라야 하는 규칙입니다.  
> 인간 기여자도 동일한 규칙을 따릅니다.

---

## 에이전트 필수 행동 순서

### 작업 시작 전

1. **작업 대상 폴더의 `README.md`를 읽는다**
   - `Agent Control` 섹션의 허용/금지/필수 규칙을 숙지한다
   - `DFD`를 보고 이 폴더의 데이터 흐름을 파악한다
   - `Progress Tracker`를 보고 이미 구현된 기능과 미구현 기능을 확인한다

2. **루트 `BLUEPRINT.md`를 참조한다** (첫 세션 또는 구조 변경 시)

3. **수정할 기능의 `Progress Tracker` 상태를 `🔄 In Progress`로 변경한다**

### 작업 중

4. **`Agent Control`의 금지 규칙을 코드 생성 전 한 번 더 확인한다**

5. **새 파일 생성 시, 해당 폴더의 `README.md`가 없으면 `templates/README_TEMPLATE.md`를 기반으로 생성한다**

### 작업 완료 후

6. **`Progress Tracker`를 업데이트한다**
   - 완료된 기능: `✅ Done` + `Last Updated` 오늘 날짜
   - 새로 발견된 작업: 새 행 추가 (`⏳ Pending`)

7. **`Next Roadmap`을 갱신한다** (다음으로 해야 할 작업 순서 명시)

8. **구조적 변경이 발생했다면 `DFD`를 업데이트한다**
   - 새 외부 의존성 추가
   - 새 데이터 흐름 경로 생성
   - 기존 흐름 변경

---

## 절대 금지 행동

| 금지 행동 | 이유 |
|-----------|------|
| `Agent Control`의 금지 규칙 위반 | 아키텍처 원칙 붕괴 |
| `Progress Tracker` 미업데이트 작업 완료 | 맥락 정보 소실 |
| `DFD` 미업데이트 구조 변경 | 에이전트/개발자 오해 유발 |
| 폴더 `README.md` 미확인 작업 시작 | 중복 구현, 규칙 위반 |
| 테스트 없이 `✅ Done` 마킹 | 허위 진행 현황 |

---

## 문서 업데이트 기준

### DFD를 업데이트해야 하는 경우
- 새로운 외부 API 또는 DB 연결 추가
- 데이터 처리 순서 변경
- 새 모듈/서비스 간 의존성 생성
- 기존 흐름 제거 또는 우회

### Progress Tracker를 업데이트해야 하는 경우
- 작업 시작 시 (`🔄 In Progress`)
- 작업 완료 시 (`✅ Done`)
- 블로커 발견 시 (`❌ Blocked` + Notes에 이유 기재)
- 새 작업 항목 발견 시 (새 행 추가, `⏳ Pending`)

### README.md를 새로 생성해야 하는 경우
- 새 폴더 생성 시
- `templates/README_TEMPLATE.md`를 복사해 해당 폴더 맥락에 맞게 작성

---

## 커밋 메시지 규칙

```
<type>(<scope>): <description>

[optional body]
[optional footer]
```

| Type | 사용 시점 |
|------|-----------|
| `feat` | 새 기능 추가 |
| `fix` | 버그 수정 |
| `docs` | 문서만 수정 (README, DFD, Progress Tracker 등) |
| `refactor` | 기능 변경 없는 코드 구조 개선 |
| `test` | 테스트 추가/수정 |
| `chore` | 빌드, 의존성 등 기타 변경 |

**에이전트는 Progress Tracker 업데이트를 별도 `docs` 커밋으로 분리하거나,  
기능 커밋에 포함해 메시지에 명시합니다.**

예: `feat(auth): add refresh token endpoint + update progress tracker`
