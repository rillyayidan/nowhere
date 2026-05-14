import { useCallback, useState } from 'react'

const fallbackLocation = {
  latitude: -6.2,
  longitude: 106.816666,
  accuracy: 0,
  label: 'Jakarta demo location',
}

export function useLocation() {
  const [location, setLocation] = useState(fallbackLocation)
  const [status, setStatus] = useState('demo')
  const [error, setError] = useState('')

  const requestLocation = useCallback(() => {
    if (!navigator.geolocation) {
      setStatus('demo')
      setError('GPS is unavailable, using demo location.')
      setLocation(fallbackLocation)
      return
    }

    setStatus('loading')
    setError('')
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setLocation({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
          accuracy: position.coords.accuracy,
          label: 'Live GPS',
        })
        setStatus('live')
      },
      () => {
        setStatus('demo')
        setError('GPS permission was denied, using demo location.')
        setLocation(fallbackLocation)
      },
      { enableHighAccuracy: true, timeout: 6000, maximumAge: 60000 },
    )
  }, [])

  return { location, status, error, requestLocation }
}
