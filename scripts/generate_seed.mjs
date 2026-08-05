import fs from "node:fs/promises";
import { FileBlob, SpreadsheetFile } from "@oai/artifact-tool";

const input = await FileBlob.load("C:/Users/CHAWAKAN/Downloads/CSF 2.0 Organizational Profile Template Thai R1.xlsx");
const workbook = await SpreadsheetFile.importXlsx(input);
const rows = workbook.worksheets.getItem("Current and Target Profile").getRange("A2:B135").values;
const q = (value) => `'${String(value ?? "").replaceAll("'", "''")}'`;
const functions = new Map();
const categories = new Map();
const subcategories = [];
for (const [rawCode, rawDescription] of rows) {
  const code = String(rawCode ?? "").trim();
  const description = String(rawDescription ?? "").trim();
  if (/^[A-Z]{2}$/.test(code)) functions.set(code, description);
  else if (/^[A-Z]{2}\.[A-Z]{2}$/.test(code)) categories.set(code, { functionCode: code.slice(0, 2), description });
  else if (/^[A-Z]{2}\.[A-Z]{2}-\d{2}$/.test(code)) subcategories.push({ code, categoryCode: code.slice(0, 5), description });
}
const lines = ["BEGIN;", "TRUNCATE subcategories, categories, functions CASCADE;"];
for (const [code, description] of functions) lines.push(`INSERT INTO functions(code,name,description) VALUES (${q(code)},${q(code)},${q(description)});`);
for (const [code, category] of categories) lines.push(`INSERT INTO categories(function_id,code,name,description) SELECT id,${q(code)},${q(code)},${q(category.description)} FROM functions WHERE code=${q(category.functionCode)};`);
for (const row of subcategories) lines.push(`INSERT INTO subcategories(category_id,code,description) SELECT id,${q(row.code)},${q(row.description)} FROM categories WHERE code=${q(row.categoryCode)};`);
lines.push("COMMIT;", "");
await fs.writeFile("db/init/002_seed.sql", lines.join("\n"), "utf8");
console.log(JSON.stringify({ functions: functions.size, categories: categories.size, subcategories: subcategories.length }));
