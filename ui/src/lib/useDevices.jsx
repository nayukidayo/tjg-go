import { useEffect, useState } from 'react'
import dayjs from 'dayjs'

export default function useDevices() {
  const [devices, setDevices] = useState(items)

  useEffect(() => {
    const es = new EventSource('/api/data/live')
    es.onmessage = ({ data }) => {
      try {
        const obj = JSON.parse(data)
        setDevices(prev => merge(prev, obj))
      } catch (err) {
        console.log(err)
      }
    }
    return () => {
      es.close()
    }
  }, [])

  return devices
}

const items = [
  { id: '10', name: '01-340-01', code: '01', gps: {}, rfid: {} },
  { id: '0A', name: '01-340-02', code: '02', gps: {}, rfid: {} },
  { id: '12', name: '01-340-03', code: '03', gps: {}, rfid: {} },
  { id: '0D', name: '01-340-04', code: '04', gps: {}, rfid: {} },
  { id: '0E', name: '01-340-08', code: '08', gps: {}, rfid: {} },
  { id: '17', name: '01-340-09', code: '09', gps: {}, rfid: {} },
  { id: '0F', name: '02-340-01', code: '51', gps: {}, rfid: {} },
  { id: '05', name: '02-340-02', code: '52', gps: {}, rfid: {} },
  { id: '15', name: '02-340-03', code: '53', gps: {}, rfid: {} },
  { id: '02', name: '02-340-04', code: '54', gps: {}, rfid: {} },
  { id: '01', name: '02-340-05', code: '55', gps: {}, rfid: {} },
  { id: '08', name: '02-340-06', code: '56', gps: {}, rfid: {} },
  { id: '1A', name: '02-340-07', code: '57', gps: {}, rfid: {} },
  { id: '04', name: '02-340-80', code: 'T01', gps: {}, rfid: {} },
  { id: '0C', name: '02-340-81', code: 'T02', gps: {}, rfid: {} },
  { id: '07', name: '02-340-82', code: 'T05', gps: {}, rfid: {} },
  { id: '12', name: '02-340-83', code: 'T06', gps: {}, rfid: {} },
  { id: '09', name: '02-340-84', code: 'T07', gps: {}, rfid: {} },
  { id: '11', name: '02-340-85', code: 'T08', gps: {}, rfid: {} },
  { id: '0B', name: '02-340-86', code: 'T09', gps: {}, rfid: {} },
  { id: '14', name: '02-340-87', code: 'T10', gps: {}, rfid: {} },
  { id: '13', name: '02-340-88', code: 'T11', gps: {}, rfid: {} },
  { id: '19', name: '02-340-89', code: 'T12', gps: {}, rfid: {} },
]

function merge(prev, obj) {
  try {
    const arr = []
    for (let i = 0; i < prev.length; i++) {
      if (prev[i].id === obj.device?.toUpperCase()) {
        const ts = dayjs(obj.ts).format('YYYY-MM-DD HH:mm:ss')
        if (obj.type === 'GPS') {
          prev[i].gps = obj.data
          prev[i].gps.ts = ts
        } else {
          prev[i].rfid.tag = obj.data[0]?.tag
          prev[i].rfid.rssi = obj.data[0]?.rssi
          prev[i].rfid.ts = ts
        }
      }
      arr.push(prev[i])
    }
    return arr
  } catch (_) {
    return prev
  }
}
