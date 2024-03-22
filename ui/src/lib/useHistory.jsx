import { useEffect, useState } from 'react'
import dayjs from 'dayjs'

export default function useDevices(device) {
  const [type, setType] = useState('GPS')
  const [start, setStart] = useState(dayjs().add(-5, 'm').valueOf() * 1000)
  const [end, setEnd] = useState(Date.now() * 1000)

  const [total, setTotal] = useState(0)
  const [from, setFrom] = useState(0)
  const [size, setSize] = useState(20)

  const [data, setData] = useState([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    fetchHistory({ device: device.id, type, start, end, from, size })
      .then(hd => {
        if (type === 'GPS') {
          setData(hd.hits)
        } else {
          setData(
            hd.hits.map(v => {
              const obj = { ts: v.ts }
              JSON.parse(v.data).forEach((x, i) => {
                obj[`tag${i + 1}`] = x.tag
                obj[`rssi${i + 1}`] = x.rssi
              })
              return obj
            })
          )
        }
        setTotal(hd.total)
        setFrom(hd.from)
        setSize(hd.size)
      })
      .catch(err => {
        console.log(err)
      })
      .finally(() => {
        setLoading(false)
      })
  }, [type, start, end, total, from, size])

  return { data, loading, type, setType, setStart, setEnd, from, setFrom, total, size }
}

async function fetchHistory(query) {
  const res = await fetch('/api/data/history', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(query),
  })
  if (!res.ok) return new Error(res.statusText)
  return res.json()
}
