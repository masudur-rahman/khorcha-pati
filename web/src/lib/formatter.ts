export const fmt = (n: number, minDec = 2) => {
  const formatted = (n || 0).toLocaleString(undefined, {
    minimumFractionDigits: minDec,
    maximumFractionDigits: minDec
  })
  return '৳' + formatted
}
