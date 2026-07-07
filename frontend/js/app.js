// FileVersion 前端逻辑 —— 通过 Wails 绑定调用 Go 后端。
// 绑定路径：window.go.main.App.<Method>

const App = window.go.main.App;

const state = {
  tab: "overall",
  files: [],
};

// ---------- 工具 ----------
function $(sel) { return document.querySelector(sel); }
function $all(sel) { return Array.from(document.querySelectorAll(sel)); }

function call(method, ...args) {
  return App[method](...args);
}

// ---------- Tab 切换 ----------
$all(".tab").forEach(btn => {
  btn.addEventListener("click", () => {
    state.tab = btn.dataset.tab;
    $all(".tab").forEach(b => b.classList.toggle("active", b === btn));
    $all(".tabpage").forEach(p => p.classList.toggle("hidden", p.dataset.page !== state.tab));
    refreshPreview();
  });
});

// ---------- 文件列表 ----------
async function loadFiles() {
  state.files = await call("View");
  renderTable();
}

async function addFiles(paths) {
  state.files = await call("AddFiles", paths);
  renderTable();
  refreshPreview();
}

async function removeFile(path) {
  state.files = await call("RemoveFile", path);
  renderTable();
  refreshPreview();
}

async function clearFiles() {
  state.files = await call("ClearFiles");
  renderTable();
  refreshPreview();
}

