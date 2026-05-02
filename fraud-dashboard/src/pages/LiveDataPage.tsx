import { useOutletContext } from 'react-router-dom'
import LiveDataView from '../features/live/LiveDataView'
import type { LiveTransactionsOutletContext } from '../features/live/useLiveTransactions'

export default function LiveDataPage() {
  const ctx = useOutletContext<LiveTransactionsOutletContext>()
  return <LiveDataView {...ctx} />
}
