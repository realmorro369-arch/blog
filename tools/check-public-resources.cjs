const fs = require("node:fs");
const path = require("node:path");
const root = path.resolve(process.argv[2] || "public");
const missing = new Set();
function walk(dir) { for (const item of fs.readdirSync(dir, { withFileTypes: true })) { const full = path.join(dir, item.name); if (item.isDirectory()) walk(full); else if (item.name.endsWith(".html")) { const text = fs.readFileSync(full, "utf8"); for (const match of text.matchAll(/(?:src|href)=["'](\/(?!\/)[^"'#?]+)(?:\?[^"']*)?["']/g)) { const rawPath = match[1]; let target = path.resolve(root, `.${decodeURIComponent(rawPath)}`); if (fs.existsSync(target) && fs.statSync(target).isDirectory()) target = path.join(target, "index.html"); if (!fs.existsSync(target)) missing.add(rawPath); } } } }
walk(root);
if (missing.size) { console.error(`生成站点存在失效站内资源：\n${[...missing].join("\n")}`); process.exit(1); }
console.log("生成站点资源检查通过");
