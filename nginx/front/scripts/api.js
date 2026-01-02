window.api = (function () {
    const API_BASE = "/api";

    async function request(path, options = {}) {
        const token = localStorage.getItem("token");

        const res = await fetch(API_BASE + path, {
            ...options,
            headers: {
                "Content-Type": "application/json",
                ...(token && { Authorization: "Bearer " + token }),
                ...(options.headers || {})
            }
        });
        console.log(res.status)
        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || res.status);
        }

        return res.json();
    }

    return {
        login(login, password) {
            return request("/auth/login", {
                method: "POST",
                body: JSON.stringify({ login, password })
            });
        },

        register(username, login, password) {
            return request("/auth/register", {
                method: "POST",
                body: JSON.stringify({ username, login, password })
            });
        },
        group_users(groupId) {
            return request(
                `/auth/group/${groupId}/users`,
                { method: "GET" }
            );
        },
        create_chat(name, pic) {
            return request("/chat/create", {
                method: "POST",
                body: JSON.stringify({ name, pic })
            });
        },
        add_chat_members(chatid, members) {
            return request(`/chat/${chatid}/members/add`, {
                method: "POST",
                body: JSON.stringify({ members })
            });
        },
        chat_list() {
            return request("/chat/list",
                { method: "GET" }
            );
        },
        chat_messages(chatId, { limit = 20, before, after } = {}) {
            const params = new URLSearchParams();

            if (limit != null) params.set("limit", limit);
            if (before != null) params.set("before", before);
            if (after != null) params.set("after", after);

            return request(
                `/chat/${chatId}/messages?${params.toString()}`,
                { method: "GET" }
            );
        },
        chat_info(chatId) {
            return request(`/chat/${chatId}/info`, {
                method: "GET"
            });
        },
        send_message(chatId, text) {
            return request(`/chat/${chatId}/messages/send`, {
                method: "POST",
                body: JSON.stringify({
                    text: text,
                    content: ""
                })
            });
        }
    };
})();