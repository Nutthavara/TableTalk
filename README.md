# TableTalk — ระบบจองโต๊ะและรีวิวร้านอาหาร

TableTalk คือเว็บสำหรับ**ค้นหาร้านอาหาร จองโต๊ะ และรีวิวร้าน**

- ทุกคนที่สมัครสมาชิกแล้ว **เปิดร้านของตัวเองได้** และ **จองหรือรีวิวร้านของคนอื่นได้** ในบัญชีเดียว ไม่ต้องสมัครแยก
- ระบบไม่มีการแบ่งบทบาท (role) ใครสร้างร้านไหน คนนั้นคือเจ้าของร้านนั้น

---

## ระบบประกอบด้วยอะไรบ้าง

| ส่วน | ทำหน้าที่ | เทคโนโลยี |
|---|---|---|
| **หน้าเว็บ** (`frontend/`) | หน้าจอทั้งหมดที่ผู้ใช้เห็น | Next.js 16 |
| **API** (`backend/`) | ตรวจกฎทุกอย่าง เช่น ที่นั่งพอไหม ใครแก้อะไรได้ | Go |
| **ฐานข้อมูล** | เก็บข้อมูลร้าน การจอง และรีวิว | PostgreSQL 18 |
| **ระบบล็อกอิน** | หน้า login, สมัครสมาชิก, ลืมรหัสผ่าน | Keycloak 26 |
| **กล่องอีเมลจำลอง** | รับอีเมลรีเซ็ตรหัสผ่านตอนพัฒนา (ไม่ส่งออกจริง) | Mailpit |

```
Browser ──► หน้าเว็บ (Next.js) ──► API (Go) ──► ฐานข้อมูล (PostgreSQL)
                 │                     │
                 └──── Keycloak ◄──────┘   (login และตรวจสอบตัวตน)
```

---

## ต้องติดตั้ง

| โปรแกรม | เวอร์ชัน | หมายเหตุ |
|---|---|---|
| Docker + Docker Compose | ใหม่ | |
| Go | 1.26.5 ขึ้นไป | |
| Node.js | 20.9 ขึ้นไป | |
| make | เวอร์ชันไหนก็ได้ | |

---

## เปิดระบบครั้งแรก (ทำครั้งเดียว)

```bash
# 1. สร้างไฟล์ตั้งค่า
cp .env.example .env
cp frontend/.env.example frontend/.env.local

# 2. สร้าง secret password แล้วนำไปใส่ใน AUTH_SECRET ของไฟล์ frontend/.env.local
openssl rand -base64 33

# 3. ติดตั้งแพ็กเกจของหน้าเว็บ
cd frontend && npm install && cd ..
```

## เปิดระบบ (ทำทุกครั้ง)

ใช้ terminal 2 หน้าต่าง โดยรันที่โฟลเดอร์หลักของโปรเจกต์

```bash
# เปิดฐานข้อมูล, Keycloak และ Mailpit (Keycloak ใช้เวลาเริ่มประมาณ 20–40 วินาที)
make up
make migrate        # อัปเดตโครงสร้างฐานข้อมูล (รันซ้ำได้ ไม่มีผลเสีย)

# terminal ที่ 1
make api            # เปิด API

# terminal ที่ 2
make web            # เปิดหน้าเว็บ
```

จากนั้นเปิด **http://localhost:3000**

**บัญชีทดลอง:** `bob` / `bob` (และ `alice` / `alice`) หรือกด **Sign up** เพื่อสมัครบัญชีใหม่ได้

> บัญชีทดลองมีไว้สำหรับตอนพัฒนาเท่านั้น ห้ามใช้ใน production

## ที่อยู่ที่ใช้บ่อย

| อะไร | ที่อยู่ |
|---|---|
| หน้าเว็บ | http://localhost:3000 |
| API | http://localhost:8081 |
| เอกสาร API (Swagger) | http://localhost:8081/docs |
| กล่องอีเมลจำลอง (Mailpit) | http://localhost:8025 |
| หน้าผู้ดูแล Keycloak | http://localhost:8080/admin (`admin` / `admin`) |
| ฐานข้อมูล | `localhost:5432` ชื่อ DB, user และรหัสผ่านเป็น `tabletalk` ทั้งหมด (ใช้กับ DBeaver ได้) |

