const { app, BrowserWindow, dialog, protocol, shell } = require("electron");
const fs = require("node:fs/promises");
const path = require("node:path");

const rendererDirectory = path.join(__dirname, "..", "dist", "public");

protocol.registerSchemesAsPrivileged([
  {
    scheme: "morroblog",
    privileges: {
      standard: true,
      secure: true,
      supportFetchAPI: true,
      corsEnabled: true,
    },
  },
]);

const contentTypes = new Map([
  [".css", "text/css; charset=utf-8"],
  [".html", "text/html; charset=utf-8"],
  [".js", "text/javascript; charset=utf-8"],
  [".json", "application/json; charset=utf-8"],
  [".png", "image/png"],
  [".svg", "image/svg+xml"],
]);

function contentTypeFor(filePath) {
  return contentTypes.get(path.extname(filePath).toLowerCase()) || "application/octet-stream";
}

function registerRendererProtocol() {
  protocol.handle("morroblog", async (request) => {
    const requestUrl = new URL(request.url);
    const relativePath = requestUrl.pathname === "/" ? "index.html" : decodeURIComponent(requestUrl.pathname).replace(/^\/+/, "");
    if (relativePath === "__manus__/logs") {
      return new Response(null, { status: 204 });
    }
    const filePath = path.resolve(rendererDirectory, relativePath);

    if (filePath !== rendererDirectory && !filePath.startsWith(`${rendererDirectory}${path.sep}`)) {
      return new Response("Forbidden", { status: 403 });
    }

    try {
      const body = await fs.readFile(filePath);
      return new Response(body, {
        headers: { "content-type": contentTypeFor(filePath) },
      });
    } catch {
      return new Response("Not found", { status: 404 });
    }
  });
}

const windowOptions = {
  width: 1440,
  height: 920,
  minWidth: 900,
  minHeight: 620,
  backgroundColor: "#e9e5dc",
  autoHideMenuBar: true,
  title: "正文校样室",
  webPreferences: {
    contextIsolation: true,
    nodeIntegration: false,
    sandbox: true,
  },
};

function reportLoadFailure(errorCode, errorDescription, validatedURL) {
  const details = [
    `加载地址：${validatedURL || "本地页面"}`,
    `错误代码：${errorCode}`,
    `错误说明：${errorDescription || "未知错误"}`,
  ].join("\n");
  dialog.showErrorBox("正文校样室未能加载", `${details}\n\n请将此信息反馈给开发者。`);
}

function createPreviewWindow(url) {
  const preview = new BrowserWindow({
    ...windowOptions,
    minWidth: 960,
    minHeight: 640,
    backgroundColor: "#eef2f5",
    title: "本地预览 · MorroBlog",
  });
  preview.loadURL(url).catch((error) => dialog.showErrorBox("预览窗口未能加载", error.message));
}

function createWindow() {
  const window = new BrowserWindow(windowOptions);
  window.webContents.on("did-fail-load", (_event, errorCode, errorDescription, validatedURL, isMainFrame) => {
    if (isMainFrame) reportLoadFailure(errorCode, errorDescription, validatedURL);
  });
  window.webContents.setWindowOpenHandler(({ url }) => {
    if (url.includes("preview=1")) {
      createPreviewWindow(url);
      return { action: "deny" };
    }
    if (/^https?:\/\//i.test(url)) shell.openExternal(url);
    return { action: "deny" };
  });
  window.loadURL("morroblog://app/").catch((error) => dialog.showErrorBox("正文校样室未能加载", error.message));
}

app.whenReady().then(() => {
  registerRendererProtocol();
  createWindow();
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") app.quit();
});
