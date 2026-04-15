import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getProfile, updateProfile } from '../api/endpoints'

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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Settings</h1>
        {!isEditing ? (
          <button
            onClick={() => setIsEditing(true)}
            className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
          >
            Edit Profile
          </button>
        ) : (
          <div className="flex gap-2">
            <button
              onClick={() => setIsEditing(false)}
              className="bg-gray-100 px-4 py-2 rounded text-sm hover:bg-gray-200"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700"
            >
              Save Changes
            </button>
          </div>
        )}
      </div>

      <div className="bg-white rounded-lg shadow p-6 max-w-md">
        <h2 className="text-sm font-semibold mb-4 text-gray-500 uppercase tracking-wider">User Information</h2>
        <div className="space-y-4">
          <div className="grid grid-cols-3 gap-2 py-2 border-b border-gray-50">
            <span className="text-gray-500 text-sm">Username</span>
            <span className="col-span-2 font-medium">@{profile.username}</span>
          </div>
          <div className="grid grid-cols-3 gap-2 py-2 border-b border-gray-50">
            <span className="text-gray-500 text-sm">Full Name</span>
            <span className="col-span-2 font-medium">{profile.firstName} {profile.lastName}</span>
          </div>
          <div className="grid grid-cols-3 gap-2 py-2 border-b border-gray-50">
            <span className="text-gray-500 text-sm">Telegram ID</span>
            <span className="col-span-2 font-mono text-xs">{profile.telegramId}</span>
          </div>

          <div className="grid grid-cols-3 gap-2 py-2 items-center">
            <span className="text-gray-500 text-sm">Mobile</span>
            <div className="col-span-2">
              {isEditing ? (
                <input
                  type="text"
                  className="w-full border rounded px-2 py-1 text-sm"
                  value={mobileNumber}
                  onChange={e => setMobileNumber(e.target.value)}
                />
              ) : (
                <span className="font-medium">{profile.mobileNumber || 'Not set'}</span>
              )}
            </div>
          </div>

          <div className="grid grid-cols-3 gap-2 py-2 items-center">
            <span className="text-gray-500 text-sm">Timezone</span>
            <div className="col-span-2">
              {isEditing ? (
                <select
                  className="w-full border rounded px-2 py-1 text-sm"
                  value={timezone}
                  onChange={e => setTimezone(e.target.value)}
                >
                  <option value="UTC">UTC</option>
                  <option value="Asia/Dhaka">Asia/Dhaka (GMT+6)</option>
                  <option value="America/New_York">America/New_York (EST)</option>
                  <option value="Europe/London">Europe/London (GMT)</option>
                </select>
              ) : (
                <span className="font-medium">{profile.timezone || 'UTC'}</span>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
