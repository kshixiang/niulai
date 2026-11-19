// 保留原始金额供后台核实；预览使用字符串截断，避免浮点进位。
export function normalizeRechargeSubmission(address, amount, txHash) {
  const raw = String(amount).trim()
  const hash = String(txHash).trim().toLowerCase()
  if (!address) throw new Error('充值地址不可用，请刷新后重试')
  if (!/^\d{1,8}(\.\d{1,6})?$/.test(raw)) throw new Error('请输入正确金额，最多6位小数')
  const [whole, fraction = ''] = raw.split('.')
  const micros = BigInt(whole) * 1000000n + BigInt(fraction.padEnd(6, '0'))
  if (micros < 1000000n || micros > 99999999990000n) throw new Error('充值金额须在1至99999999.99 USDT之间')
  if (!/^[a-f0-9]{64}$/.test(hash)) throw new Error('请输入64位交易哈希（TxID）')
  return {
    payload: { address, amount: raw, tx_hash: hash },
    creditedAmount: `${BigInt(whole)}.${fraction.padEnd(2, '0').slice(0, 2)}`,
  }
}
