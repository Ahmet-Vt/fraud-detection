import { useEffect, useMemo, useRef, useState } from 'react'

export type LiveTransaction = {
  id: string
  userid: string
  userName: string
  userSurname: string
  amount: number
  location: string
  fraudstatus: boolean
  createdat: string
}

type ConnectionMode = 'idle' | 'sse' | 'disconnected'

export type FraudRatePoint = {
  minuteLabel: string
  anomalyRate: number | null
  requestCount: number
  fraudCount: number
}

const chartBucketMs = 15000
const chartWindowPoints = 240

const sseReconnectBaseMs = 1000
const sseReconnectMaxMs = 60_000

function formatChartBucketTimeLabel(date: Date): string {
  return date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

function createInitialFraudRateSeries(now = Date.now()): FraudRatePoint[] {
  const start = now
  return Array.from({ length: chartWindowPoints }, (_, index) => ({
    minuteLabel: formatChartBucketTimeLabel(new Date(start + index * chartBucketMs)),
    anomalyRate: null,
    requestCount: 0,
    fraudCount: 0,
  }))
}

type UseLiveTransactionsOptions = {
  sseUrl?: string
  maxItems?: number
  onFraudTransaction?: (tx: LiveTransaction) => void
}

type SseEvaluatedTransactionPayload = {
  transaction_id: string
  user_id: string
  user_name: string
  user_surname: string
  amount: number
  location: string
  fraud_status: boolean
  created_at: string
}

function mapSseEvaluatedTransaction(tx: SseEvaluatedTransactionPayload): LiveTransaction {
  return {
    id: tx.transaction_id,
    userid: tx.user_id,
    userName: tx.user_name,
    userSurname: tx.user_surname,
    amount: tx.amount,
    location: tx.location,
    fraudstatus: tx.fraud_status,
    createdat: tx.created_at,
  }
}

function isSseEvaluatedTransactionPayload(value: unknown): value is SseEvaluatedTransactionPayload {
  if (!value || typeof value !== 'object') return false
  const v = value as Record<string, unknown>
  return (
    typeof v.transaction_id === 'string' &&
    typeof v.user_id === 'string' &&
    typeof v.user_name === 'string' &&
    typeof v.user_surname === 'string' &&
    typeof v.amount === 'number' &&
    typeof v.location === 'string' &&
    typeof v.fraud_status === 'boolean' &&
    typeof v.created_at === 'string'
  )
}

function isLiveTransaction(value: unknown): value is LiveTransaction {
  if (!value || typeof value !== 'object') return false
  const v = value as Record<string, unknown>
  return (
    typeof v.id === 'string' &&
    typeof v.userid === 'string' &&
    typeof v.userName === 'string' &&
    typeof v.userSurname === 'string' &&
    typeof v.amount === 'number' &&
    typeof v.location === 'string' &&
    typeof v.fraudstatus === 'boolean' &&
    typeof v.createdat === 'string'
  )
}

function normalizePayload(input: unknown): LiveTransaction[] {
  const list = Array.isArray(input) ? input : [input]

  const mapped: LiveTransaction[] = []
  for (const item of list) {
    if (isLiveTransaction(item)) {
      mapped.push(item)
      continue
    }
    if (isSseEvaluatedTransactionPayload(item)) {
      mapped.push(mapSseEvaluatedTransaction(item))
    }
  }

  return mapped
}

function prependUnique(
  incoming: LiveTransaction[],
  previous: LiveTransaction[],
  maxItems: number,
): LiveTransaction[] {
  const next = [...incoming, ...previous]
  const seen = new Set<string>()
  const deduped: LiveTransaction[] = []
  for (const tx of next) {
    if (seen.has(tx.id)) continue
    seen.add(tx.id)
    deduped.push(tx)
    if (deduped.length >= maxItems) break
  }
  return deduped
}

export function useLiveTransactions({
  sseUrl,
  maxItems = 100,
  onFraudTransaction,
}: UseLiveTransactionsOptions) {
  const onFraudTransactionRef = useRef(onFraudTransaction)
  onFraudTransactionRef.current = onFraudTransaction
  const seenFraudNotifyIdsRef = useRef(new Set<string>())

  const [transactions, setTransactions] = useState<LiveTransaction[]>([])
  const [fraudRateSeries, setFraudRateSeries] = useState<FraudRatePoint[]>(() =>
    createInitialFraudRateSeries(),
  )
  const [connectionMode, setConnectionMode] = useState<ConnectionMode>('idle')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const minuteBufferRef = useRef<LiveTransaction[]>([])
  const nextChartWriteIndexRef = useRef(0)

  useEffect(() => {
    const flushMinuteBuffer = () => {
      const batch = minuteBufferRef.current
      minuteBufferRef.current = []

      const total = batch.length
      const suspiciousCount = batch.filter((tx) => tx.fraudstatus).length
      const anomalyRate = total === 0 ? 0 : (suspiciousCount / total) * 100
      const point: FraudRatePoint = {
        minuteLabel: formatChartBucketTimeLabel(new Date()),
        anomalyRate: Number(anomalyRate.toFixed(2)),
        requestCount: total,
        fraudCount: suspiciousCount,
      }

      setFraudRateSeries((prev) => {
        if (nextChartWriteIndexRef.current < chartWindowPoints) {
          const next = [...prev]
          next[nextChartWriteIndexRef.current] = {
            ...next[nextChartWriteIndexRef.current],
            ...point,
          }
          nextChartWriteIndexRef.current += 1
          return next
        }

        const shifted = prev.slice(1)
        shifted.push(point)
        return shifted
      })
    }

    const intervalId = window.setInterval(flushMinuteBuffer, chartBucketMs)
    return () => window.clearInterval(intervalId)
  }, [])

  useEffect(() => {
    if (!sseUrl) {
      setConnectionMode('disconnected')
      setError('sse url is not configured')
      setIsLoading(false)
      return
    }

    let unmounted = false
    let source: EventSource | null = null
    let reconnectTimer: ReturnType<typeof window.setTimeout> | null = null
    let reconnectAttempt = 0

    const resetReconnectBackoff = () => {
      reconnectAttempt = 0
    }

    const nextReconnectDelayMs = () => {
      const exp = Math.min(sseReconnectMaxMs, sseReconnectBaseMs * 2 ** reconnectAttempt)
      reconnectAttempt += 1
      const jitterRange = Math.max(1, Math.floor(exp * 0.25))
      return exp + Math.floor(Math.random() * jitterRange)
    }

    const clearReconnectTimer = () => {
      if (reconnectTimer != null) {
        window.clearTimeout(reconnectTimer)
        reconnectTimer = null
      }
    }

    const markConnected = () => {
      if (unmounted) return
      resetReconnectBackoff()
      setConnectionMode('sse')
      setIsLoading(false)
      setError(null)
    }

    const handlePayload = (raw: string) => {
      const parsed = JSON.parse(raw) as unknown
      const payload =
        parsed &&
        typeof parsed === 'object' &&
        'data' in (parsed as Record<string, unknown>)
          ? (parsed as Record<string, unknown>).data
          : parsed
      const normalized = normalizePayload(payload)
      if (normalized.length > 0) {
        for (const tx of normalized) {
          if (tx.fraudstatus && !seenFraudNotifyIdsRef.current.has(tx.id)) {
            seenFraudNotifyIdsRef.current.add(tx.id)
            onFraudTransactionRef.current?.(tx)
          }
        }
        minuteBufferRef.current.push(...normalized)
        setTransactions((prev) => prependUnique(normalized, prev, maxItems))
      }
    }

    const scheduleReconnect = () => {
      if (unmounted || reconnectTimer != null) return
      const wait = nextReconnectDelayMs()
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null
        connect()
      }, wait)
    }

    const connect = () => {
      if (unmounted) return
      source?.close()
      source = null

      try {
        source = new EventSource(sseUrl)

        source.onopen = () => {
          markConnected()
        }
        source.onmessage = (event) => {
          if (unmounted) return
          try {
            handlePayload(event.data)
          } catch {
            setError('sse message parse failed')
          }
        }
        source.addEventListener('transaction_evaluated', (event) => {
          if (unmounted) return
          const message = event as MessageEvent
          try {
            handlePayload(message.data)
          } catch {
            setError('sse message parse failed')
          }
        })
        source.addEventListener('ready', () => {
          markConnected()
        })
        source.onerror = () => {
          source?.close()
          source = null
          if (unmounted) return
          setConnectionMode('disconnected')
          setError('sse connection lost')
          setIsLoading(false)
          scheduleReconnect()
        }
      } catch {
        if (unmounted) return
        setConnectionMode('disconnected')
        setError('sse initialization failed')
        setIsLoading(false)
        scheduleReconnect()
      }
    }

    connect()

    return () => {
      unmounted = true
      clearReconnectTimer()
      source?.close()
    }
  }, [sseUrl, maxItems])

  const sortedTransactions = useMemo(
    () =>
      [...transactions].sort((a, b) => {
        const at = new Date(a.createdat).getTime()
        const bt = new Date(b.createdat).getTime()
        return bt - at
      }),
    [transactions],
  )

  const chartTotals = useMemo(() => {
    let requestCount = 0
    let fraudCount = 0
    for (const point of fraudRateSeries) {
      requestCount += point.requestCount
      fraudCount += point.fraudCount
    }
    return { requestCount, fraudCount }
  }, [fraudRateSeries])

  return {
    transactions: sortedTransactions,
    fraudRateSeries,
    chartTotals,
    connectionMode,
    isLoading,
    error,
  }
}

export type LiveTransactionsOutletContext = ReturnType<typeof useLiveTransactions>
