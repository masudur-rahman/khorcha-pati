import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProfile, updateProfile } from '../api/endpoints'
import { User, Smartphone, Globe, Save, X, Edit3, Shield } from 'lucide-react'

export default function Settings() {
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

  return (
    <div className="space-y-8 pb-8">
      <header className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 tracking-tight">Settings</h1>
          <p className="text-gray-500 text-sm mt-1">Manage your account and preferences</p>
        </div>
        {!isEditing ? (
          <button
            onClick={() => setIsEditing(true)}
            className="flex items-center justify-center gap-2 bg-blue-600 text-white px-6 py-3 rounded-2xl text-sm font-bold hover:bg-blue-700 transition-all shadow-lg shadow-blue-100 group cursor-pointer"
          >
            <Edit3 size={18} />
            Edit Profile
          </button>
        ) : (
          <div className="flex gap-3">
            <button
              onClick={() => setIsEditing(false)}
              className="flex items-center justify-center gap-2 bg-white text-gray-500 px-6 py-3 rounded-2xl text-sm font-bold border border-gray-100 hover:bg-gray-50 transition-all shadow-sm"
            >
              <X size={18} />
              Cancel
            </button>
            <button
              onClick={handleSave}
              className="flex items-center justify-center gap-2 bg-emerald-600 text-white px-6 py-3 rounded-2xl text-sm font-bold hover:bg-emerald-700 transition-all shadow-lg shadow-emerald-100"
              disabled={update.isPending}
            >
              <Save size={18} />
              {update.isPending ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        )}
      </header>

      <div className="grid lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2 space-y-8">
            <section className="bg-white rounded-[2rem] shadow-sm border border-gray-100 overflow-hidden">
                <div className="p-8 border-b border-gray-50 bg-gray-50/30">
                    <h2 className="text-xs font-bold text-gray-400 uppercase tracking-[0.2em]">Personal Information</h2>
                </div>
                <div className="p-8 space-y-6">
                    <div className="grid sm:grid-cols-2 gap-8">
                        <InfoItem 
                            label="Telegram Username" 
                            value={`@${profile.username}`} 
                            icon={<User size={18} className="text-blue-500" />} 
                        />
                        <InfoItem 
                            label="Full Name" 
                            value={`${profile.firstName} ${profile.lastName || ''}`} 
                            icon={<Shield size={18} className="text-emerald-500" />} 
                        />
                        <InfoItem 
                            label="Telegram ID" 
                            value={profile.telegramId.toString()} 
                            isCode 
                        />
                    </div>
                </div>
            </section>

            <section className="bg-white rounded-[2rem] shadow-sm border border-gray-100 overflow-hidden">
                <div className="p-8 border-b border-gray-50 bg-gray-50/30">
                    <h2 className="text-xs font-bold text-gray-400 uppercase tracking-[0.2em]">Contact & Preferences</h2>
                </div>
                <div className="p-8 space-y-8">
                    <div className="space-y-2">
                        <label className="text-[10px] font-bold uppercase tracking-widest text-gray-400 ml-1">Mobile Number</label>
                        <div className="relative group cursor-pointer">
                            <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-blue-500 transition-colors">
                                <Smartphone size={18} />
                            </div>
                            {isEditing ? (
                                <input
                                    type="text"
                                    className="w-full bg-gray-50 border border-gray-100 rounded-2xl pl-12 pr-4 py-4 text-sm focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all outline-none font-bold"
                                    value={mobileNumber}
                                    onChange={e => setMobileNumber(e.target.value)}
                                    placeholder="e.g. +88017..."
                                />
                            ) : (
                                <div className="w-full bg-gray-50 border border-transparent rounded-2xl pl-12 pr-4 py-4 text-sm font-bold text-gray-900">
                                    {profile.mobileNumber || 'Not provided'}
                                </div>
                            )}
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label className="text-[10px] font-bold uppercase tracking-widest text-gray-400 ml-1">Timezone</label>
                        <div className="relative group cursor-pointer">
                            <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-blue-500 transition-colors">
                                <Globe size={18} />
                            </div>
                            {isEditing ? (
                                <select
                                    className="w-full bg-gray-50 border border-gray-100 rounded-2xl pl-12 pr-4 py-4 text-sm focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all outline-none font-bold appearance-none cursor-pointer"
                                    value={timezone}
                                    onChange={e => setTimezone(e.target.value)}
                                >
                                    <option value="UTC">Universal Time (UTC)</option>
                                    <option value="Asia/Dhaka">Asia/Dhaka (GMT+6)</option>
                                    <option value="America/New_York">New York (EST)</option>
                                    <option value="Europe/London">London (GMT)</option>
                                </select>
                            ) : (
                                <div className="w-full bg-gray-50 border border-transparent rounded-2xl pl-12 pr-4 py-4 text-sm font-bold text-gray-900">
                                    {timezone || 'UTC'}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </section>
        </div>

        <div className="space-y-6">
            <div className="bg-blue-600 rounded-[2rem] p-8 text-white shadow-xl shadow-blue-100 relative overflow-hidden">
                <div className="absolute top-0 right-0 -translate-y-1/4 translate-x-1/4 w-32 h-32 bg-white/10 rounded-full blur-2xl"></div>
                <h3 className="text-xl font-bold mb-2">Pro Tip</h3>
                <p className="text-blue-100 text-sm leading-relaxed font-medium">
                    You can also update your timezone directly via Telegram by sending your location to the bot.
                </p>
            </div>
            
            <div className="bg-slate-900 rounded-[2rem] p-8 text-white">
                <h3 className="text-lg font-bold mb-4">Account Security</h3>
                <p className="text-slate-400 text-xs leading-relaxed mb-6 font-medium">
                    Your account is linked to your Telegram profile. For maximum security, enable Two-Step Verification in your Telegram settings.
                </p>
                <div className="p-4 bg-white/5 rounded-2xl border border-white/10 flex items-center gap-3">
                    <div className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></div>
                    <span className="text-[10px] font-bold uppercase tracking-widest text-slate-300">Telegram Secured</span>
                </div>
            </div>
        </div>
      </div>
    </div>
  )
}

function InfoItem({ label, value, icon, isCode }: { label: string; value: string; icon?: React.ReactNode; isCode?: boolean }) {
    return (
        <div className="space-y-1.5">
            <p className="text-[10px] font-bold text-gray-400 uppercase tracking-widest ml-1">{label}</p>
            <div className="flex items-center gap-3 bg-gray-50 px-4 py-3 rounded-2xl border border-transparent">
                {icon}
                <span className={`text-sm font-bold text-gray-900 ${isCode ? 'font-mono text-xs' : ''}`}>{value}</span>
            </div>
        </div>
    )
}
