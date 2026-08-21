const fs = require("node:fs");
const path = require("node:path");

const postsDir = path.resolve(process.argv[2] || "source/_posts");
const configPath = path.resolve(process.argv[3] || "_config.butterfly.yml");
const failures = [];

for (const name of fs.readdirSync(postsDir).filter((item) => item.endsWith(".md"))) {
  const file = path.join(postsDir, name);
  const text = fs.readFileSync(file, "utf8");
  const frontMatter = text.match(/^---\s*\n([\s\S]*?)\n---/);
  if (!frontMatter) { failures.push(`${name}: 缺少 Front Matter`); continue; }
  for (const field of ["title", "date"]) if (!new RegExp(`^${field}:\\s*.+`, "m").test(frontMatter[1])) failures.push(`${name}: 缺少 ${field}`);
  for (const field of ["cover", "top_img"]) {
    const raw = frontMatter[1].match(new RegExp(`^${field}:\\s*["']?([^"'\\n]+)`, "m"))?.[1]?.trim();
    if (raw?.startsWith("/") && !fs.existsSync(path.resolve("source", `.${raw}`))) failures.push(`${name}: ${field} 指向不存在的本地资源 ${raw}`);
  }
}

const themeConfig = fs.readFileSync(configPath, "utf8");
if (/^\s*-\s*\/img\/covers\//m.test(themeConfig) || /default_top_img:\s*\/img\/cover\.png/.test(themeConfig)) failures.push("主题默认封面仍引用已删除的 covers 或 cover.png");
if (failures.length) { console.error(failures.join("\n")); process.exit(1); }
console.log("Front Matter 与主题默认资源检查通过");
