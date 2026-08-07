# Readable White UI Design

## Goal

ปรับหน้าจอหลักของ CSF Compliance ให้เหมาะกับการอ่านข้อมูลจำนวนมากเป็นเวลานาน โดยคงโครงสร้างและฟังก์ชันเดิมไว้ ลดสิ่งรบกวนทางสายตา และทำให้ผู้ใช้รู้ได้ทันทีว่ากำลังอยู่ที่องค์กร โปรเจกต์ และ Function ใด

## Design direction

ใช้แนวทาง **Calm Editorial**: หน้าตาเป็น workspace ที่โปร่ง สงบ และมีลำดับชั้นแบบเอกสารอ่านง่าย ไม่ใช้กรอบหรือเงาหนักเกินจำเป็น

ใช้กฎสี 60/30/10:

- 60% สีขาว: พื้นที่อ่าน การ์ด ฟอร์ม และรายละเอียดของ assessment
- 30% สีเทาอ่อน/เขียวเทาอ่อน: canvas, section ย่อย, พื้นที่พักสายตา และขอบเขตของกลุ่มข้อมูล
- 10% สี teal: primary action, เมนูที่เลือก, สถานะสำคัญ และเส้นนำสายตา

Signature move คือ **reading rail**: เส้น teal แนวตั้ง/ขอบซ้ายที่ใช้เฉพาะกับบริบทปัจจุบันและรายการที่อยู่ใน scope เพื่อช่วยสแกนหน้าจอโดยไม่เพิ่ม decoration

## Scope

ปรับเฉพาะ presentation และ interaction ที่ช่วยการอ่านในหน้าหลักเหล่านี้:

- Login
- Organization dashboard และการสร้าง/ลบ organization
- Organization workspace และรายการ project
- Project dashboard, summary และ Function navigation
- Assessment cards, Current/Target profile และ stakeholder response panel

ไม่เปลี่ยน schema, API, role permission, workflow หรือ business calculation

## Layout rules

- ใช้ content column ที่มีความกว้างอ่านง่ายบน desktop และ padding ที่สม่ำเสมอ
- ให้แต่ละหน้ามี heading หลักเพียงหนึ่งจุด และลด label ที่ซ้ำซ้อน
- ใช้ white surface แยกเนื้อหาออกจาก canvas สีอ่อน
- Assessment list ยุบรายการไว้ก่อน และเปิดรายละเอียดทีละรายการ
- ในรายละเอียด assessment แยก Current profile และ Target profile ชัดเจน; บนมือถือเรียงเป็นแนวตั้ง
- Response panel เป็นส่วนต่อเนื่องของ assessment แต่มีพื้นผิวรองลงมา เพื่อไม่แย่งความสนใจจากข้อมูลหลัก
- CTA หลักมีหนึ่งจุดต่อบริบท; action รองใช้ปุ่ม outline และ action อันตรายใช้สีแดงแบบ semantic เท่านั้น

## Typography and spacing

- ใช้ font stack เดิมที่มีอยู่ในระบบเพื่อไม่เพิ่ม dependency และหลีกเลี่ยง layout shift
- เพิ่ม line-height ของข้อความอ่านและ textarea ให้สบายตา
- ใช้ type scale ที่แยก heading, label, body และ metadata ชัดเจน
- ใช้ spacing scale เดียวกันระหว่าง card, section และ form field
- ลด uppercase/letter spacing ให้เหลือเฉพาะ eyebrow หรือ section index ที่เป็น navigation cue

## Interaction and accessibility

- รักษา semantic button สำหรับ action และ label ที่เชื่อมกับ form control
- focus-visible ต้องเห็นชัดทั้งบนพื้นขาวและพื้น teal
- สถานะ success, error, warning ใช้สีตามความหมายและมีข้อความ ไม่ใช้สีอย่างเดียว
- รองรับ `prefers-reduced-motion`
- บน mobile ให้ grid ยุบเป็นหนึ่งคอลัมน์ ปุ่มเต็มความกว้างเมื่อจำเป็น และไม่ทำให้ข้อความถูกตัด

## Implementation constraints

- แก้ใน global tokens/styles และ component markup เท่าที่จำเป็น
- reuse component เดิม; ไม่เพิ่ม UI kit, icon set หรือ animation package
- คง API และ test contract เดิม
- ตรวจด้วย frontend test, TypeScript check และ `git diff --check`

## Acceptance criteria

1. หน้าหลักทั้งหมดใช้ white-first palette ตาม 60/30/10 และมีลำดับข้อมูลที่สม่ำเสมอ
2. หน้า assessment อ่านทีละรายการได้ชัดเจนทั้ง desktop และ mobile
3. Primary/secondary/danger actions แยกความสำคัญได้ทันที
4. ไม่มี horizontal overflow ที่ breakpoint หลัก และ focus state ยังมองเห็น
5. ชุดทดสอบเดิมและ TypeScript check ผ่านโดยไม่เพิ่ม dependency ใหม่
