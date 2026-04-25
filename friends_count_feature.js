;(function () {
    function $(id) {
        return document.getElementById(id)
    }

    function hideProfileStatsBlocks() {
        const style = document.createElement("style")
        style.textContent = `
      .profile-stats { display: none !important; }
      #nav-friends-count {
        margin-left: 6px;
        font-size: 12px;
        font-weight: 400;
        color: var(--text3);
      }
    `
        document.head.appendChild(style)
    }

    function ensureFriendsCountLabel() {
        const navFriends = $("nav-friends")
        if (!navFriends) return null

        let label = $("nav-friends-count")
        if (label) return label

        const navLabel = navFriends.querySelector(".nav-label")
        if (!navLabel) return null

        label = document.createElement("span")
        label.id = "nav-friends-count"
        label.textContent = "0"
        navLabel.appendChild(label)
        return label
    }

    function setFriendsCount(count) {
        const label = ensureFriendsCountLabel()
        if (!label) return
        label.textContent = String(Math.max(0, Number(count) || 0))
    }

    async function refreshFriendsCount() {
        if (typeof window.api !== "function" || typeof window.loadSession !== "function") return
        const s = window.loadSession()
        if (!s?.token) return

        const res = await window.api("GET", "/friends", null, s.token)
        if (!res.ok) return

        const ids = res.data?.friend_ids || []
        setFriendsCount(ids.length)
    }

    function patchHooks() {
        if (typeof window.loadProfileStats === "function" && !window.loadProfileStats.__friendsCountPatched) {
            const original = window.loadProfileStats
            window.loadProfileStats = async function patchedLoadProfileStats(...args) {
                await original.apply(this, args)
                await refreshFriendsCount()
            }
            window.loadProfileStats.__friendsCountPatched = true
        }

        if (typeof window.loadFriends === "function" && !window.loadFriends.__friendsCountPatched) {
            const original = window.loadFriends
            window.loadFriends = async function patchedLoadFriends(...args) {
                await original.apply(this, args)
                await refreshFriendsCount()
            }
            window.loadFriends.__friendsCountPatched = true
        }
    }

    function init() {
        hideProfileStatsBlocks()
        ensureFriendsCountLabel()
        patchHooks()
        refreshFriendsCount()
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", init)
    } else {
        init()
    }
})()
