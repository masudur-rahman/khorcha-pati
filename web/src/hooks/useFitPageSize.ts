import { RefObject, useLayoutEffect, useRef, useState } from 'react'

interface Params {
  topRef: RefObject<HTMLElement | null>        // the in-flow filter block (type switcher + Add)
  cardRef: RefObject<HTMLElement | null>       // mobile: a real rendered card
  rowRef: RefObject<HTMLElement | null>        // desktop: a real rendered table row
  wrapRef: RefObject<HTMLElement | null>       // desktop: the table wrap (thead + tbody)
  paginationRef: RefObject<HTMLElement | null> // bottom boundary (both)
  gap?: number
  extra?: number
  navId?: string
  min?: number
  max?: number
}

interface Fit {
  size: number        // rows per page
  frameHeight: number // px to reserve for the row area on non-last pages (0 = natural)
}

// Rows-per-page sized to fill one screenful below the filter block, measured
// entirely from the DOM. The effect runs every render but latches after the first
// successful measurement (measuredRef): a cold refresh loads behind a loading gate,
// so measuring at mount would hit null refs and fall back to the default.
//
// Nothing is pinned/sticky. floor() guarantees the block (gap + header + rows +
// pagination) is <= one viewport, so once the summary cards scroll out the filter
// reaches the top and the page simply can't scroll past it — no floating needed.
//
//   rows = floor((viewport − filter − gapBelow − pagination) / rowH)  (desktop)
//
// frameHeight = header + rows*rowH is the height a table wrap reserves on every page
// EXCEPT the last, so the header and pagination sit at the exact same Y across page
// switches even if a row's height changes. The last page uses natural height so the
// pagination hugs the last row (no dead space).
export function useFitPageSize({
  topRef, cardRef, rowRef, wrapRef, paginationRef,
  extra = 8, navId = 'khp-mobile-nav', min = 4, max = 20,
}: Params): Fit {
  const [fit, setFit] = useState<Fit>({ size: 8, frameHeight: 0 })
  const measuredRef = useRef(false)

  useLayoutEffect(() => {
    if (measuredRef.current) return
    const clamp = (n: number) => Math.max(min, Math.min(max, n))
    const isMobile = window.innerWidth < 768
    const pagination = paginationRef.current?.offsetHeight ?? 0

    if (isMobile && topRef.current && cardRef.current) {
      const topBar = topRef.current.offsetHeight
      const cardH = cardRef.current.offsetHeight || 80
      const nav = document.getElementById(navId)?.offsetHeight ?? 0
      const inside = window.innerHeight - topBar - pagination - nav - extra
      const rows = clamp(Math.floor((inside + 8) / (cardH + 8)))
      setFit({ size: rows, frameHeight: 0 }) // mobile: natural vertical scroll, no reserve
      measuredRef.current = true
    } else if (!isMobile && rowRef.current && wrapRef.current) {
      const wrapTop = wrapRef.current.getBoundingClientRect().top
      const rowTop = rowRef.current.getBoundingClientRect().top
      const rowH = rowRef.current.offsetHeight || 52
      const theadH = Math.max(0, rowTop - wrapTop)
      const filterRect = topRef.current?.getBoundingClientRect()
      const filterH = filterRect?.height ?? 0
      const gapBelow = filterRect ? Math.max(0, wrapTop - filterRect.bottom) : 0
      const available = window.innerHeight - filterH - gapBelow - pagination - extra
      const rows = clamp(Math.floor((available - theadH) / rowH))
      setFit({ size: rows, frameHeight: theadH + rows * rowH })
      measuredRef.current = true
    }
    // No dep array: retry each render until the table mounts, then the latch stops it.
  })

  return fit
}
