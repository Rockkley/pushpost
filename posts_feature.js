/*
  PushPost posts frontend module
  - Adds "Посты" section dynamically into the existing app shell
  - Works with existing globals from current page: api, loadSession, showSection, esc, fmtDate, fmtTime
*/
(function () {
    const state = {
        activeTab: "feed", // feed | my
        feedCursor: "",
        myCursor: "",
        feedLoading: false,
        myLoading: false,
        creating: false,
        userCache: new Map(),
    };

    const POSTS_LIMIT = 15;

    function init() {
        if (typeof window === "undefined") return;
        if (!document.getElementById("screen-app")) return;

        injectStyles();
        mountNavItem();
        mountSection();
        patchShowSection();
        bindEvents();
    }

    function $(id) {
        return document.getElementById(id);
    }

    function injectStyles() {
        if ($("posts-feature-style")) return;

        const style = document.createElement("style");
        style.id = "posts-feature-style";
        style.textContent = `
      .posts-layout { display: flex; flex-direction: column; gap: 14px; }
      .post-composer { background: var(--bg2); border: 1px solid var(--border); border-radius: var(--radius); padding: 14px; }
      .post-composer textarea {
        width: 100%; min-height: 92px; resize: vertical; max-height: 260px;
        background: var(--bg3); border: 1px solid var(--border); border-radius: 10px;
        color: var(--text); padding: 12px 14px; font-family: var(--font); font-size: 14px; outline: none;
      }
      .post-composer textarea:focus { border-color: var(--accent); }
      .post-composer-actions { margin-top: 10px; display: flex; justify-content: space-between; align-items: center; gap: 8px; }
      .post-count { font-size: 12px; color: var(--text3); }

      .posts-tabs { display: flex; background: var(--bg2); border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden; }
      .posts-tab { flex: 1; border: none; background: transparent; color: var(--text3); padding: 11px 10px; cursor: pointer; font-family: var(--font); font-size: 13px; }
      .posts-tab.active { color: var(--accent); background: var(--accent-bg); font-weight: 600; }

      .posts-list { background: var(--bg2); border: 1px solid var(--border); border-radius: var(--radius); overflow: hidden; }
      .post-item { padding: 14px 16px; border-bottom: 1px solid var(--border); }
      .post-item:last-child { border-bottom: none; }
      .post-head { display: flex; justify-content: space-between; gap: 10px; align-items: center; margin-bottom: 8px; }
      .post-author { font-size: 13px; font-weight: 600; color: var(--text); }
      .post-time { font-size: 11px; color: var(--text3); }
      .post-content { font-size: 14px; line-height: 1.55; white-space: pre-wrap; word-break: break-word; }
      .post-foot { margin-top: 10px; display: flex; justify-content: flex-end; }
      .post-empty { padding: 28px 20px; text-align: center; color: var(--text3); }
      .post-more-wrap { padding: 12px; border-top: 1px solid var(--border); display: flex; justify-content: center; }
      .post-error { margin-top: 10px; font-size: 12px; color: var(--danger); }
    `;
        document.head.appendChild(style);
    }

    function mountNavItem() {
        if ($("nav-posts")) return;

        const nav = document.querySelector(".sidebar-nav");
        if (!nav) return;

        const btn = document.createElement("button");
        btn.className = "nav-item";
        btn.id = "nav-posts";
        btn.innerHTML = `<span class="nav-icon">📝</span><span class="nav-label">Посты</span>`;
        btn.onclick = () => window.showSection && window.showSection("posts");

        const searchBtn = $("nav-search");
        if (searchBtn && searchBtn.parentNode === nav) nav.insertBefore(btn, searchBtn);
        else nav.appendChild(btn);
    }

    function mountSection() {
        if ($("section-posts")) return;

        const main = document.querySelector(".main-content");
        if (!main) return;

        const section = document.createElement("div");
        section.className = "section";
        section.id = "section-posts";
        section.innerHTML = `
      <div class="page-header">
        <div class="page-title">Посты</div>
        <button class="page-action" id="posts-refresh-btn">↻ Обновить</button>
      </div>

      <div class="posts-layout">
        <div class="post-composer">
          <textarea id="post-compose-input" maxlength="5000" placeholder="Что у тебя нового?"></textarea>
          <div class="post-composer-actions">
            <div class="post-count"><span id="post-compose-count">0</span>/5000</div>
            <button class="btn btn-accent" id="post-compose-send">Опубликовать</button>
          </div>
          <div class="post-error" id="post-compose-error" style="display:none"></div>
        </div>

        <div class="posts-tabs">
          <button class="posts-tab active" id="posts-tab-feed">Лента друзей</button>
          <button class="posts-tab" id="posts-tab-my">Мои посты</button>
        </div>

        <div class="posts-list">
          <div id="posts-body"><div class="post-empty"><span class="spinner"></span></div></div>
          <div class="post-more-wrap" id="posts-more-wrap" style="display:none">
            <button class="btn btn-ghost" id="posts-more-btn">Показать ещё</button>
          </div>
        </div>
      </div>
    `;

        const settingsSection = $("section-settings");
        if (settingsSection && settingsSection.parentNode === main) main.insertBefore(section, settingsSection);
        else main.appendChild(section);
    }

    function patchShowSection() {
        if (typeof window.showSection !== "function" || window.__postsShowSectionPatched) return;

        const original = window.showSection;
        window.showSection = function patchedShowSection(name) {
            original(name);
            if (name === "posts") {
                refreshCurrentTab(true);
            }
        };

        window.__postsShowSectionPatched = true;
    }

    function bindEvents() {
        const input = $("post-compose-input");
        const sendBtn = $("post-compose-send");
        const refreshBtn = $("posts-refresh-btn");
        const tabFeed = $("posts-tab-feed");
        const tabMy = $("posts-tab-my");
        const moreBtn = $("posts-more-btn");

        if (!input || !sendBtn || !refreshBtn || !tabFeed || !tabMy || !moreBtn) return;

        input.addEventListener("input", () => {
            const len = input.value.length;
            $("post-compose-count").textContent = String(len);
            hideComposeError();
        });

        sendBtn.addEventListener("click", createPost);
        refreshBtn.addEventListener("click", () => refreshCurrentTab(true));

        tabFeed.addEventListener("click", () => switchTab("feed"));
        tabMy.addEventListener("click", () => switchTab("my"));

        moreBtn.addEventListener("click", () => {
            if (state.activeTab === "feed") loadFeed(false);
            else loadMyPosts(false);
        });

        document.addEventListener("click", async (e) => {
            const btn = e.target.closest("[data-post-delete]");
            if (!btn) return;
            const postID = btn.getAttribute("data-post-delete");
            if (!postID) return;
            await deletePost(postID, btn);
        });
    }

    function switchTab(tab) {
        state.activeTab = tab;

        $("posts-tab-feed")?.classList.toggle("active", tab === "feed");
        $("posts-tab-my")?.classList.toggle("active", tab === "my");

        refreshCurrentTab(true);
    }

    function refreshCurrentTab(reset) {
        if (state.activeTab === "feed") loadFeed(reset);
        else loadMyPosts(reset);
    }

    async function createPost() {
        if (state.creating) return;

        const session = window.loadSession && window.loadSession();
        if (!session?.token) {
            showComposeError("Сессия истекла, войди снова");
            return;
        }

        const input = $("post-compose-input");
        const content = (input?.value || "").trim();

        if (!content) {
            showComposeError("Текст поста не может быть пустым");
            return;
        }

        if (content.length > 5000) {
            showComposeError("Пост не должен превышать 5000 символов");
            return;
        }

        state.creating = true;
        const sendBtn = $("post-compose-send");
        if (sendBtn) {
            sendBtn.disabled = true;
            sendBtn.textContent = "…";
        }

        const res = await window.api("POST", "/posts/", { content }, session.token);

        state.creating = false;
        if (sendBtn) {
            sendBtn.disabled = false;
            sendBtn.textContent = "Опубликовать";
        }

        if (!res.ok) {
            showComposeError(formatPostsError(res.data));
            return;
        }

        if (input) {
            input.value = "";
            $("post-compose-count").textContent = "0";
        }
        hideComposeError();

        // Обновляем обе вкладки, чтобы пользователь сразу видел новый пост.
        state.feedCursor = "";
        state.myCursor = "";
        refreshCurrentTab(true);
    }

    async function deletePost(postID, btn) {
        const session = window.loadSession && window.loadSession();
        if (!session?.token) return;

        if (!window.confirm("Удалить этот пост?")) return;

        btn.disabled = true;
        const old = btn.textContent;
        btn.textContent = "…";

        const res = await window.api("DELETE", `/posts/${encodeURIComponent(postID)}`, null, session.token);

        btn.disabled = false;
        btn.textContent = old;

        if (!res.ok) {
            window.alert(formatPostsError(res.data));
            return;
        }

        state.feedCursor = "";
        state.myCursor = "";
        refreshCurrentTab(true);
    }

    async function loadFeed(reset) {
        if (state.feedLoading) return;

        const session = window.loadSession && window.loadSession();
        if (!session?.token) return;

        if (reset) state.feedCursor = "";

        state.feedLoading = true;
        setMoreLoading(true);

        const query = new URLSearchParams({ limit: String(POSTS_LIMIT) });
        if (state.feedCursor) query.set("cursor", state.feedCursor);

        const res = await window.api("GET", `/posts/feed?${query.toString()}`, null, session.token);

        state.feedLoading = false;
        setMoreLoading(false);

        if (!res.ok) {
            renderPostsError("Не удалось загрузить ленту");
            return;
        }

        const posts = res.data?.posts || [];
        const nextCursor = res.data?.next_cursor || "";

        if (reset) renderPostsList(posts, true);
        else appendPosts(posts);

        state.feedCursor = nextCursor;
        toggleMore(Boolean(nextCursor));
    }

    async function loadMyPosts(reset) {
        if (state.myLoading) return;

        const session = window.loadSession && window.loadSession();
        if (!session?.token || !session.userId) {
            renderPostsError("Сначала загрузи профиль пользователя");
            return;
        }

        if (reset) state.myCursor = "";

        state.myLoading = true;
        setMoreLoading(true);

        const query = new URLSearchParams({ limit: String(POSTS_LIMIT) });
        if (state.myCursor) query.set("cursor", state.myCursor);

        const path = `/posts/by-user/${encodeURIComponent(session.userId)}?${query.toString()}`;
        const res = await window.api("GET", path, null, session.token);

        state.myLoading = false;
        setMoreLoading(false);

        if (!res.ok) {
            renderPostsError("Не удалось загрузить твои посты");
            return;
        }

        const posts = res.data?.posts || [];
        const nextCursor = res.data?.next_cursor || "";

        if (reset) renderPostsList(posts, true);
        else appendPosts(posts);

        state.myCursor = nextCursor;
        toggleMore(Boolean(nextCursor));
    }

    function renderPostsError(message) {
        const body = $("posts-body");
        if (!body) return;
        body.innerHTML = `<div class="post-empty">${escapeHtml(message)}</div>`;
        toggleMore(false);
    }

    async function renderPostsList(posts, replace) {
        const body = $("posts-body");
        if (!body) return;

        if (!posts || posts.length === 0) {
            body.innerHTML = `<div class="post-empty">${state.activeTab === "feed" ? "Лента пока пустая" : "У тебя пока нет постов"}</div>`;
            return;
        }

        const html = await buildPostsHtml(posts);
        body.innerHTML = replace ? html : body.innerHTML + html;
    }

    async function appendPosts(posts) {
        if (!posts || posts.length === 0) return;
        const body = $("posts-body");
        if (!body) return;

        const html = await buildPostsHtml(posts);

        const empty = body.querySelector(".post-empty");
        if (empty) body.innerHTML = html;
        else body.insertAdjacentHTML("beforeend", html);
    }

    async function buildPostsHtml(posts) {
        const session = window.loadSession && window.loadSession();
        const myId = session?.userId;

        const parts = [];
        for (const post of posts) {
            const authorID = post.author_id || post.authorId || post.AuthorID || "";
            const authorName = await resolveAuthorName(authorID);
            const postID = post.id || post.ID || post.Id || "";
            const content = post.content || post.Content || "";
            const createdAt = post.created_at || post.createdAt || post.CreatedAt || "";
            const canDelete = Boolean(myId && authorID && myId === authorID);

            parts.push(`
        <article class="post-item" data-post-id="${escapeHtml(postID)}">
          <div class="post-head">
            <div class="post-author">${escapeHtml(authorName)}</div>
            <div class="post-time">${formatDateTime(createdAt)}</div>
          </div>
          <div class="post-content">${escapeHtml(content)}</div>
          ${canDelete ? `<div class="post-foot"><button class="btn btn-danger btn-sm" data-post-delete="${escapeHtml(postID)}">Удалить</button></div>` : ""}
        </article>
      `);
        }

        return parts.join("");
    }

    async function resolveAuthorName(userID) {
        if (!userID) return "Неизвестный";

        if (state.userCache.has(userID)) return state.userCache.get(userID);

        const session = window.loadSession && window.loadSession();
        if (!session?.token) return shortID(userID);

        const res = await window.api("GET", `/users/${encodeURIComponent(userID)}`, null, session.token);
        if (res.ok && res.data) {
            const name = res.data.username || res.data.Username || shortID(userID);
            state.userCache.set(userID, name);
            return name;
        }

        return shortID(userID);
    }

    function shortID(id) {
        return String(id).slice(0, 8) + "…";
    }

    function toggleMore(show) {
        const wrap = $("posts-more-wrap");
        if (!wrap) return;
        wrap.style.display = show ? "flex" : "none";
    }

    function setMoreLoading(isLoading) {
        const btn = $("posts-more-btn");
        if (!btn) return;
        btn.disabled = isLoading;
        btn.textContent = isLoading ? "…" : "Показать ещё";
    }

    function showComposeError(message) {
        const err = $("post-compose-error");
        if (!err) return;
        err.style.display = "block";
        err.textContent = message;
    }

    function hideComposeError() {
        const err = $("post-compose-error");
        if (!err) return;
        err.style.display = "none";
        err.textContent = "";
    }

    function formatDateTime(iso) {
        if (!iso) return "";
        try {
            const d = new Date(iso);
            return `${window.fmtDate ? window.fmtDate(d.toISOString()) : d.toLocaleDateString("ru-RU")} ${window.fmtTime ? window.fmtTime(d.toISOString()) : d.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" })}`;
        } catch {
            return "";
        }
    }

    function formatPostsError(data) {
        const code = data?.code;
        const map = {
            content_empty: "Пост не может быть пустым",
            content_too_long: "Пост слишком длинный",
            unauthorized: "Сессия истекла, войди снова",
            field_invalid: "Некорректные данные",
            validation_failed: "Ошибка валидации",
            post_not_found: "Пост не найден",
            not_post_author: "Ты не автор этого поста",
        };

        if (code && map[code]) return map[code];
        if (data?.fields) return Object.values(data.fields).join(", ");
        return code || "Не удалось выполнить действие с постом";
    }

    function escapeHtml(value) {
        const str = String(value ?? "");
        if (typeof window.esc === "function") return window.esc(str);
        return str.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
    }

    init();
})();
