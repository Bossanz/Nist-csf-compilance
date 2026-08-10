# Slug-Based Routing Design

**Date:** 2026-08-10

**Status:** Proposed

## Goal

แยกหน้า Next.js ให้เป็น route จริง และใช้ชื่อที่อ่านได้ของ Organization กับ Project ใน browser URL แทน UUID โดยยังคงใช้ UUID เป็น primary key และ internal identifier ของระบบ

## Scope

### In scope

- แยกหน้า Login, Organization list, Organization workspace และ Project assessment เป็น Next.js App Router routes
- เพิ่ม slug สำหรับ Organization และ Project
- รองรับ slug ภาษาอังกฤษและภาษาไทย
- ป้องกัน slug ซ้ำภายในขอบเขตที่เหมาะสม
- ทำให้ refresh, browser back และ direct link กลับมาหน้าเดิมได้
- คง API ที่ใช้ UUID สำหรับ profile, responses, documents และ mutation เดิมไว้

### Out of scope

- ระบบแก้ไขชื่อ Organization หรือ Project
- ระบบ redirect จาก slug เก่าหลังเปลี่ยนชื่อ
- การเปลี่ยนรูปแบบ UUID ใน Database
- การเพิ่มระบบ router library หรือ state management library ใหม่

## URL Structure

```text
/                                           redirect ไป /organizations
/login                                      หน้าเข้าสู่ระบบ
/organizations                              รายการ Organization
/organizations/[organizationSlug]           Workspace ของ Organization
/organizations/[organizationSlug]/projects/[projectSlug]
                                            หน้า Project assessment
/invite/[token]                              หน้ารับคำเชิญเดิม
```

ตัวอย่าง URL:

```text
/organizations/acme-corporation
/organizations/acme-corporation/projects/nist-csf-readiness
/organizations/บริษัท-เอ-บี-ซี/projects/ประเมินความพร้อม-2026
```

## Slug Rules

Slug จะถูกสร้างตอนสร้างข้อมูลและคงที่ตลอดอายุของข้อมูล เพื่อไม่ให้ URL เปลี่ยนตามชื่อในอนาคต

- ตัดช่องว่างหัวท้าย
- แปลงตัวอักษรเป็น lowercase เมื่อทำได้
- คงตัวอักษร Unicode และตัวเลข เพื่อรองรับภาษาไทย
- เปลี่ยนกลุ่มช่องว่างและเครื่องหมายคั่นเป็น `-`
- รวม `-` ซ้ำให้เหลือตัวเดียว และตัด `-` ที่หัวท้าย
- ถ้าชื่อไม่มีตัวอักษรหรือตัวเลข ให้ใช้ `item`
- ถ้า slug ซ้ำ ให้เติม suffix เช่น `acme-corporation-2`

ขอบเขตความไม่ซ้ำ:

- Organization slug ต้องไม่ซ้ำทั้งระบบ
- Project slug ต้องไม่ซ้ำภายใน Organization เดียวกัน

## Data Model

เพิ่มคอลัมน์:

```sql
organizations.slug text
projects.slug text
```

เพิ่ม unique indexes:

```sql
UNIQUE (organizations.slug)
UNIQUE (projects.organization_id, projects.slug)
```

UUID ยังคงเป็น primary key และยังอยู่ใน response ของ API เพื่อให้ frontend เรียก endpoint เดิมสำหรับข้อมูล assessment ได้ แต่จะไม่ถูกใช้เป็น segment ของ browser URL

สำหรับฐานข้อมูลเดิม migration จะเพิ่มคอลัมน์แบบรองรับข้อมูลเก่า และให้ backend backfill slug ที่ยังว่างก่อนเปิดใช้งาน route ใหม่

## API Resolution

เพิ่ม lookup endpoint ที่ใช้ slug:

```text
GET /api/organizations/by-slug/:organizationSlug
GET /api/organizations/:organizationID/projects/by-slug/:projectSlug
```

ทั้งสอง endpoint ต้องผ่าน session authentication และ organization authorization ก่อนคืนข้อมูล เพื่อไม่ให้ slug กลายเป็นช่องทาง bypass สิทธิ์

หลัง resolve ได้ UUID แล้ว หน้า Project จะใช้ endpoint เดิมต่อไป:

```text
GET /api/projects/:projectID/profile
GET /api/projects/:projectID/summary
GET /api/projects/:projectID/responses
PUT /api/projects/:projectID/profile/:subcategoryID
```

## Frontend Structure

`web/src/app/page.tsx` จะทำหน้าที่ redirect เท่านั้น ส่วนแต่ละ route จะเป็นเจ้าของการโหลดข้อมูลของตัวเอง:

- `web/src/app/login/page.tsx` ตรวจ session และทำ login จากนั้น `router.replace("/organizations")`
- `web/src/app/organizations/page.tsx` โหลด session และรายการ Organization
- `web/src/app/organizations/[organizationSlug]/page.tsx` resolve Organization จาก slug แล้วโหลด projects, users และ invitation actions
- `web/src/app/organizations/[organizationSlug]/projects/[projectSlug]/page.tsx` resolve Organization/Project จาก slug แล้วโหลด profile, summary และ responses
- `web/src/app/invite/[token]/page.tsx` คง flow รับคำเชิญเดิม

Components ที่มีอยู่ เช่น `OrganizationDashboard`, `OrganizationWorkspace`, `ProfileEditor` และ `AssessmentRail` จะยังเป็น UI components รับข้อมูลและ callbacks จาก route page ไม่ย้าย business logic ไปไว้ใน component เพิ่ม

การกลับหน้าจะใช้ `Link` หรือ `router.push` ไปยัง route ที่อ่านได้ ไม่พึ่ง React state ในหน้าเดียวอีกต่อไป

## Error Handling

- ถ้า session หมดอายุหรือไม่ได้ login ให้ redirect ไป `/login`
- ถ้า slug ไม่พบ ให้แสดง not-found state และไม่เรียกข้อมูล assessment ต่อ
- ถ้า slug พบแต่ไม่มีสิทธิ์ ให้ใช้ error response เดิมของ API และไม่เปิดเผยข้อมูล Organization/Project
- ถ้าสร้างข้อมูลสำเร็จ ให้ navigate ไปยัง URL ที่สร้างจาก slug ที่ backend คืนมา

## Testing

- เพิ่ม unit test สำหรับ slug generation และ collision suffix
- เพิ่ม store/API tests สำหรับสร้างและ resolve slug
- เพิ่ม migration/backfill test สำหรับข้อมูลเดิม
- แยก page tests ตาม route และทดสอบว่า login, organization และ project navigation ใช้ URL ที่ถูกต้อง
- รัน frontend test, typecheck และ production build
- รัน Go tests ที่เกี่ยวข้องกับ store และ httpapi

## Acceptance Criteria

- เปิด `/organizations` แล้วเห็นรายการ Organization
- เปิด `/organizations/<organization-slug>` แล้วเห็น workspace ที่ถูกต้อง
- เปิด `/organizations/<organization-slug>/projects/<project-slug>` แล้วเห็น assessment ที่ถูกต้อง
- กด refresh ที่ Project แล้วไม่กลับไปหน้าแรก
- Browser back กลับจาก Project ไป Organization workspace และกลับไป Organization list ได้
- URL ไม่มี UUID ของ Organization หรือ Project
- ชื่อซ้ำไม่ทำให้ route ชนกัน
- สิทธิ์ Counselor และ Stakeholder ยังคงทำงานเหมือนเดิม