## ปิดระบบ

| ต้องการ | คำสั่ง |
|---|---|
| ปิด API หรือหน้าเว็บ | กด `Ctrl+C` ใน terminal นั้น |
| ปิดฐานข้อมูลและ Keycloak (ข้อมูลยังอยู่) | `make down` |
| **ลบข้อมูลทั้งหมด** แล้วเริ่มใหม่ | `make reset-infra` แล้วตามด้วย `make up` และ `make migrate` |

---

## ใช้งานระบบได้อะไรบ้าง

| หน้า | ทำอะไรได้ |
|---|---|
| `/restaurants` | ดูรายการร้าน ค้นหาด้วยชื่อ เรียงตามคะแนนหรือจำนวนรีวิว |
| `/restaurants/[id]` | ดูรายละเอียดร้าน จองโต๊ะ อ่านและเขียนรีวิว (ถ้าเป็นร้านของเราเอง จะเห็นปุ่มจัดการร้านแทน) |
| `/my-reservations` | การจองของเรา แยกเป็น กำลังจะถึง / ผ่านไปแล้ว / ยกเลิกแล้ว |
| `/my-reservations/[id]/edit` | แก้ไขการจอง |
| `/my-restaurants` | ร้านที่เราเป็นเจ้าของ (ดู แก้ไข ลบ) |
| `/my-restaurants/new` | เพิ่มร้านใหม่ |

