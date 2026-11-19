import test from 'node:test'
import assert from 'node:assert/strict'
import { normalizeRechargeSubmission } from './recharge.js'

const address = 'TJRabPrwbZy45sbavfcjinPJC18kjpRTv8'
const txHash = 'AB'.repeat(32)

test('申报保持原始金额精度并归一化哈希，预览直接截断两位', () => {
  const result = normalizeRechargeSubmission(address, ' 12.999999 ', ` ${txHash} `)
  assert.deepEqual(result, {
    payload: { address, amount: '12.999999', tx_hash: 'ab'.repeat(32) },
    creditedAmount: '12.99',
  })
})

test('拒绝低于最低金额、超过上限、科学计数法和超精度输入', () => {
  for (const amount of ['0', '0.999999', '-1', '1e3', 'NaN', '1.1234567', '100000000', '99999999.999999']) {
    assert.throws(() => normalizeRechargeSubmission(address, amount, txHash))
  }
  assert.equal(normalizeRechargeSubmission(address, '1.009999', txHash).creditedAmount, '1.00')
  assert.equal(normalizeRechargeSubmission(address, '99999999.99', txHash).creditedAmount, '99999999.99')
})

test('无地址或哈希格式错误时不能申报', () => {
  assert.throws(() => normalizeRechargeSubmission('', '1', txHash))
  for (const hash of ['', 'g'.repeat(64), 'a'.repeat(63), `0x${txHash}`]) {
    assert.throws(() => normalizeRechargeSubmission(address, '1', hash))
  }
})
