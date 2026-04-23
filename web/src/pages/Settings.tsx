import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProfile, updateProfile } from '../api/endpoints'
import { useSearch } from '../context/SearchContext'

import TopBar from '../components/layout/TopBar'
import Card from '../components/ui/Card'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import Select from '../components/ui/Select'
import { ICONS } from '../components/ui/Icons'

export default function Settings() {
  const { searchTerm } = useSearch()
  const qc = useQueryClient()
  const { data: profile, isLoading } = useQuery({ queryKey: ['profile'], queryFn: getProfile })
  const update = useMutation({
    mutationFn: updateProfile,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['profile'] })
      setIsEditing(false)
    },
  })

  const [isEditing, setIsEditing] = useState(false)
  const [mobileNumber, setMobileNumber] = useState('')
  const [timezone, setTimezone] = useState('')

  useEffect(() => {
    if (profile) {
      setMobileNumber(profile.mobileNumber || '')
      setTimezone(profile.timezone || 'UTC')
    }
  }, [profile])

  if (isLoading) return <p className="text-gray-500">Loading...</p>
  if (!profile) return <p className="text-gray-400">Could not load profile</p>

  const handleSave = () => {
    update.mutate({ mobileNumber, timezone })
  }

  const matchesSearch = (text: string) => !searchTerm || text.toLowerCase().includes(searchTerm.toLowerCase())

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
      <TopBar title="Settings" subtitle="Manage your account and preferences" />

      <header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <h2 style={{ fontSize: 16, fontWeight: 700, color: 'var(--color-text-primary)' }}>Profile Details</h2>
        {!isEditing ? (
          <Button onClick={() => setIsEditing(true)} icon={ICONS.edit(18)}>Edit Profile</Button>
        ) : (
          <div style={{ display: 'flex', gap: 12 }}>
            <Button variant="secondary" onClick={() => setIsEditing(false)}>Cancel</Button>
            <Button 
                onClick={handleSave} 
                disabled={update.isPending}
            >
                {update.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>
        )}
      </header>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 24, alignItems: 'start' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
            {matchesSearch('personal information') && (
                <Card padding={0} style={{ overflow: 'hidden' }}>
                    <div style={{ padding: '20px 24px', borderBottom: '1px solid var(--color-border)', background: 'var(--color-bg)' }}>
                        <h3 style={{ fontSize: 11, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.1em' }}>Personal Information</h3>
                    </div>
                    <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 20 }}>
                        <InfoItem 
                            label="Telegram User" 
                            value={`@${profile.username}`} 
                            icon={ICONS.user(18)} 
                            visible={matchesSearch('Telegram User') || matchesSearch(`@${profile.username}`)}
                        />
                        <InfoItem 
                            label="Full Name" 
                            value={`${profile.firstName} ${profile.lastName || ''}`} 
                            icon={ICONS.shield(18)} 
                            visible={matchesSearch('Full Name') || matchesSearch(`${profile.firstName} ${profile.lastName || ''}`)}
                        />
                        <InfoItem 
                            label="Telegram ID" 
                            value={profile.telegramId.toString()} 
                            isCode 
                            visible={matchesSearch('Telegram ID') || matchesSearch(profile.telegramId.toString())}
                        />
                    </div>
                </Card>
            )}

            {matchesSearch('Contact & Preferences') && (
                <Card padding={0} style={{ overflow: 'hidden', border: isEditing ? '1px solid var(--color-primary)' : undefined }}>
                    <div style={{ padding: '20px 24px', borderBottom: '1px solid var(--color-border)', background: isEditing ? 'var(--color-primary-subtle)' : 'var(--color-bg)' }}>
                        <h3 style={{ fontSize: 11, fontWeight: 700, color: isEditing ? 'var(--color-primary)' : 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.1em' }}>
                            Contact & Preferences {isEditing && '— EDITING'}
                        </h3>
                    </div>
                    <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 24 }}>
                        {isEditing ? (
                            <>
                                <Input 
                                    label="Mobile Number" 
                                    value={mobileNumber} 
                                    onChange={e => setMobileNumber(e.target.value)} 
                                    placeholder="+88017..." 
                                />
                                <Select 
                                    label="Timezone" 
                                    value={timezone} 
                                    onChange={e => setTimezone(e.target.value)} 
                                    options={[
                                        { value: 'UTC', label: 'Universal Time (UTC)' },
                                        { value: 'Asia/Dhaka', label: 'Asia/Dhaka (GMT+6)' },
                                        { value: 'America/New_York', label: 'New York (EST)' },
                                        { value: 'Europe/London', label: 'London (GMT)' },
                                    ]} 
                                />
                            </>
                        ) : (
                            <>
                                <InfoItem 
                                    label="Mobile Number" 
                                    value={profile.mobileNumber || 'Not provided'} 
                                    icon={ICONS.creditCard(18)} 
                                    visible={matchesSearch('Mobile Number') || matchesSearch(profile.mobileNumber || 'Not provided')}
                                />
                                <InfoItem 
                                    label="Timezone" 
                                    value={timezone || 'UTC'} 
                                    icon={ICONS.budget(18)} 
                                    visible={matchesSearch('Timezone') || matchesSearch(timezone || 'UTC')}
                                />
                            </>
                        )}
                    </div>
                </Card>
            )}
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
            {matchesSearch('Pro Tip') && (
                <div style={{
                    background: `linear-gradient(135deg, var(--color-primary) 0%, oklch(0.55 0.14 165) 100%)`,
                    borderRadius: 24, padding: 32, color: 'white', position: 'relative', overflow: 'hidden',
                }}>
                    <div style={{ position: 'absolute', top: -20, right: -20, width: 120, height: 120, borderRadius: '50%', background: 'rgba(255,255,255,0.1)' }} />
                    <h3 style={{ fontSize: 18, fontWeight: 700, marginBottom: 8 }}>Pro Tip</h3>
                    <p style={{ fontSize: 14, opacity: 0.9, lineHeight: 1.6 }}>
                        You can also update your timezone directly via Telegram by sending your location to the bot.
                    </p>
                </div>
            )}
            
            {matchesSearch('Account Security') && (
                <div style={{ background: 'var(--color-text-primary)', borderRadius: 24, padding: 32, color: 'white' }}>
                    <h3 style={{ fontSize: 16, fontWeight: 700, marginBottom: 12 }}>Account Security</h3>
                    <p style={{ fontSize: 13, opacity: 0.7, lineHeight: 1.6, marginBottom: 24 }}>
                        Your account is linked to your Telegram profile. For maximum security, enable Two-Step Verification in your Telegram settings.
                    </p>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10, background: 'rgba(255,255,255,0.05)', padding: '12px 16px', borderRadius: 12, border: '1px solid rgba(255,255,255,0.1)' }}>
                        <div style={{ width: 8, height: 8, background: 'var(--color-success)', borderRadius: '50%' }} />
                        <span style={{ fontSize: 10, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em' }}>Telegram Secured</span>
                    </div>
                </div>
            )}
        </div>
      </div>
    </div>
  )
}

function InfoItem({ label, value, icon, isCode, visible = true }: { label: string; value: string; icon?: React.ReactNode; isCode?: boolean, visible?: boolean }) {
    if (!visible) return null
    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <p style={{ fontSize: 10, fontWeight: 700, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.08em', marginLeft: 4 }}>{label}</p>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, background: 'var(--color-bg)', padding: '12px 16px', borderRadius: 12, border: '1px solid var(--color-border)' }}>
                {icon && <span style={{ color: 'var(--color-primary)', display: 'flex' }}>{icon}</span>}
                <span style={{ fontSize: 14, fontWeight: 700, color: 'var(--color-text-primary)', fontFamily: isCode ? 'monospace' : 'inherit' }}>{value}</span>
            </div>
        </div>
    )
}
