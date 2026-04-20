/* ======================================================================
   OrbitLens Reader Engine
   Reads book config from <script id="book-config"> and powers the reader.
   ====================================================================== */

(function () {
  'use strict';

  const cfg = JSON.parse(document.getElementById('book-config').textContent);
  const CACHE_BUST = '?t=' + Date.now();
  const LS_THEME = 'orbit-reader-theme';
  const LS_FONT = 'orbit-reader-font';
  const LS_SCROLL = (slug) => `orbit-reader-scroll-${cfg.bookId}-${slug}`;

  // Sibling books for the cover-page switcher (order = L1 → L3)
  const BOOKS = [
    { id: 'psychological-os',    code: 'BOOK · 1', layer: 'L1 / INTERIOR',  title: '心理OS',                            path: '../psychological-os/' },
    { id: 'structure-driven-org', code: 'BOOK · 2', layer: 'L2 / STRUCTURE', title: '構造駆動エンジニアリング組織論', path: '../structure-driven-org/' },
    { id: 'git-archaeology',     code: 'BOOK · 3', layer: 'L3 / STRATA',    title: 'git考古学',                          path: '../git-archaeology/' }
  ];

  // Restore preferences
  const savedTheme = localStorage.getItem(LS_THEME);
  if (savedTheme === 'light') document.body.classList.add('theme-light');
  const savedFont = localStorage.getItem(LS_FONT);
  if (savedFont) document.documentElement.style.setProperty('--font-size', savedFont);

  // ---------- Build DOM ----------
  document.title = cfg.title + ' — OrbitLens';

  const root = document.getElementById('reader-root');
  root.innerHTML = `
    <div class="starfield"></div>

    <nav class="orbit-bar">
      <a href="../" class="crumb-logo">
        <img src="../images/orbitlens-royal-purple-shadow.png" alt="">
        <span class="crumb-brand">ORBITLENS</span>
      </a>
      <span class="crumb-sep">/</span>
      <a href="../" class="crumb-scope">LIBRARY</a>
      <span class="crumb-sep">/</span>
      <span class="crumb-book" id="crumb-book">${escapeHTML(cfg.title)}</span>
      <span class="private-chip" title="Read traces, not people">
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="3" y="7" width="10" height="7" rx="1"/>
          <path d="M5 7V5a3 3 0 016 0v2"/>
        </svg>
        ${cfg.lang === 'ja' ? '観測' : 'Observe'}
      </span>
      <span class="orbit-spacer"></span>
      <button class="orbit-menu-toggle" id="orbit-menu-toggle" aria-label="Menu" aria-expanded="false">☰</button>
      <div class="orbit-tools" id="orbit-tools">
        <button class="orbit-tool" id="tool-theme" data-closes-menu title="${cfg.lang === 'ja' ? 'テーマ切替 (T)' : 'Toggle theme (T)'}">
          <span class="tool-icon">◐</span>
          <span class="tool-label">${cfg.lang === 'ja' ? 'テーマ' : 'Theme'}</span>
        </button>
        <button class="orbit-tool" id="tool-font-s" title="${cfg.lang === 'ja' ? '文字を小さく' : 'Smaller text'}">
          <span class="tool-icon">A−</span>
          <span class="tool-label">${cfg.lang === 'ja' ? '小さく' : 'Smaller'}</span>
        </button>
        <button class="orbit-tool" id="tool-font-l" title="${cfg.lang === 'ja' ? '文字を大きく' : 'Larger text'}">
          <span class="tool-icon">A+</span>
          <span class="tool-label">${cfg.lang === 'ja' ? '大きく' : 'Larger'}</span>
        </button>
        <button class="orbit-tool" id="tool-help" data-closes-menu title="${cfg.lang === 'ja' ? 'ショートカット (?)' : 'Shortcuts (?)'}">
          <span class="tool-icon">?</span>
          <span class="tool-label">${cfg.lang === 'ja' ? 'ショートカット' : 'Shortcuts'}</span>
        </button>
        <a class="orbit-tool" id="tool-en" data-closes-menu href="${cfg.chapters[0].en}" target="_blank" rel="noopener">
          <span class="tool-icon">EN ↗</span>
          <span class="tool-label">${cfg.lang === 'ja' ? '英語版' : 'English'}</span>
        </a>
      </div>
      <div class="progress-track" id="progress-track" role="slider" tabindex="0" aria-label="${cfg.lang === 'ja' ? 'ページ位置' : 'Page position'}">
        <div class="progress-markers" id="progress-markers"></div>
        <div class="progress-bar" id="progress-bar"></div>
        <div class="progress-thumb" id="progress-thumb" aria-hidden="true"></div>
      </div>
    </nav>

    <div class="reader-layout">
      <aside class="book-toc-sidebar">
        <a href="#" class="sidebar-home" id="sidebar-home">← ${cfg.lang === 'ja' ? '目次' : 'Contents'}</a>
        <div class="sidebar-label">${cfg.lang === 'ja' ? '全' + cfg.chapters.length + '章' : cfg.chapters.length + ' chapters'}</div>
        <ul class="sidebar-chapters" id="sidebar-chapters">
          ${cfg.chapters.map((c, i) => `
            <li><a href="#${c.slug}" data-slug="${c.slug}">
              <span class="ch-num">${chapterNumLabel(i, cfg.chapters.length, cfg.lang)}</span>
              <span class="ch-title">${escapeHTML(c.title)}</span>
            </a></li>`).join('')}
        </ul>
      </aside>

      <main class="reader-main" id="reader-main"></main>

      <aside class="inline-toc" id="inline-toc" style="display:none">
        <div class="inline-toc-label">${cfg.lang === 'ja' ? 'この章の目次' : 'On this page'}</div>
        <ol id="inline-toc-list"></ol>
      </aside>
    </div>

    <button class="mobile-toc-toggle" id="mobile-toc-toggle" aria-label="Chapters">☰</button>
    <div class="mobile-toc-drawer" id="mobile-toc-drawer">
      <div class="mobile-toc-panel">
        <div class="mobile-toc-handle"></div>
        <div class="sidebar-label">${cfg.lang === 'ja' ? '章目次' : 'Chapters'}</div>
        <ul class="sidebar-chapters" id="mobile-chapters">
          ${cfg.chapters.map((c, i) => `
            <li><a href="#${c.slug}" data-slug="${c.slug}">
              <span class="ch-num">${chapterNumLabel(i, cfg.chapters.length, cfg.lang)}</span>
              <span class="ch-title">${escapeHTML(c.title)}</span>
            </a></li>`).join('')}
        </ul>
      </div>
    </div>

    <div class="shortcut-overlay" id="shortcut-overlay">
      <div class="shortcut-panel">
        <h3>${cfg.lang === 'ja' ? 'キーボードショートカット' : 'Keyboard shortcuts'}</h3>
        <dl>
          <dt>j / ↓</dt><dd>${cfg.lang === 'ja' ? '次のセクション' : 'Next section'}</dd>
          <dt>k / ↑</dt><dd>${cfg.lang === 'ja' ? '前のセクション' : 'Previous section'}</dd>
          <dt>n / →</dt><dd>${cfg.lang === 'ja' ? '次の章' : 'Next chapter'}</dd>
          <dt>p / ←</dt><dd>${cfg.lang === 'ja' ? '前の章' : 'Previous chapter'}</dd>
          <dt>h</dt><dd>${cfg.lang === 'ja' ? '目次へ' : 'Home / contents'}</dd>
          <dt>t</dt><dd>${cfg.lang === 'ja' ? 'テーマ切替' : 'Toggle theme'}</dd>
          <dt>+ / −</dt><dd>${cfg.lang === 'ja' ? '文字サイズ' : 'Font size'}</dd>
          <dt>?</dt><dd>${cfg.lang === 'ja' ? 'この画面' : 'This panel'}</dd>
        </dl>
        <div class="close-hint">Esc · ${cfg.lang === 'ja' ? '閉じる' : 'Close'}</div>
      </div>
    </div>

    <div class="toast" id="toast"></div>

    <div class="firewall">
      <div class="firewall-msg">
        <svg class="firewall-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="3" y="7" width="10" height="7" rx="1"/>
          <path d="M5 7V5a3 3 0 016 0v2"/>
        </svg>
        <span class="firewall-tag">Firewall</span>
        <strong>${cfg.lang === 'ja' ? '観測であって、評価ではない。' : 'Observation, not judgment.'}</strong>
        <span>${cfg.lang === 'ja' ? '個人の点数はこの層から漏らさない。' : 'No personal scores leak past this layer.'}</span>
      </div>
      <span style="font-family: var(--mono); font-size: 10px; letter-spacing: 1.5px; color: var(--dim)">T-30 · UX-12</span>
    </div>
  `;

  // ---------- Refs ----------
  const main = document.getElementById('reader-main');
  const inlineToc = document.getElementById('inline-toc');
  const inlineTocList = document.getElementById('inline-toc-list');
  const progressBar = document.getElementById('progress-bar');
  const sidebarChapters = document.getElementById('sidebar-chapters');
  const mobileChapters = document.getElementById('mobile-chapters');
  const crumbBook = document.getElementById('crumb-book');
  const mobileToggle = document.getElementById('mobile-toc-toggle');
  const mobileDrawer = document.getElementById('mobile-toc-drawer');
  const shortcutOverlay = document.getElementById('shortcut-overlay');
  const toolEn = document.getElementById('tool-en');
  const toastEl = document.getElementById('toast');

  // ---------- Markdown setup ----------
  const renderer = new marked.Renderer();
  const originalImage = renderer.image.bind(renderer);
  renderer.image = (href, title, text) => {
    const rewrite = (h) => {
      if (h.startsWith('./images/') || h.startsWith('images/')) {
        return cfg.rawBase + h.replace(/^\.\//, '') + CACHE_BUST;
      }
      return h;
    };
    if (typeof href === 'object' && href.href) {
      href.href = rewrite(href.href);
      return originalImage(href);
    }
    if (typeof href === 'string') {
      return originalImage(rewrite(href), title, text);
    }
    return originalImage(href, title, text);
  };
  marked.use({ renderer, gfm: true, breaks: false });

  // ---------- Utils ----------
  function escapeHTML(s) {
    return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
  }
  function slugify(s) {
    return String(s).toLowerCase().trim()
      .replace(/[\s—–·・]+/g, '-')
      .replace(/[^\w\-ぁ-んァ-ン一-龯]/g, '')
      .replace(/-+/g, '-').replace(/^-|-$/g, '');
  }
  function chapterNumLabel(i, total, lang) {
    // ja: 序章, 第1章 ... last is 最終章 if labeled epilogue
    const slug = cfg.chapters[i].slug;
    if (slug === 'epilogue') return lang === 'ja' ? '最終章' : 'Epilogue';
    if (i === 0) return lang === 'ja' ? '序章' : 'Prologue';
    if (lang === 'ja') return '第' + i + '章';
    return 'Ch ' + String(i).padStart(2, '0');
  }
  function showToast(msg) {
    toastEl.textContent = msg;
    toastEl.classList.add('show');
    clearTimeout(toastEl._t);
    toastEl._t = setTimeout(() => toastEl.classList.remove('show'), 1600);
  }
  function estimateReadingMins(body) {
    const chars = body.length;
    // Japanese: ~600 chars/min; English mixed: ~900 chars/min
    const rate = cfg.lang === 'ja' ? 600 : 900;
    return Math.max(1, Math.round(chars / rate));
  }

  // ---------- Core: load chapter ----------
  let activeObserver = null;
  let currentSlug = null;

  async function renderHome() {
    inlineToc.style.display = 'none';
    crumbBook.textContent = cfg.title;
    // Clear any leftover progress markers from a previously rendered chapter
    const markersEl = document.getElementById('progress-markers');
    if (markersEl) markersEl.innerHTML = '';
    tocHeadings = [];
    tocLinkMap = new Map();

    let html = `
      <section class="book-cover-section">
        <div class="book-eyebrow">${cfg.lang === 'ja' ? cfg.bookRomanNum + ' · 観測対象 ' + cfg.layerLabel : 'BOOK ' + cfg.bookRomanNum + ' · LAYER ' + cfg.layerLabel}</div>
        <div class="book-code-hero">${cfg.bookCode || cfg.bookRomanNum}</div>
        <h1 class="book-title-main">${escapeHTML(cfg.title)}</h1>
        <div class="book-subtitle-main">${escapeHTML(cfg.subtitle || '')}</div>
        <div class="book-summary-main">${escapeHTML(cfg.summary || '')}</div>
        <div class="book-meta">
          <span class="book-meta-item"><span class="book-meta-dot"></span><strong>${cfg.chapters.length}</strong> ${cfg.lang === 'ja' ? '章' : 'chapters'}</span>
          <span class="book-meta-item"><span class="book-meta-dot"></span>JP · ${cfg.lang === 'ja' ? '日本語原文' : 'Japanese original'}</span>
          <a class="book-meta-item book-meta-link" href="${cfg.chapters[0].en}" target="_blank" rel="noopener"><span class="book-meta-dot"></span>EN · dev.to ↗</a>
          <a class="book-meta-item book-meta-link" href="https://github.com/machuz/eis" target="_blank" rel="noopener">github.com/machuz/eis ↗</a>
        </div>
      </section>

      <div class="compass-divider">
        <span class="line"></span>
        <svg viewBox="0 0 22 22" fill="none">
          <path d="M11 2 L13 11 L11 9 L9 11 Z" fill="currentColor"/>
          <path d="M11 20 L9 11 L11 13 L13 11 Z" fill="currentColor" opacity="0.5"/>
          <path d="M2 11 L11 9 L9 11 L11 13 Z" fill="currentColor" opacity="0.5"/>
          <path d="M20 11 L11 13 L13 11 L11 9 Z" fill="currentColor" opacity="0.5"/>
          <circle cx="11" cy="11" r="1" fill="currentColor"/>
        </svg>
        <span class="line"></span>
      </div>

      <div class="toc-grid">
        <div class="toc-heading">
          <span>${cfg.lang === 'ja' ? '目次' : 'Table of Contents'}</span>
          <span class="toc-total">${cfg.chapters.length} ${cfg.lang === 'ja' ? '章' : 'chapters'}</span>
        </div>
        <ul class="toc-list">
          ${cfg.chapters.map((c, i) => `
            <li><a href="#${c.slug}">
              <span class="toc-num">${chapterNumLabel(i, cfg.chapters.length, cfg.lang)}</span>
              <span class="toc-title-text">${escapeHTML(c.title)}</span>
              <span class="toc-en-badge" onclick="event.stopPropagation();window.open('${c.en}','_blank');return false;" role="button">EN ↗</span>
            </a></li>`).join('')}
        </ul>
      </div>

      <section class="book-switch">
        <div class="book-switch-label">${cfg.lang === 'ja' ? '他の Book' : 'Other books'}</div>
        <div class="book-switch-list">
          ${BOOKS.map(b => `
            <a href="${b.path}" class="book-switch-item${b.id === cfg.bookId ? ' current' : ''}"${b.id === cfg.bookId ? ' aria-current="page"' : ''}>
              <span class="bs-code">${b.code}</span>
              <span class="bs-layer">${b.layer}</span>
              <span class="bs-title">${escapeHTML(b.title)}</span>
            </a>
          `).join('')}
        </div>
      </section>

      <div class="book-footer">
        <div class="quote">${cfg.closingQuote || ''}</div>
        <div class="foot-links">
          <a href="https://github.com/machuz/eis/tree/main/books/${cfg.bookId}" target="_blank" rel="noopener">GitHub</a>
          <a href="../">${cfg.lang === 'ja' ? '図書館へ' : 'Library'}</a>
          <a href="${cfg.chapters[0].en}" target="_blank" rel="noopener">EN · dev.to</a>
        </div>
      </div>
    `;

    main.innerHTML = html;
    clearActiveSidebar();
    document.title = cfg.title + ' — OrbitLens';
    window.scrollTo({ top: 0, behavior: 'auto' });
    toolEn.href = cfg.chapters[0].en;
  }

  async function loadChapter(slug) {
    const idx = cfg.chapters.findIndex(c => c.slug === slug);
    if (idx === -1) { renderHome(); return; }
    const chapter = cfg.chapters[idx];

    main.innerHTML = `<div class="chapter-loading">${cfg.lang === 'ja' ? '読み込み中…' : 'Loading…'}</div>`;

    try {
      const res = await fetch(cfg.rawBase + slug + '.md' + CACHE_BUST);
      if (!res.ok) throw new Error('fetch failed');
      const md = await res.text();
      // Strip frontmatter and the leading H1 (we render the title ourselves above the article)
      const body = md
        .replace(/^---[\s\S]*?---\s*/, '')
        .replace(/^\s*#\s+[^\n]+\n+/, '');
      const readMins = estimateReadingMins(body);
      const prev = idx > 0 ? cfg.chapters[idx - 1] : null;
      const next = idx < cfg.chapters.length - 1 ? cfg.chapters[idx + 1] : null;

      const barHtml = `
        <div class="chapter-bar">
          <div class="ch-info">
            <strong>${chapterNumLabel(idx, cfg.chapters.length, cfg.lang)}</strong>
            <span>${idx + 1} / ${cfg.chapters.length}</span>
            <span class="ch-reading-time">≈ ${readMins} ${cfg.lang === 'ja' ? '分' : 'min'}</span>
          </div>
          <a href="${chapter.en}" target="_blank" rel="noopener" class="en-link">EN · dev.to ↗</a>
        </div>
      `;

      const mobileSectionsHtml = `
        <details class="mobile-section-list" id="mobile-section-list">
          <summary>
            <span class="msl-label">${cfg.lang === 'ja' ? 'この章の目次' : 'On this page'}</span>
            <span class="msl-current" id="msl-current"></span>
            <span class="msl-chevron" aria-hidden="true">▾</span>
          </summary>
          <ol class="mobile-section-items" id="mobile-section-items"></ol>
        </details>
      `;

      const rendered = marked.parse(body);
      const articleHtml = `<article class="chapter-content">${rendered}</article>`;

      const navHtml = `
        <nav class="chapter-nav">
          <div class="prev">${prev ? `<a href="#${prev.slug}"><span class="nav-label">← ${cfg.lang === 'ja' ? '前章' : 'Previous'}</span>${escapeHTML(prev.title)}</a>` : ''}</div>
          <a class="home" href="#">${cfg.lang === 'ja' ? '目次' : 'Contents'}</a>
          <div class="next">${next ? `<a href="#${next.slug}"><span class="nav-label">${cfg.lang === 'ja' ? '次章' : 'Next'} →</span>${escapeHTML(next.title)}</a>` : ''}</div>
        </nav>
      `;

      const titleHtml = `<h1 class="chapter-title-display">${escapeHTML(chapter.title)}</h1>`;
      main.innerHTML = barHtml + titleHtml + mobileSectionsHtml + articleHtml + navHtml;

      // Auto-close mobile section accordion after tap
      const msl = main.querySelector('.mobile-section-list');
      if (msl) {
        msl.addEventListener('click', (e) => {
          if (e.target.closest('a')) {
            setTimeout(() => msl.removeAttribute('open'), 150);
          }
        });
      }

      // Enhance rendered content
      enhanceCodeBlocks();
      enhanceHeadings();
      linkifyCrossRefs();
      buildInlineTOC();
      setActiveSidebar(slug);
      currentSlug = slug;
      toolEn.href = chapter.en;
      crumbBook.textContent = chapter.title;
      document.title = chapter.title + ' — ' + cfg.title;

      // Restore scroll
      const savedScroll = Number(sessionStorage.getItem(LS_SCROLL(slug)) || 0);
      requestAnimationFrame(() => {
        window.scrollTo({ top: savedScroll || 0, behavior: 'auto' });
      });
    } catch (e) {
      console.error(e);
      main.innerHTML = `<div class="chapter-loading">${cfg.lang === 'ja' ? '読み込みに失敗しました。' : 'Failed to load.'}</div>`;
    }
  }

  function clearActiveSidebar() {
    sidebarChapters.querySelectorAll('a').forEach(a => a.classList.remove('current'));
    mobileChapters.querySelectorAll('a').forEach(a => a.classList.remove('current'));
  }
  function setActiveSidebar(slug) {
    clearActiveSidebar();
    [sidebarChapters, mobileChapters].forEach(list => {
      const link = list.querySelector(`a[data-slug="${slug}"]`);
      if (link) link.classList.add('current');
    });
  }

  function enhanceCodeBlocks() {
    main.querySelectorAll('pre').forEach((pre) => {
      const btn = document.createElement('button');
      btn.className = 'copy-btn';
      btn.textContent = 'Copy';
      btn.addEventListener('click', async () => {
        try {
          const code = pre.querySelector('code')?.textContent || pre.textContent;
          await navigator.clipboard.writeText(code);
          btn.textContent = '✓ Copied';
          btn.classList.add('copied');
          setTimeout(() => {
            btn.textContent = 'Copy';
            btn.classList.remove('copied');
          }, 1400);
        } catch {}
      });
      pre.appendChild(btn);
    });
  }
  function enhanceHeadings() {
    let h2idx = 0;
    main.querySelectorAll('article.chapter-content h2, article.chapter-content h3').forEach((h) => {
      const text = h.textContent;
      const id = h.id || slugify(text) || 'h-' + Math.random().toString(36).slice(2, 7);
      h.id = id;
      if (h.tagName === 'H2') {
        h2idx++;
        h.setAttribute('data-section', String(h2idx));
      }
      const anchor = document.createElement('a');
      anchor.className = 'heading-anchor';
      anchor.href = '#' + id;
      anchor.textContent = '§';
      anchor.setAttribute('aria-label', 'Link to section');
      anchor.addEventListener('click', (e) => {
        e.preventDefault();
        history.replaceState(null, '', '#' + (window.location.hash.slice(1) || '') + '-' + id);
        h.scrollIntoView({ behavior: 'smooth' });
      });
      h.prepend(anchor);
    });
  }

  // Linkify in-text cross-references like "ch3 §5" or "第3章 §5"
  // to #chN-sM hashes that jump-load the target chapter + scroll to the Nth h2.
  function linkifyCrossRefs() {
    const article = main.querySelector('article.chapter-content');
    if (!article) return;
    const walker = document.createTreeWalker(article, NodeFilter.SHOW_TEXT, {
      acceptNode: (node) => {
        const p = node.parentElement;
        if (!p || p.closest('a, code, pre, h1, h2, h3, h4')) return NodeFilter.FILTER_REJECT;
        return /(ch|第)\s*\d+/i.test(node.nodeValue) ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_REJECT;
      }
    });
    const nodes = [];
    let n; while ((n = walker.nextNode())) nodes.push(n);
    // Matches: ch3 §5 / ch3§5 / ch3 / 第3章 §5 / 第3章§5 / 第3章
    const re = /(ch(\d+)|第(\d+)章)(?:\s*§\s*(\d+))?/gi;
    nodes.forEach((node) => {
      const text = node.nodeValue;
      if (!re.test(text)) return;
      re.lastIndex = 0;
      const frag = document.createDocumentFragment();
      let lastIdx = 0;
      let m;
      while ((m = re.exec(text)) !== null) {
        if (m.index > lastIdx) frag.appendChild(document.createTextNode(text.slice(lastIdx, m.index)));
        const chNum = m[2] || m[3];
        const sNum = m[4];
        const chSlug = 'ch' + chNum;
        // Only linkify if the referenced chapter exists in this book
        if (cfg.chapters.some(c => c.slug === chSlug) && chSlug !== currentSlug) {
          const a = document.createElement('a');
          a.className = 'xref';
          a.textContent = m[0];
          a.href = '#' + chSlug + (sNum ? '-s' + sNum : '');
          frag.appendChild(a);
        } else if (sNum && chSlug === currentSlug) {
          // Same-chapter section reference
          const a = document.createElement('a');
          a.className = 'xref';
          a.textContent = m[0];
          a.href = '#' + chSlug + '-s' + sNum;
          frag.appendChild(a);
        } else {
          frag.appendChild(document.createTextNode(m[0]));
        }
        lastIdx = m.index + m[0].length;
      }
      if (lastIdx < text.length) frag.appendChild(document.createTextNode(text.slice(lastIdx)));
      node.parentNode.replaceChild(frag, node);
    });
  }

  let tocHeadings = [];
  let tocLinkMap = new Map();

  function buildInlineTOC() {
    const heads = Array.from(main.querySelectorAll('article.chapter-content h2, article.chapter-content h3'));
    if (!heads.length) {
      inlineToc.style.display = 'none';
      tocHeadings = [];
      const mk = document.getElementById('progress-markers');
      if (mk) mk.innerHTML = '';
      return;
    }
    inlineTocList.innerHTML = heads.map((h) => {
      const lvl = h.tagName === 'H3' ? 3 : 2;
      const text = h.textContent.replace(/^§/, '').trim();
      return `<li><a href="#${h.id}" class="level-${lvl}">${escapeHTML(text)}</a></li>`;
    }).join('');
    inlineToc.style.display = '';

    tocHeadings = heads;
    tocLinkMap = new Map();
    inlineTocList.querySelectorAll('a').forEach((a) => {
      tocLinkMap.set(a.getAttribute('href').slice(1), a);
      a.addEventListener('click', (e) => {
        e.preventDefault();
        const id = a.getAttribute('href').slice(1);
        const el = document.getElementById(id);
        if (el) el.scrollIntoView({ behavior: 'smooth' });
      });
    });

    // Populate mobile section list (Zenn-style "この章の目次")
    const mobileItems = document.getElementById('mobile-section-items');
    if (mobileItems) {
      mobileItems.innerHTML = heads.map((h) => {
        const lvl = h.tagName === 'H3' ? 3 : 2;
        const text = h.textContent.replace(/^§/, '').trim();
        return `<li><a href="#${h.id}" class="level-${lvl}">${escapeHTML(text)}</a></li>`;
      }).join('');
    }

    // Populate progress markers — one tick per H2
    const markers = document.getElementById('progress-markers');
    if (markers) {
      const h2Heads = heads.filter(h => h.tagName === 'H2');
      const doc = document.documentElement;
      const scrollMax = doc.scrollHeight - doc.clientHeight;
      markers.innerHTML = h2Heads.map((h) => {
        const top = h.getBoundingClientRect().top + window.scrollY;
        const pct = scrollMax > 0 ? Math.min(100, (top / scrollMax) * 100) : 0;
        return `<span class="progress-marker" style="left: ${pct}%" data-target="${h.id}"></span>`;
      }).join('');
    }

    updateActiveHeading();
  }

  function updateActiveHeading() {
    if (!tocHeadings.length) return;
    const probe = window.scrollY + window.innerHeight * 0.25;
    let current = tocHeadings[0];
    for (const h of tocHeadings) {
      const top = h.getBoundingClientRect().top + window.scrollY;
      if (top <= probe) current = h;
      else break;
    }
    // At page bottom: force last
    if ((window.innerHeight + window.scrollY) >= (document.documentElement.scrollHeight - 8)) {
      current = tocHeadings[tocHeadings.length - 1];
    }
    tocLinkMap.forEach(a => a.classList.remove('active'));
    const active = tocLinkMap.get(current.id);
    if (active) {
      active.classList.add('active');
      // Keep active entry in view within the TOC column
      const rect = active.getBoundingClientRect();
      const parentRect = inlineToc.getBoundingClientRect();
      if (rect.top < parentRect.top + 20 || rect.bottom > parentRect.bottom - 20) {
        active.scrollIntoView({ block: 'nearest' });
      }
    }
    // Mobile section-list summary label reflects the current H2
    const mslCurrent = document.getElementById('msl-current');
    if (mslCurrent && current && current.tagName === 'H2') {
      mslCurrent.textContent = '· ' + current.textContent.replace(/^§/, '').trim();
    }
  }

  // ---------- Progress bar + scroll save ----------
  const progressTrack = document.getElementById('progress-track');
  const progressThumb = document.getElementById('progress-thumb');
  let progressRaf = null;
  function updateProgress() {
    if (progressRaf) return;
    progressRaf = requestAnimationFrame(() => {
      progressRaf = null;
      const doc = document.documentElement;
      const scrollTop = window.scrollY;
      const scrollHeight = doc.scrollHeight - doc.clientHeight;
      const pct = scrollHeight > 0 ? (scrollTop / scrollHeight) * 100 : 0;
      const clamped = Math.min(100, Math.max(0, pct));
      progressBar.style.width = clamped + '%';
      if (progressThumb) progressThumb.style.left = clamped + '%';
      // Save scroll position per chapter
      const slug = decodeURIComponent(window.location.hash.slice(1));
      if (slug) sessionStorage.setItem(LS_SCROLL(slug), String(scrollTop));
    });
  }
  window.addEventListener('scroll', updateProgress, { passive: true });
  window.addEventListener('scroll', updateActiveHeading, { passive: true });
  window.addEventListener('resize', updateActiveHeading);
  // Recalculate progress markers when layout changes (resize, font-size, etc.)
  let resizeMarkersTimer = null;
  window.addEventListener('resize', () => {
    clearTimeout(resizeMarkersTimer);
    resizeMarkersTimer = setTimeout(() => {
      if (document.querySelector('article.chapter-content')) buildInlineTOC();
    }, 120);
  });

  // Progress track — draggable scrubber
  let isDragging = false;
  function trackScrollFromClientX(clientX) {
    if (!progressTrack) return;
    const rect = progressTrack.getBoundingClientRect();
    const ratio = Math.max(0, Math.min(1, (clientX - rect.left) / rect.width));
    const doc = document.documentElement;
    const target = ratio * (doc.scrollHeight - doc.clientHeight);
    window.scrollTo({ top: target, behavior: isDragging ? 'auto' : 'smooth' });
  }
  if (progressTrack) {
    progressTrack.addEventListener('pointerdown', (e) => {
      isDragging = true;
      try { progressTrack.setPointerCapture(e.pointerId); } catch (_) {}
      progressTrack.classList.add('dragging');
      trackScrollFromClientX(e.clientX);
    });
    progressTrack.addEventListener('pointermove', (e) => {
      if (!isDragging) return;
      trackScrollFromClientX(e.clientX);
    });
    progressTrack.addEventListener('pointerup', () => {
      isDragging = false;
      progressTrack.classList.remove('dragging');
    });
    progressTrack.addEventListener('pointercancel', () => {
      isDragging = false;
      progressTrack.classList.remove('dragging');
    });
  }

  // ---------- Hash routing ----------
  function scrollToSection(idx) {
    const h2 = main.querySelector(`article.chapter-content h2[data-section="${idx}"]`);
    if (h2) h2.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
  function parseHash(raw) {
    if (!raw) return { slug: null, sectionIdx: null };
    // #chN-sM or #chN or #epilogue-sM or #epilogue
    const m = raw.match(/^(ch\d+|epilogue)(?:-s(\d+))?/);
    if (m) return { slug: m[1], sectionIdx: m[2] ? parseInt(m[2], 10) : null };
    // Legacy: #chN-<slug> where slug is a heading slug — match chapter prefix
    const chMatch = raw.match(/^(ch\d+|epilogue)/);
    if (chMatch) return { slug: chMatch[1], sectionIdx: null };
    return { slug: null, sectionIdx: null };
  }
  async function handleHashChange() {
    const raw = decodeURIComponent(window.location.hash.slice(1));
    const { slug, sectionIdx } = parseHash(raw);
    if (!slug) { renderHome(); return; }
    const chapter = cfg.chapters.find(c => c.slug === slug);
    if (!chapter) { renderHome(); return; }

    if (currentSlug !== slug) {
      await loadChapter(slug);
    }
    if (sectionIdx) {
      requestAnimationFrame(() => scrollToSection(sectionIdx));
    }
  }
  window.addEventListener('hashchange', handleHashChange);

  // ---------- Toolbar actions ----------
  document.getElementById('tool-theme').addEventListener('click', toggleTheme);
  document.getElementById('tool-font-s').addEventListener('click', () => adjustFont(-1));
  document.getElementById('tool-font-l').addEventListener('click', () => adjustFont(+1));
  document.getElementById('tool-help').addEventListener('click', () => shortcutOverlay.classList.add('open'));
  shortcutOverlay.addEventListener('click', (e) => {
    if (e.target === shortcutOverlay) shortcutOverlay.classList.remove('open');
  });

  // Mobile hamburger — stores font-size buttons so A-/A+ can be tapped repeatedly.
  // Theme / help / EN carry data-closes-menu and auto-close on tap.
  const orbitMenuToggle = document.getElementById('orbit-menu-toggle');
  const orbitTools = document.getElementById('orbit-tools');
  function closeOrbitMenu() {
    if (!orbitTools) return;
    orbitTools.classList.remove('open');
    if (orbitMenuToggle) orbitMenuToggle.setAttribute('aria-expanded', 'false');
  }
  if (orbitMenuToggle && orbitTools) {
    orbitMenuToggle.addEventListener('click', (e) => {
      e.stopPropagation();
      const opened = orbitTools.classList.toggle('open');
      orbitMenuToggle.setAttribute('aria-expanded', opened ? 'true' : 'false');
    });
    orbitTools.addEventListener('click', (e) => {
      const el = e.target.closest('[data-closes-menu]');
      if (el) closeOrbitMenu();
    });
    document.addEventListener('click', (e) => {
      if (!orbitTools.classList.contains('open')) return;
      if (e.target.closest('.orbit-tools, .orbit-menu-toggle')) return;
      closeOrbitMenu();
    });
  }

  function toggleTheme() {
    const light = document.body.classList.toggle('theme-light');
    localStorage.setItem(LS_THEME, light ? 'light' : 'dark');
    showToast(light ? 'Light theme' : 'Dark theme');
  }

  function adjustFont(delta) {
    const sizes = [14, 15, 16, 17, 18, 19, 20, 22];
    const cur = parseInt(getComputedStyle(document.documentElement).getPropertyValue('--font-size')) || 17;
    let idx = sizes.indexOf(cur);
    if (idx === -1) idx = sizes.indexOf(17);
    const next = sizes[Math.max(0, Math.min(sizes.length - 1, idx + delta))];
    document.documentElement.style.setProperty('--font-size', next + 'px');
    localStorage.setItem(LS_FONT, next + 'px');
    showToast(next + 'px');
  }

  // ---------- Keyboard ----------
  document.addEventListener('keydown', (e) => {
    if (e.target.matches('input, textarea')) return;
    if (e.metaKey || e.ctrlKey || e.altKey) return;

    if (shortcutOverlay.classList.contains('open')) {
      if (e.key === 'Escape' || e.key === '?') shortcutOverlay.classList.remove('open');
      return;
    }

    const curSlug = decodeURIComponent(window.location.hash.slice(1));
    const curChapter = cfg.chapters.find(c => curSlug === c.slug || curSlug.startsWith(c.slug + '-'));
    const curIdx = curChapter ? cfg.chapters.indexOf(curChapter) : -1;

    if (e.key === 'n' || e.key === 'ArrowRight') {
      e.preventDefault();
      if (curIdx < cfg.chapters.length - 1) location.hash = cfg.chapters[curIdx + 1]?.slug || cfg.chapters[0].slug;
    } else if (e.key === 'p' || e.key === 'ArrowLeft') {
      e.preventDefault();
      if (curIdx > 0) location.hash = cfg.chapters[curIdx - 1].slug;
    } else if (e.key === 'h') {
      e.preventDefault();
      location.hash = '';
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } else if (e.key === 't') {
      toggleTheme();
    } else if (e.key === '+' || e.key === '=') {
      adjustFont(+1);
    } else if (e.key === '-' || e.key === '_') {
      adjustFont(-1);
    } else if (e.key === '?') {
      e.preventDefault();
      shortcutOverlay.classList.toggle('open');
    } else if (e.key === 'Escape') {
      if (mobileDrawer.classList.contains('open')) mobileDrawer.classList.remove('open');
    } else if (e.key === 'j' || e.key === 'ArrowDown' || e.key === 'k' || e.key === 'ArrowUp') {
      // Section scroll inside chapter
      if (!curChapter) return;
      e.preventDefault();
      const heads = Array.from(document.querySelectorAll('article.chapter-content h2'));
      if (!heads.length) return;
      const y = window.scrollY + 120;
      let target;
      if (e.key === 'j' || e.key === 'ArrowDown') {
        target = heads.find(h => h.getBoundingClientRect().top + window.scrollY > y + 20);
      } else {
        target = [...heads].reverse().find(h => h.getBoundingClientRect().top + window.scrollY < y - 20);
      }
      if (target) target.scrollIntoView({ behavior: 'smooth' });
    }
  });

  // ---------- Mobile drawer ----------
  mobileToggle.addEventListener('click', () => mobileDrawer.classList.toggle('open'));
  mobileDrawer.addEventListener('click', (e) => {
    if (e.target === mobileDrawer) mobileDrawer.classList.remove('open');
  });
  mobileChapters.addEventListener('click', (e) => {
    if (e.target.closest('a')) {
      setTimeout(() => mobileDrawer.classList.remove('open'), 120);
    }
  });

  // ---------- Boot ----------
  handleHashChange();
  updateProgress();
})();
