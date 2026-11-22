const reportedFrontendErrors = new WeakSet();

function logError(scope, error, context = {}) {
  if (error !== null && typeof error === 'object') {
    if (reportedFrontendErrors.has(error)) return;
    reportedFrontendErrors.add(error);
  }
  const message = error instanceof Error ? error.message : String(error);
  const details = { scope, page: window.location.pathname, message: message.slice(0, 2000), ...context };
  console.error('[frontend]', details);
  sendFrontendLog(JSON.stringify(details), 'error');
}

function sendFrontendLog(msg, level = 'info') {
  // 日誌傳送失敗只輸出至 Console，避免遞迴呼叫日誌 API。
  fetch('/api/log', {
    method: 'POST',
    keepalive: true,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ msg, level }),
  }).then(async res => {
    if (!res.ok) {
      console.error('[frontend] 錯誤日誌傳送失敗', { status: res.status });
      return;
    }
    const data = await res.json();
    if (data === null || data.ok !== true) {
      console.error('[frontend] 日誌 API 回應格式不正確，請確認連到正確的記帳服務');
    }
  }).catch(() => {
    console.error('[frontend] 日誌傳送或回應解析失敗');
  });
}

async function requestJSON(url, options = {}, validate = () => true) {
  const method = options.method || 'GET';
  let status = null;
  try {
    const res = await fetch(url, options);
    status = res.status;
    if (!res.ok) {
      throw new Error(`API 請求失敗，HTTP ${status}`);
    }
    if (method === 'DELETE') return null;

    let data;
    try {
      data = await res.json();
    } catch {
      // 不將可能包含資料內容的 JSON 解析錯誤原文寫入日誌。
      throw new Error('API 回應不是有效 JSON，請確認連到正確的記帳服務');
    }
    if (!validate(data)) {
      throw new Error('API 回應格式不正確，請確認網址是否連到正確的記帳服務');
    }
    return data;
  } catch (error) {
    logError('API 請求失敗', error, { method, path: url.split('?')[0], status });
    throw error;
  }
}

function isCategory(data) {
  return data !== null && typeof data === 'object'
    && typeof data.id === 'string' && typeof data.name === 'string'
    && typeof data.icon === 'string';
}

function isRecord(data) {
  return data !== null && typeof data === 'object'
    && Number.isInteger(data.id) && typeof data.categoryId === 'string'
    && typeof data.date === 'string' && Number.isFinite(data.amount);
}

function isValidationResult(data) {
  return data !== null && typeof data === 'object' && typeof data.valid === 'boolean';
}

window.addEventListener('error', event => {
  if (event instanceof ErrorEvent) {
    logError('未處理的 JavaScript 錯誤', event.error || event.message, { line: event.lineno });
  } else {
    logError('頁面資源載入失敗', '無法載入頁面所需資源', { tag: event.target.tagName });
  }
}, true);

window.addEventListener('unhandledrejection', event => {
  logError('未處理的非同步錯誤', event.reason);
});
