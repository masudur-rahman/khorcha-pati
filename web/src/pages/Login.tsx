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

  // QR Session State
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
    if (qrStatus === 'pending') return
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

  const styles = {
    wrapper: {
      minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center',
      background: 'var(--color-bg)',
      padding: 24, position: 'relative' as const,
    },
    card: {
      width: '100%', maxWidth: 420, background: 'var(--color-surface)', borderRadius: 24,
      padding: '48px 40px', position: 'relative' as const, zIndex: 1,
      boxShadow: '0 25px 60px rgba(0,0,0,0.06), 0 0 0 1px var(--color-border)',
    },
    tabBar: {
      display: 'flex', background: 'var(--color-bg)', borderRadius: 12, padding: 4, marginBottom: 28, gap: 4,
    },
    tab: (active: boolean) => ({
      flex: 1, padding: '10px 0', textAlign: 'center' as const, borderRadius: 10, fontSize: 13, fontWeight: 600,
      cursor: 'pointer', border: 'none', transition: 'all 0.2s',
      background: active ? 'var(--color-surface)' : 'transparent',
      color: active ? 'var(--color-primary)' : 'var(--color-text-tertiary)',
      boxShadow: active ? '0 2px 8px rgba(0,0,0,0.05)' : 'none',
    }),
    input: {
      width: '100%', padding: '14px 16px', borderRadius: 12, fontSize: 14,
      border: '1px solid var(--color-border)', background: 'var(--color-bg)',
      color: 'var(--color-text-primary)', outline: 'none', transition: 'border 0.2s',
      boxSizing: 'border-box' as const,
    },
    label: {
      fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)',
      textTransform: 'uppercase' as const, letterSpacing: '0.06em', marginBottom: 6, display: 'block',
    },
    btn: {
      width: '100%', padding: '14px 0', borderRadius: 12, fontSize: 14, fontWeight: 700,
      background: 'var(--color-primary)', color: 'white', border: 'none', cursor: 'pointer',
      boxShadow: '0 4px 16px var(--color-primary-shadow)', transition: 'all 0.15s', marginTop: 8,
    },
  }

  if (magicLoading) {
    return (
      <div style={styles.wrapper}>
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 mb-4" style={{ borderColor: 'var(--color-primary)' }}></div>
        <p style={{ color: 'var(--color-text-secondary)', fontWeight: 700 }}>Authenticating...</p>
      </div>
    )
  }

  return (
    <div style={styles.wrapper}>
      <div style={styles.card}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <h1 style={{ fontSize: 28, fontWeight: 700, color: 'var(--color-text-primary)', margin: '0 0 6px', letterSpacing: '-0.02em' }}>Welcome Back</h1>
          <p style={{ fontSize: 13, color: 'var(--color-text-tertiary)', margin: 0, fontWeight: 500 }}>Sign in to your expense dashboard</p>
        </div>

        {magicError && (
          <div style={{ marginBottom: 24, padding: 12, background: 'var(--color-danger-subtle)', border: '1px solid var(--color-danger)', color: 'var(--color-danger)', fontSize: 13, borderRadius: 12, textAlign: 'center', fontWeight: 700 }}>
            {magicError}
          </div>
        )}

        <div style={styles.tabBar}>
          <button style={styles.tab(tab === 'otp')} onClick={() => setTab('otp')}>OTP Code</button>
          <button style={styles.tab(tab === 'qr')} onClick={() => { setTab('qr'); if (qrStatus === 'idle') startQR(); }}>QR Scan</button>
        </div>

        <div style={{ minHeight: 220 }}>
          {tab === 'otp' ? (
            <OTPLogin styles={styles} />
          ) : (
            <QRLogin 
              status={qrStatus} 
              deepLink={qrDeepLink} 
              error={qrError} 
              onRetry={startQR} 
              styles={styles}
            />
          )}
        </div>

        <div style={{ marginTop: 28, paddingTop: 20, borderTop: '1px solid var(--color-border)', textAlign: 'center' }}>
          <p style={{ fontSize: 11, color: 'var(--color-text-tertiary)', fontWeight: 500 }}>
            Need help? Contact <a href="https://t.me/expense_tracker_bot" target="_blank" rel="noreferrer" style={{ color: 'var(--color-primary)', fontWeight: 700, textDecoration: 'none' }}>Support Bot</a>
          </p>
        </div>
      </div>
    </div>
  )
}

