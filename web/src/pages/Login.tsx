import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { requestOTP, verifyOTP, initQR, pollQR } from '../api/endpoints'
import { QRCodeSVG } from 'qrcode.react'

export default function Login() {
  const [tab, setTab] = useState<'otp' | 'qr'>('otp')
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="bg-white rounded-lg shadow-md p-8 w-full max-w-md">
        <h1 className="text-2xl font-bold text-center mb-6">Expense Tracker</h1>
        <div className="flex mb-6 border-b">
          <button
            className={`flex-1 pb-2 text-sm font-medium ${tab === 'otp' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-500'}`}
            onClick={() => setTab('otp')}
          >Enter Code</button>
          <button
            className={`flex-1 pb-2 text-sm font-medium ${tab === 'qr' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-500'}`}
            onClick={() => setTab('qr')}
          >Scan QR</button>
        </div>
        {tab === 'otp' ? <OTPLogin /> : <QRLogin />}
      </div>
    </div>
  )
}

function OTPLogin() {
  const [identifier, setIdentifier] = useState('')
  const [code, setCode] = useState('')
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()

  const handleSend = async () => {
    setError('')
    setLoading(true)
    try {
      await requestOTP(identifier)
      setSent(true)
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  const handleVerify = async () => {
    setError('')
    setLoading(true)
    try {
      const data = await verifyOTP(identifier, code)
      login(data.accessToken, (data as any).refreshToken)
      navigate('/')
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      <input
        className="w-full border rounded px-3 py-2 text-sm"
        placeholder="Username or phone number"
        value={identifier}
        onChange={e => setIdentifier(e.target.value)}
        disabled={sent}
      />
      {sent && (
        <input
          className="w-full border rounded px-3 py-2 text-sm"
          placeholder="6-digit code"
          value={code}
          onChange={e => setCode(e.target.value)}
          maxLength={6}
        />
      )}
      {error && <p className="text-red-600 text-sm">{error}</p>}
      <button
        className="w-full bg-blue-600 text-white rounded py-2 text-sm font-medium disabled:opacity-50"
        onClick={sent ? handleVerify : handleSend}
        disabled={loading || (!sent && !identifier) || (sent && code.length < 6)}
      >
        {loading ? 'Please wait...' : sent ? 'Verify Code' : 'Send Code'}
      </button>
      {sent && (
        <button className="w-full text-sm text-gray-500" onClick={() => { setSent(false); setCode('') }}>
          Resend code
        </button>
      )}
    </div>
  )
}

function QRLogin() {
  const [deepLink, setDeepLink] = useState('')
  const [status, setStatus] = useState<'idle' | 'pending' | 'approved' | 'error'>('idle')
  const [error, setError] = useState('')
  const { login } = useAuth()
  const navigate = useNavigate()

  const startQR = async () => {
    setError('')
    try {
      const data = await initQR()
      setDeepLink(data.deepLink)
      setStatus('pending')
      pollSession(data.sessionID)
    } catch (e: any) {
      setError(e.message)
    }
  }

  const pollSession = async (sessionID: string) => {
    const interval = setInterval(async () => {
      try {
        const data = await pollQR(sessionID)
        if (data.status === 'approved' && data.accessToken) {
          clearInterval(interval)
          login(data.accessToken, (data as any).refreshToken)
          setStatus('approved')
          navigate('/')
        } else if (data.status === 'expired') {
          clearInterval(interval)
          setStatus('error')
          setError('QR session expired. Try again.')
        }
      } catch {
        clearInterval(interval)
        setStatus('error')
        setError('Polling failed')
      }
    }, 3000)
  }

  if (status === 'idle' || status === 'error') {
    return (
      <div className="text-center space-y-4">
        <p className="text-sm text-gray-600">
          Generate a QR code and scan it with your Telegram app to log in.
        </p>
        {error && <p className="text-red-600 text-sm">{error}</p>}
        <button
          className="bg-blue-600 text-white rounded px-4 py-2 text-sm font-medium"
          onClick={startQR}
        >Generate QR Code</button>
      </div>
    )
  }

  return (
    <div className="text-center space-y-4">
      <QRCodeSVG value={deepLink} size={200} className="mx-auto" />
      <p className="text-sm text-gray-600">
        Scan this QR with your phone camera, or <a href={deepLink} className="text-blue-600 underline" target="_blank" rel="noopener noreferrer">click here</a> on mobile.
      </p>
      <p className="text-xs text-gray-400 animate-pulse">Waiting for confirmation...</p>
    </div>
  )
}
