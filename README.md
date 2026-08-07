# My Blog

基于 **Hexo + Butterfly 主题** 的个人博客，部署到 GitHub Pages。

## 本地预览

```bash
pnpm install
pnpm exec hexo server
```

然后打开 http://localhost:4000

## 写文章

```bash
pnpm exec hexo new "文章标题"
```

生成的文章在 `source/_posts/` 目录，用 Markdown 编写，完成后提交推送即可自动发布。

## 发布到 GitHub Pages

1. 在 GitHub 创建仓库 `<你的用户名>.github.io`（本项目对应 `realmorro369-arch.github.io`）
2. 把本目录推送上去：

```bash
git remote add origin git@github.com:realmorro369-arch/realmorro369-arch.github.io.git
git branch -M main
git push -u origin main
```

3. 到仓库 Settings → Pages，把 **Build and deployment 的 Source** 选为 **GitHub Actions**

之后每次 `git push` 都会自动构建并发布。

## 常用配置

- 站点信息（标题、作者、语言、URL）：`_config.yml`
- 主题外观（头像、封面、菜单、特效、统计）：`_config.butterfly.yml`
- 友链：`source/_data/link.yml`
- 封面/头像图片：`source/img/`
