// HTTP 环境的 WebView 可能没有 Clipboard API，回退到选区复制。
export async function copyText(text) {
  if (!text) throw new Error('没有可复制的内容')
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
      return
    }
  } catch { /* 尝试兼容方式 */ }
  const previous = document.activeElement
  const field = document.createElement('textarea')
  field.value = text
  field.readOnly = true
  field.style.cssText = 'position:fixed;top:0;left:-9999px;opacity:0'
  document.body.appendChild(field)
  try {
    field.select()
    field.setSelectionRange(0, text.length)
    if (!document.execCommand('copy')) throw new Error('复制失败，请长按手动复制')
  } finally {
    field.remove()
    previous?.focus?.({ preventScroll: true })
  }
}