**ลืมรหัสผ่าน:** กด "Forgot password?" ในหน้า login ตอนพัฒนาอีเมลรีเซ็ตรหัสจะไปอยู่ที่ Mailpit (http://localhost:8025) ไม่ได้ส่งออกไปจริง

---

## กฎสำคัญของระบบ

### การจอง
- **capacity คือจำนวนคนที่ร้านรับได้พร้อมกัน** ไม่ได้นับเป็นโต๊ะ
- จองได้เฉพาะ**เวลาในอนาคต** และต้อง**อยู่ในเวลาเปิดร้านทั้งช่วง**
- ระบบเช็กว่า**ทุกนาที**ในช่วงที่จอง จำนวนคนรวมต้องไม่เกิน capacity
  - ตัวอย่าง: ร้านรับได้ 10 คน มีคนจอง 17:00–17:30 ไว้ 6 คน และ 18:00–18:30 อีก 6 คน ถ้าเราจอง 17:00–19:00 จำนวน 4 คน **จองได้** เพราะไม่มีช่วงไหนเกิน 10 คนเลย
- การจองที่เวลาต่อกันพอดี (เช่น 17:00–18:00 กับ 18:00–19:00) **ไม่นับว่าทับกัน**
- ถ้ามีคนกดจองที่นั่งสุดท้ายพร้อมกัน ระบบให้จองได้แค่คนเดียว ไม่มีทางจองเกิน
- **แก้ไขหรือยกเลิกการจองได้ก่อนเส้นตายเท่านั้น** เส้นตายคือเวลาเริ่มจองลบด้วยจำนวนชั่วโมงที่ร้านกำหนด (อย่างน้อย 2 ชั่วโมง) ตัวอย่าง: จองไว้ 18:00 และร้านกำหนด 2 ชั่วโมง → ต้องแก้หรือยกเลิกภายใน 16:00
- การยกเลิก**ไม่ได้ลบข้อมูล** แค่เปลี่ยนสถานะเป็น "ยกเลิกแล้ว"

### ร้านอาหาร
- ร้านต้องมี**รูปอย่างน้อย 1 รูป** (สูงสุด 10 รูป รูปละไม่เกิน 5 MB รองรับ JPEG, PNG และ WebP)
- เฉพาะ**เจ้าของร้าน**เท่านั้นที่แก้ไขหรือลบร้านได้
- เปลี่ยน capacity หรือเวลาเปิดปิดไม่ได้ ถ้าจะทำให้การจองที่มีอยู่แล้วผิดกฎ
- ลบร้านไม่ได้ ถ้ายังมีการจองที่ยังไม่ถึงเวลาหรือยังไม่จบ
- ยังไม่รองรับร้านที่เปิดข้ามเที่ยงคืน

### รีวิว
- ให้คะแนนเป็นจำนวนเต็ม **1–5 ดาว** ความคิดเห็นจะเขียนหรือไม่ก็ได้
- **1 คนรีวิวได้ 1 ครั้งต่อร้าน** (แก้ไขหรือลบรีวิวของตัวเองได้)
- **เจ้าของร้านรีวิวร้านตัวเองไม่ได้**
- ตอนนี้ยังไม่บังคับว่าต้องเคยจองร้านนั้นก่อนถึงจะรีวิวได้
- คะแนนเฉลี่ยแสดงทศนิยม 1 ตำแหน่ง ร้านที่ยังไม่มีรีวิวจะแสดงว่า "No reviews yet" ไม่ใช่ 0 ดาว

### เวลา
- ทุกร้านมี timezone ของตัวเอง (ค่าเริ่มต้นคือ `Asia/Bangkok`) เวลาเปิดปิดและเวลาจองทั้งหมดคิดตาม timezone ของร้าน

---

## เอกสาร API (Swagger)

เปิด **http://localhost:8081/docs** จะเห็นรายการ API ทั้งหมด และลองยิงได้จากหน้านั้นเลย

**วิธีลองยิง API ที่ต้อง login:**

1. ขอ token แล้ว copy ไว้ใน clipboard:
   ```bash
   curl -s -d grant_type=password -d client_id=tabletalk-dev-cli \
     -d username=bob -d password=bob \
     http://localhost:8080/realms/tabletalk/protocol/openid-connect/token | jq -r .access_token | pbcopy
   ```
2. ในหน้า Swagger กด **Authorize** วาง token (ไม่ต้องพิมพ์คำว่า `Bearer` นำหน้า) กด **Authorize** แล้วกด **Close**
3. เลือก API ที่ต้องการ กด **Try it out** แล้วกด **Execute**

> token มีอายุ **5 นาที** ถ้าได้ error 401 ให้ขอ token ใหม่
>
> หน้า Swagger เปิดเฉพาะเมื่อตั้ง `API_DOCS_ENABLED=true` ใน `.env` ซึ่ง `.env.example` ตั้งไว้ให้แล้ว ไฟล์ spec อยู่ที่ `backend/api/swagger.yaml`

**รูปแบบ error** ทุกตัวหน้าตาเหมือนกัน:

```json
{ "error": { "code": "INSUFFICIENT_CAPACITY", "message": "Not enough seats are available for the selected time." } }
```

ถ้ากรอกข้อมูลผิด (`VALIDATION_ERROR`) จะมี `details` บอกด้วยว่าช่องไหนผิด

---

## การทดสอบ (Tests)

ต้องเปิดฐานข้อมูลไว้ก่อน (`make up`)

```bash
make test     # รัน test ทั้งหมดของ API
make lint     # ตรวจรูปแบบโค้ด (Go และหน้าเว็บ)
```

> ใช้ `make test` เสมอ ถ้ารัน `go test` เองโดยไม่ได้ตั้ง `TEST_DATABASE_URL` test ที่ใช้ฐานข้อมูลจะถูก**ข้ามไปเฉย ๆ** โดยไม่ fail ทำให้ดูเหมือนผ่าน

test ครอบคลุมกฎสำคัญทั้งหมด เช่น เวลาเปิดปิด, ห้ามจองย้อนหลัง, ที่นั่งไม่พอ, จองพร้อมกันหลายคน, เส้นตายการยกเลิก, เจ้าของห้ามรีวิวร้านตัวเอง และแก้ข้อมูลของคนอื่นไม่ได้ รายละเอียดอยู่ใน [ARCHITECTURE.md](ARCHITECTURE.md#testing-strategy)

---

## ค่าตั้งค่า (Environment variables)

### `.env` (โฟลเดอร์หลัก ใช้กับ Docker และ API)

| ตัวแปร | ค่าเริ่มต้น | ใช้ทำอะไร |
|---|---|---|
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | `tabletalk` | user, รหัสผ่าน และชื่อฐานข้อมูล |
| `POSTGRES_PORT` | `5432` | port ของฐานข้อมูล |
| `KEYCLOAK_PORT` | `8080` | port ของ Keycloak |
| `KEYCLOAK_ADMIN` / `KEYCLOAK_ADMIN_PASSWORD` | `admin` | บัญชีผู้ดูแล Keycloak |
| `MAILPIT_UI_PORT` | `8025` | port ของกล่องอีเมลจำลอง |
| `HTTP_ADDR` | `:8081` | port ของ API |
| `DATABASE_URL` | — | ที่อยู่ฐานข้อมูลที่ API ใช้ (**ต้องมี**) |
| `TEST_DATABASE_URL` | — | ฐานข้อมูลสำหรับ test |
| `OIDC_ISSUER_URL` | — | ที่อยู่ของ Keycloak realm (**ต้องมี** และต้องตรงกับใน token ทุกตัวอักษร) |
| `OIDC_AUDIENCE` | `tabletalk-api` | ชื่อที่ token ต้องระบุว่าใช้กับ API นี้ |
| `UPLOAD_DIR` | `./uploads` | โฟลเดอร์เก็บรูปร้าน |
| `PUBLIC_BASE_URL` | `http://localhost:8081` | ใช้สร้างลิงก์รูปภาพ |
| `API_DOCS_ENABLED` | `false` | `true` เพื่อเปิดหน้า Swagger ที่ `/docs` |

> ถ้าเปลี่ยน `KEYCLOAK_PORT` ต้องแก้ `OIDC_ISSUER_URL` ในไฟล์นี้ และ `AUTH_KEYCLOAK_ISSUER` ของหน้าเว็บให้ตรงกันด้วย

### `frontend/.env.local` (หน้าเว็บ)

| ตัวแปร | ตัวอย่าง | ใช้ทำอะไร |
|---|---|---|
| `AUTH_SECRET` | ค่าสุ่ม (สร้างด้วย `openssl rand -base64 33`) | เข้ารหัส cookie ของการ login |
| `AUTH_URL` | `http://localhost:3000` | ที่อยู่ของหน้าเว็บ |
| `AUTH_KEYCLOAK_ID` | `tabletalk-web` | ชื่อ client ใน Keycloak |
| `AUTH_KEYCLOAK_SECRET` | `dev-only-tabletalk-web-secret` | รหัสของ client (**ใช้เฉพาะตอนพัฒนา**) |
| `AUTH_KEYCLOAK_ISSUER` | `http://localhost:8080/realms/tabletalk` | ที่อยู่ของ Keycloak realm |
| `API_BASE_URL` | `http://localhost:8081` | ที่อยู่ของ API |

> ไฟล์ `.env` และ `.env.local` **ไม่ถูก commit** ขึ้น git รหัสลับทุกตัวใน repo นี้เป็นค่าสำหรับพัฒนาเท่านั้น ห้ามใส่รหัสจริงลงใน repo

---

## โครงสร้างโฟลเดอร์

```
backend/                 API (Go)
  api/swagger.yaml       เอกสาร API
  cmd/api                จุดเริ่มต้นของ API
  cmd/migrate            ตัวรัน migration
  internal/handler       รับ request และส่ง response
  internal/service       กฎทางธุรกิจทั้งหมด (การจอง, ร้าน, รีวิว)
  internal/repository    คำสั่ง SQL ทั้งหมด
  internal/auth          ตรวจ token จาก Keycloak
  internal/storage       เก็บรูปภาพ
  migrations             โครงสร้างฐานข้อมูล
frontend/                หน้าเว็บ (Next.js)
  app/                   หน้าต่าง ๆ ของเว็บ
  features/              โค้ดแยกตามเรื่อง (ร้าน, การจอง, รีวิว)
  components/            ชิ้นส่วนหน้าจอที่ใช้ซ้ำ
keycloak/                ตั้งค่า realm และ theme ของหน้า login
docker-compose.yml       ฐานข้อมูล, Keycloak, Mailpit
```

---

## อ่านเพิ่มเติม

- **[ARCHITECTURE.md](ARCHITECTURE.md):** ระบบออกแบบอย่างไรและทำไม (algorithm การจอง, การกันจองซ้อน, data model และการตัดสินใจต่าง ๆ)
- **[AGENTS.md](AGENTS.md):** คู่มือสำหรับคนหรือ AI ที่จะมาพัฒนาต่อ (conventions, กฎที่ห้ามทำพัง และปัญหาที่เคยเจอ)
