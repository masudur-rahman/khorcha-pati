import { useState, useEffect, useRef } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { requestOTP, verifyOTP, initQR, pollQR, verifyMagicLink } from '../api/endpoints'
import { QRCodeSVG } from 'qrcode.react'

export default function Login() {
  const [tab, setTab] = useState<'otp' | 'qr'>('otp')
  const [magicLoading, setMagicLoading] = useState(false)
  const [magicError, setMagicError] = useState('')
  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  // QR Session State lifted to parent
  const [qrDeepLink, setQrDeepLink] = useState('')
  const [, setQrSessionID] = useState('')
  const [qrStatus, setQrStatus] = useState<'idle' | 'pending' | 'approved' | 'error'>('idle')
  const [qrError, setQrError] = useState('')
  const pollingRef = useRef<any>(null)

  useEffect(() => {
    const params = new URLSearchParams(location.search)
    const token = params.get('token')
    if (token) {
      handleMagicLogin(token)
    }
  }, [location])

  // Cleanup polling on unmount
  useEffect(() => {
    return () => {
      if (pollingRef.current) clearInterval(pollingRef.current)
    }
  }, [])

  const handleMagicLogin = async (token: string) => {
    setMagicLoading(true)
    setMagicError('')
    try {
      const data = await verifyMagicLink(token)
      login(data.accessToken, (data as any).refreshToken)
      navigate('/')
    } catch (e: any) {
      setMagicError(e.message || 'Magic link login failed')
    } finally {
      setMagicLoading(false)
    }
  }

  const startQR = async () => {
    if (qrStatus === 'pending') return // Don't restart if already pending
    setQrError('')
    setQrStatus('idle')
    try {
      const data = await initQR()
      setQrDeepLink(data.deepLink)
      setQrSessionID(data.sessionID)
      setQrStatus('pending')
      startPolling(data.sessionID)
    } catch (e: any) {
      setQrError(e.message)
      setQrStatus('error')
    }
  }

  const startPolling = (sessionID: string) => {
    if (pollingRef.current) clearInterval(pollingRef.current)
    pollingRef.current = setInterval(async () => {
      try {
        const data = await pollQR(sessionID)
        if (data.status === 'approved' && data.accessToken) {
          if (pollingRef.current) clearInterval(pollingRef.current)
          login(data.accessToken, (data as any).refreshToken)
          setQrStatus('approved')
          navigate('/')
        } else if (data.status === 'denied') {
          if (pollingRef.current) clearInterval(pollingRef.current)
          setQrStatus('error')
          setQrError('Access Denied. You cancelled the login from Telegram.')
        } else if (data.status === 'expired') {
          if (pollingRef.current) clearInterval(pollingRef.current)
          setQrStatus('error')
          setQrError('QR session expired. Please generate a new one.')
        }
      } catch {
        if (pollingRef.current) clearInterval(pollingRef.current)
        setQrStatus('error')
        setQrError('Connection lost. Please try again.')
      }
    }, 3000)
  }

  if (magicLoading) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
        <p className="text-gray-600 font-bold">Authenticating...</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md border border-gray-100">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2 tracking-tight">Welcome Back</h1>
          <p className="text-gray-500 text-sm font-medium">Sign in to your expense dashboard</p>
        </div>

        {magicError && (
          <div className="mb-6 p-3 bg-red-50 border border-red-200 text-red-600 text-sm rounded-xl text-center font-bold">
            {magicError}
          </div>
        )}

        <div className="flex mb-8 bg-gray-100 p-1.5 rounded-xl">
          <button
            className={`flex-1 py-2.5 text-sm font-bold rounded-lg transition-all cursor-pointer ${tab === 'otp' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}`}
            onClick={() => setTab('otp')}
          >OTP Code</button>
          <button
            className={`flex-1 py-2.5 text-sm font-bold rounded-lg transition-all cursor-pointer ${tab === 'qr' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}`}
            onClick={() => {
                setTab('qr');
                if (qrStatus === 'idle') startQR();
            }}
          >QR Scan</button>
        </div>

        {tab === 'otp' ? (
          <OTPLogin />
        ) : (
          <QRLogin 
            status={qrStatus} 
            deepLink={qrDeepLink} 
            error={qrError} 
            onRetry={startQR} 
          />
        )}

        <div className="mt-8 pt-6 border-t border-gray-50 text-center">
          <p className="text-xs text-gray-400 font-medium">
            Need help? Contact <a href="https://t.me/expense_tracker_bot" className="text-blue-600 hover:underline font-bold">Support Bot</a>
          </p>
        </div>
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
      <div>
        <label className="block text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1.5 ml-1">Identity</label>
        <input
          className="w-full bg-gray-50 border border-gray-100 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all outline-none font-medium"
          placeholder="Username or phone"
          value={identifier}
          onChange={e => setIdentifier(e.target.value)}
          disabled={sent}
        />
      </div>
      {sent && (
        <div>
          <label className="block text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1.5 ml-1">Verification Code</label>
          <input
            className="w-full bg-gray-50 border border-gray-100 rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all outline-none tracking-widest text-center text-lg font-bold"
            placeholder="000000"
            value={code}
            onChange={e => setCode(e.target.value)}
            maxLength={6}
          />
        </div>
      )}
      {error && (
        <p className="text-red-500 text-xs font-bold px-1 mt-1">{error}</p>
      )}
      <button
        className="w-full bg-blue-600 hover:bg-blue-700 text-white rounded-xl py-3.5 text-sm font-bold shadow-lg shadow-blue-100 transition-all disabled:opacity-50 disabled:shadow-none mt-2 cursor-pointer"
        onClick={sent ? handleVerify : handleSend}
        disabled={loading || (!sent && !identifier) || (sent && code.length < 6)}
      >
        {loading ? 'Please wait...' : sent ? 'Verify Code' : 'Send Code'}
      </button>
      {sent && (
        <button 
          className="w-full text-xs text-gray-400 font-bold hover:text-blue-600 transition-colors py-2 cursor-pointer" 
          onClick={() => { setSent(false); setCode('') }}
        >
          Resend code
        </button>
      )}
    </div>
  )
}

