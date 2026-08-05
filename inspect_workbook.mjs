import { FileBlob, SpreadsheetFile } from "@oai/artifact-tool";

const inputPath = "C:/Users/CHAWAKAN/Downloads/CSF 2.0 Organizational Profile Template Thai R1.xlsx";
const input = await FileBlob.load(inputPath);
const workbook = await SpreadsheetFile.importXlsx(input);

const summary = await workbook.inspect({
  kind: "workbook,sheet,table",
  maxChars: 12000,
  tableMaxRows: 8,
  tableMaxCols: 18,
  tableMaxCellChars: 120,
});
console.log("SUMMARY");
console.log(summary.ndjson);

const sheets = await workbook.inspect({ kind: "sheet", include: "id,name", maxChars: 5000 });
console.log("SHEETS");
console.log(sheets.ndjson);

const profile = workbook.worksheets.getItem("Current and Target Profile");
const rows = profile.getRange("A2:P135").values;
const idRows = rows.map((row) => String(row[0] ?? ""));
const functionRows = idRows.filter((id) => /^[A-Z]{2}$/.test(id));
const categoryRows = idRows.filter((id) => /^[A-Z]{2}\.[A-Z]{2}$/.test(id));
const subcategoryRows = idRows.filter((id) => /^[A-Z]{2}\.[A-Z]{2}-\d{2}$/.test(id));
const included = rows.map((row) => row[2]).filter((value) => value !== null && value !== "");
console.log("PROFILE_STATS");
console.log(JSON.stringify({
  dataRows: rows.length,
  functions: functionRows.length,
  categories: categoryRows.length,
  subcategories: subcategoryRows.length,
  includedValues: included.reduce((acc, value) => { const key = String(value); acc[key] = (acc[key] ?? 0) + 1; return acc; }, {}),
  columns: profile.getRange("A1:P1").values[0],
}, null, 2));

for (const item of workbook.worksheets.items) {
  const used = item.getUsedRange();
  if (!used) continue;
  console.log(`USED ${item.name}`);
  console.log(JSON.stringify({ values: used.values, formulas: used.formulas }, null, 2).slice(0, 30000));
}
