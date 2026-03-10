export default function DownloadsPage() {
  return (
    <section className="page">
      <header className="page-header">
        <div>
          <h1>下载管理</h1>
          <p>下载进行中 / 已完成双列表会在后续阶段接入真实任务队列。</p>
        </div>
      </header>
      <div className="placeholder-grid">
        <article className="panel">
          <div className="panel-head"><h2>进行中</h2><span>排队 / 下载中 / 校验中</span></div>
          <p>当前还未接入任务后端。</p>
        </article>
        <article className="panel">
          <div className="panel-head"><h2>已完成</h2><span>可清理、可导出</span></div>
          <p>当前还未接入任务后端。</p>
        </article>
      </div>
    </section>
  )
}