function QRLogin({ status, deepLink, error, onRetry }: { 
    status: string; 
    deepLink: string; 
    error: string; 
    onRetry: () => void 
}) {
  if (status === 'idle') {
    return (
      <div className="text-center py-10">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p className="text-gray-500 text-sm font-medium">Generating secure QR code...</p>
      </div>
    )
  }

  if (status === 'error') {
    const isDenied = error.includes('Denied')
    return (
      <div className="text-center space-y-6 py-4">
        <div className={`w-16 h-16 ${isDenied ? 'bg-orange-50 text-orange-600' : 'bg-red-50 text-red-600'} rounded-full flex items-center justify-center mx-auto text-2xl`}>
          {isDenied ? '🚫' : '⚠️'}
        </div>
        <div className="space-y-1">
          <p className="text-sm font-bold text-gray-900">{isDenied ? 'Login Denied' : 'QR Generation Failed'}</p>
          <p className="text-xs text-gray-500 font-medium px-4">{error}</p>
        </div>
        <button
          className="w-full bg-gray-900 hover:bg-black text-white rounded-xl py-3.5 text-sm font-bold shadow-lg transition-all cursor-pointer"
          onClick={onRetry}
        >Generate New QR</button>
      </div>
    )
  }

  return (
    <div className="text-center space-y-6 py-2">
      <div className="bg-white p-4 rounded-2xl border-2 border-dashed border-blue-100 inline-block mx-auto shadow-sm">
        <QRCodeSVG value={deepLink} size={180} />
      </div>
      <div className="space-y-2">
        <p className="text-sm text-gray-600 font-bold">
          Scan with your Telegram app
        </p>
        <p className="text-xs text-gray-400 font-medium">
          Or <a href={deepLink} className="text-blue-600 hover:underline font-bold" target="_blank" rel="noopener noreferrer">click here</a> to open in Telegram
        </p>
      </div>
      <div className="flex items-center justify-center space-x-3 text-blue-600 bg-blue-50 py-2 rounded-xl">
        <div className="animate-pulse flex space-x-1">
          <div className="h-1.5 w-1.5 bg-blue-600 rounded-full"></div>
          <div className="h-1.5 w-1.5 bg-blue-600 rounded-full animation-delay-200"></div>
          <div className="h-1.5 w-1.5 bg-blue-600 rounded-full animation-delay-400"></div>
        </div>
        <span className="text-[10px] font-bold uppercase tracking-[0.2em]">Waiting for scan</span>
      </div>
    </div>
  )
}