function OTPLogin({ styles }: { styles: any }) {
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

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      if (!sent && identifier && !loading) {
        handleSend()
      } else if (sent && code.length === 6 && !loading) {
        handleVerify()
      }
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div>
        <label style={styles.label}>Identity</label>
        <input style={styles.input} placeholder="Username or phone"
          value={identifier} onChange={e => setIdentifier(e.target.value)}
          disabled={sent}
          onKeyDown={handleKeyDown}
        />
      </div>
      {sent && (
        <div className="animate-in fade-in slide-in-from-top-1 duration-300">
          <label style={styles.label}>Verification Code</label>
          <input style={{...styles.input, textAlign: 'center', letterSpacing: '0.3em', fontSize: 20, fontWeight: 700}}
            placeholder="000000" value={code} onChange={e => setCode(e.target.value)} maxLength={6}
            onKeyDown={handleKeyDown}
            autoFocus
          />
        </div>
      )}
      {error && (
        <p style={{ color: 'var(--color-danger)', fontSize: 11, fontWeight: 700, textAlign: 'center', margin: 0 }}>{error}</p>
      )}
      <button style={styles.btn}
        onClick={sent ? handleVerify : handleSend}
        disabled={loading || (!sent && !identifier) || (sent && code.length < 6)}
      >
        {loading ? 'Please wait...' : sent ? 'Verify Code' : 'Send Code'}
      </button>
      {sent && (
        <button style={{ background: 'none', border: 'none', fontSize: 12, color: 'var(--color-text-tertiary)', cursor: 'pointer', padding: 8, fontWeight: 600 }}
          onClick={() => { setSent(false); setCode('') }}>
          Resend code
        </button>
      )}
    </div>
  )
}

function QRLogin({ status, deepLink, error, onRetry, styles }: { 
    status: string; 
    deepLink: string; 
    error: string; 
    onRetry: () => void;
    styles: any;
}) {
  if (status === 'idle') {
    return (
      <div style={{ textAlign: 'center', padding: '32px 0' }}>
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 mx-auto mb-4" style={{ borderColor: 'var(--color-primary)' }}></div>
        <p style={{ color: 'var(--color-text-tertiary)', fontSize: 13, fontWeight: 500 }}>Generating secure QR code...</p>
      </div>
    )
  }

  if (status === 'error') {
    const isDenied = error.includes('Denied')
    return (
      <div style={{ textAlign: 'center', padding: '16px 0', display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ width: 64, height: 64, borderRadius: '50%', background: 'var(--color-danger-subtle)', color: 'var(--color-danger)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto', fontSize: 24, border: '1px solid var(--color-danger)' }}>
          {isDenied ? '🚫' : '⚠️'}
        </div>
        <div style={{ margin: '8px 0' }}>
          <p style={{ fontSize: 14, fontWeight: 700, color: 'var(--color-text-primary)', margin: '0 0 4px' }}>{isDenied ? 'Login Denied' : 'QR Generation Failed'}</p>
          <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', margin: 0, padding: '0 16px' }}>{error}</p>
        </div>
        <button
          style={{ ...styles.btn, background: 'var(--color-text-primary)', color: 'var(--color-surface)', boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
          onClick={onRetry}
        >Generate New QR</button>
      </div>
    )
  }

  return (
    <div style={{ textAlign: 'center', padding: '4px 0' }}>
      <div style={{
        width: 180, height: 180, margin: '0 auto 20px', borderRadius: 16,
        border: '2px dashed var(--color-primary-subtle)', display: 'flex', alignItems: 'center', justifyContent: 'center',
        background: 'var(--color-surface)',
      }}>
        <QRCodeSVG value={deepLink} size={150} />
      </div>
      <div style={{ marginBottom: 16 }}>
        <p style={{ fontSize: 14, fontWeight: 700, color: 'var(--color-text-secondary)', margin: '0 0 8px' }}>Scan with your Telegram app</p>
        <p style={{ fontSize: 12, color: 'var(--color-text-tertiary)', margin: 0 }}>
          Or <a href={deepLink} target="_blank" rel="noreferrer" style={{ color: 'var(--color-primary)', fontWeight: 700, textDecoration: 'none' }}>click here</a> to open in Telegram
        </p>
      </div>
      <div style={{
        display: 'inline-flex', alignItems: 'center', gap: 8, marginTop: 8,
        background: 'var(--color-primary-subtle)', padding: '8px 24px', borderRadius: 12,
      }}>
        <div style={{ display: 'flex', gap: 4 }}>
          <div style={{ height: 6, width: 6, background: 'var(--color-primary)', borderRadius: '50%', animation: 'pulse 1.5s infinite' }}></div>
        </div>
        <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--color-primary)', textTransform: 'uppercase', letterSpacing: '0.12em' }}>Waiting for scan</span>
      </div>
      <style>{`
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.4; } }
      `}</style>
    </div>
  )
}