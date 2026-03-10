import { useEffect, useMemo, useState } from 'react'
import { buildCoverProxyURL } from '../lib/api'

type ComicCoverProps = {
  comicId: number
  title: string
  coverURL?: string
  coverLocalRelPath?: string
  className: string
  loading?: 'eager' | 'lazy'
}

export default function ComicCover({ comicId, title, coverURL = '', coverLocalRelPath = '', className, loading = 'lazy' }: ComicCoverProps) {
  const [failed, setFailed] = useState(false)

  const hasCover = useMemo(() => {
    return coverURL.trim() !== '' || coverLocalRelPath.trim() !== ''
  }, [coverLocalRelPath, coverURL])

  const src = useMemo(() => {
    if (!hasCover) {
      return ''
    }
    return buildCoverProxyURL(comicId)
  }, [comicId, hasCover])

  useEffect(() => {
    setFailed(false)
  }, [src])

  return (
    <div className={className}>
      {src !== '' && !failed ? (
        <img loading={loading} src={src} alt={title} onError={() => setFailed(true)} />
      ) : (
        <span className="cover-fallback">{title.slice(0, 1) || '漫'}</span>
      )}
    </div>
  )
}
