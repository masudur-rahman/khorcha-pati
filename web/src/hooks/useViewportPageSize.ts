import { RefObject, useLayoutEffect, useState } from 'react'

interface Options {
  desktopReserve?: number
  mobileReserve?: number
  fallbackRow?: number
  fallbackCard?: number
  min?: number
  max?: number
}

// Display-oriented page size for tables embedded in a scrolling page (e.g. the
// Admin panels). Same principle as the standalone transactions table — scale the
// row count to the display height using a measured item height, computed once and
// kept stable — but without framing the body to the viewport, since these panels
// live on a page that scrolls through several sections.
export function useViewportPageSize(
  rowRef: RefObject<HTMLElement | null>,
  cardRef: RefObject<HTMLElement | null>,
  opts: Options = {},
): number {
  const {
    desktopReserve = 360, mobileReserve = 340,
    fallbackRow = 56, fallbackCard = 110, min = 5, max = 20,
  } = opts
  const [size, setSize] = useState(10)

  useLayoutEffect(() => {
    const clamp = (n: number) => Math.max(min, Math.min(max, n))
    const isMobile = window.innerWidth < 768
    const itemH = isMobile
      ? (cardRef.current?.offsetHeight || fallbackCard)
      : (rowRef.current?.offsetHeight || fallbackRow)
    const reserve = isMobile ? mobileReserve : desktopReserve
    setSize(clamp(Math.floor((window.innerHeight - reserve) / itemH)))
    // Intentionally run once — the count must stay stable for the session.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return size
}
