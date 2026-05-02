import { useEffect, useMemo, useState } from 'react'
import { apiUserTransactionDetailUrl } from '../../../api/endpoints'
import {
  userTransactionDetailQuery,
  type DailySummaryPoint,
  type UserTransaction,
  type UserTransactionDateRange,
} from '../userDetailTransactions'

export function useUserTransactionDetail(userId: string, dateRange: UserTransactionDateRange) {
  const [transactions, setTransactions] = useState<UserTransaction[]>([])
  const [dailySummary, setDailySummary] = useState<DailySummaryPoint[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const kind = dateRange.kind
  const fromDay = dateRange.kind === 'between' ? dateRange.fromDay : ''
  const toDay = dateRange.kind === 'between' ? dateRange.toDay : ''

  const rangeForFetch = useMemo<UserTransactionDateRange>(() => {
    if (kind === 'all') return { kind: 'all' }
    return { kind: 'between', fromDay, toDay }
  }, [kind, fromDay, toDay])

  useEffect(() => {
    if (!userId) {
      setTransactions([])
      setDailySummary([])
      setIsLoading(false)
      setError(null)
      return
    }

    let cancelled = false
    const controller = new AbortController()

    const loadTransactions = async () => {
      setIsLoading(true)
      setError(null)
      try {
        const detailQuery = userTransactionDetailQuery(rangeForFetch)
        const detailResponse = await fetch(apiUserTransactionDetailUrl(userId, detailQuery), {
          signal: controller.signal,
        })

        if (!detailResponse.ok) {
          throw new Error(`transactions detail request failed with ${detailResponse.status}`)
        }

        const body = (await detailResponse.json()) as {
          transactions: Array<{
            id: string
            amount: number
            location: string
            fraud_status: boolean
            created_at?: string
          }>
          summary: Array<{
            bucket_start: string
            request_count: number
            anomaly_count: number
            amount_sum: number
            amount_avg: number
          }>
        }
        const payload = body.transactions
        const summaryPayload = body.summary
        if (cancelled) return
        const mapped = payload.map((tx) => ({
          id: tx.id,
          amount: tx.amount,
          location: tx.location,
          fraudStatus: tx.fraud_status,
          createdAt: tx.created_at ?? '',
        }))

        const mappedSummary = summaryPayload
          .map((item) => ({
            bucketStart: item.bucket_start,
            requestCount: item.request_count,
            anomalyCount: item.anomaly_count,
            amountSum: item.amount_sum,
            amountAvg: item.amount_avg,
          }))
          .sort((a, b) => new Date(a.bucketStart).getTime() - new Date(b.bucketStart).getTime())

        setTransactions(mapped)
        setDailySummary(mappedSummary)
      } catch (err) {
        if (cancelled) return
        if (err instanceof DOMException && err.name === 'AbortError') return
        setError('Could not load user transactions')
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }

    void loadTransactions()
    return () => {
      cancelled = true
      controller.abort()
    }
  }, [userId, rangeForFetch])

  return { transactions, dailySummary, isLoading, error }
}