// 把路径数组转为 Go 需要的格式
function renderTable() {
  const body = $("#fileBody");
  body.innerHTML = "";
  const has = state.files.length > 0;
  $("#dropzone").classList.toggle("hidden", has);
  $("#fileCount").textContent = state.files.length + " 个文件";
  for (const f of state.files) {
    const tr = document.createElement("tr");
    const resClass = f.status === "conflict" ? "res-skip" : (f.preview && f.preview !== f.oldName ? "res-ok" : "");
    tr.innerHTML =
      `<td class="old" title="${f.path}">${escapeHtml(f.oldName)}</td>` +
      `<td class="prev">${escapeHtml(f.preview || "")}</td>` +
      `<td class="${resClass}">${f.status === "conflict" ? "⚠ " + escapeHtml(f.err || "冲突") : (f.preview !== f.oldName ? "→ 将改名" : "不变")}</td>`;
    tr.addEventListener("dblclick", () => removeFile(f.path));
    body.appendChild(tr);
  }
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

// ---------- 读取表单 → PlanOptions ----------
function readOptions() {
  const tpl = {
    Pattern: $("#tplPattern").value,
    Start: parseInt($("#tplStart").value) || 1,
    Increment: parseInt($("#tplInc").value) || 1,
    Digits: parseInt($("#tplDigits").value) || 1,
    PadChar: $("#tplPad").value === " " ? " " : "0",
    Random: $("#tplRandom").checked,
    RandomLower: $("#tplRandomLower").checked,
    ExtOverride: $("#tplExt").value.trim(),
    AutoConflict: $("#conflict").value === "auto",
  };
  const rep = {
    Find: $("#repFind").value,
    Replace: $("#repReplace").value,
    UseRegex: $("#repRegex").checked,
    IgnoreCase: $("#repIgnore").checked,
  };
  const ar = {
    Prefix: $("#arPrefix").value,
    Suffix: $("#arSuffix").value,
    InsertAt: parseInt($("#arInsertAt").value) || 0,
    InsertStr: $("#arInsertStr").value,
    RemoveStr: $("#arRemoveStr").value,
    RemoveFrom: parseInt($("#arRemoveFrom").value) || 0,
    RemoveCount: parseInt($("#arRemoveCount").value) || 0,
  };
  // clean / version 选项仅在“整体”Tab 可见，其余 Tab 不生效
  const onOverall = state.tab === "overall";
  return {
    tab: state.tab,
    template: tpl,
    replace: rep,
    addRemove: ar,
    clean: onOverall && $("#optClean").checked,
    version: onOverall && $("#optVersion").checked,
    versionMove: onOverall && $("#optVersionMove").checked,
    conflict: $("#conflict").value,
  };
}

// ---------- 预览（dry-run） ----------
let previewTimer = null;
async function refreshPreview() {
  if (state.files.length === 0) return;
  const opt = readOptions();
  // 只有在对应 Tab 有配置时才启用该操作
  if (state.tab === "overall") { opt.Replace = null; opt.AddRemove = null; }
  if (state.tab === "replace") { opt.Template = null; opt.AddRemove = null; }
  if (state.tab === "addremove") { opt.Template = null; opt.Replace = null; }
  state.files = await call("ComputePreview", opt);
  renderTable();
}

// 任意输入变化即刷新预览（防抖）
$all(".panel input, .panel select").forEach(el => {
  el.addEventListener("input", () => {
    clearTimeout(previewTimer);
    previewTimer = setTimeout(refreshPreview, 250);
  });
  el.addEventListener("change", () => {
    clearTimeout(previewTimer);
    previewTimer = setTimeout(refreshPreview, 250);
  });
});

// ---------- 执行 ----------
async function doRename() {
  if (state.files.length === 0) { setStatus("请先添加文件"); return; }
  const opt = readOptions();
  if (state.tab === "overall") { opt.Replace = null; opt.AddRemove = null; }
  if (state.tab === "replace") { opt.Template = null; opt.Replace = opt.Replace; opt.AddRemove = null; }
  if (state.tab === "addremove") { opt.Template = null; opt.Replace = null; }
  setStatus("执行中…");
  const res = await call("ExecuteRename", opt);
  if (res.ok) {
    setStatus(`完成：处理 ${res.count} 个文件` + (res.revert ? "（可撤销）" : ""));
    state.files = res.files || [];
    renderTable();
  } else {
    setStatus("失败：" + (res.error || "未知错误"));
  }
}

function setStatus(t) { $("#statusText").textContent = t; }

// ---------- 按钮 ----------
$("#btnAdd").addEventListener("click", () => $("#fileInput").click());
$("#fileInput").addEventListener("change", e => {
  const paths = Array.from(e.target.files).map(f => f.path || f.name);
  if (paths.length) addFiles(paths);
  e.target.value = "";
});
$("#btnRemove").addEventListener("click", () => {
  const sel = document.querySelector("#fileBody tr:hover");
  if (sel) removeFile(sel.querySelector(".old").title);
});
$("#btnClear").addEventListener("click", clearFiles);
$("#btnRename").addEventListener("click", doRename);
$("#btnRevert").addEventListener("click", async () => {
  const r = await call("Revert");
  if (r.ok) setStatus(`已撤销 ${r.count} 个文件`);
  else setStatus("撤销失败：" + (r.error || ""));
  await loadFiles();
});
$("#btnInstall").addEventListener("click", async () => {
  const r = await call("Install");
  setStatus(r ? "安装失败：" + r : "已安装右键菜单");
});
$("#btnUninstall").addEventListener("click", async () => {
  const r = await call("Uninstall");
  setStatus(r ? "卸载失败：" + r : "已卸载");
});
$("#btnHelp").addEventListener("click", () => {
  alert(
    "FileVersion 批量改名\n\n" +
    "整体：用模板 A_# / *V.# 批量命名，支持编号、随机、改扩展名。\n" +
    "替换：查找/替换文件名内容（支持正则、忽略大小写）。\n" +
    "添加/删除：前后缀、指定位置插入、按子串或位置删除字符。\n" +
    "clean：去 (N) 计数并把日期规整为 V.YYYY_MM_DD。\n" +
    "版本号：追加/更新 V.时间戳（copy=复制，move=原地改名）。\n\n" +
    "预览列实时显示改名结果；冲突可配置跳过/覆盖/自动编号。\n" +
    "“撤销上次”可还原上一批操作。"
  );
});

// ---------- 拖拽 ----------
const dz = $("#dropzone");
const tw = $(".tablewrap");
tw.addEventListener("dragover", e => { e.preventDefault(); tw.style.outline = "2px dashed #2f6db0"; });
tw.addEventListener("dragleave", () => { tw.style.outline = "none"; });
tw.addEventListener("drop", async e => {
  e.preventDefault();
  tw.style.outline = "none";
  const paths = [];
  if (e.dataTransfer.files) {
    for (const f of e.dataTransfer.files) paths.push(f.path);
  }
  if (paths.length) addFiles(paths);
});

// ---------- 启动 ----------
(async function init() {
  await loadFiles();
  setStatus("就绪");
})();
