---
title: 用 Hexo + Butterfly 搭建一个免费博客
date: 2026-08-07 12:30:00
tags:
  - 教程
  - Hexo
cover: https://tc.alcy.cc/tc/20260429/0bb4d26f7dc33cfbd48fe7b74b42e21b.webp
---

这篇文章是一份搭建记录，照着做你也能拥有一个同款博客。

## 用到的技术

- **Hexo**：基于 Node.js 的静态博客生成器
- **Butterfly**：Hexo 生态中最流行的主题之一
- **GitHub Pages**：免费的静态网站托管

## 初始化项目

```bash
pnpm add hexo-cli -g
hexo init blog
cd blog
pnpm install
```

## 安装主题

```bash
git clone https://github.com/jerryc127/hexo-theme-butterfly.git themes/butterfly
pnpm add hexo-renderer-pug hexo-wordcount
```

然后把根目录 `_config.yml` 里的 `theme` 改成 `butterfly`。

## 写文章与发布

```bash
hexo new "文章标题"
hexo clean && hexo generate
```

配置好 GitHub Actions 后，每次 `git push` 都会自动构建并发布，非常省心。

> 提示：`hexo new` 生成的文章放在 `source/_posts/` 目录，用 Markdown 写就行。
