async function fetch_chats() {
    const res = await window.api.chat_list();
    return res
}

async function fetch_last_message(chatId) {
    const res = await window.api.chat_messages(chatId, limit=1, page=0);
    return res
}

function show_chats(chats) {
    const main = document.querySelector(".main");
    main.innerHTML = "";

    chats.forEach(chat => {
        const chatEl = document.createElement("div");
        chatEl.className = "chat";
        chatEl.dataset.chatId = chat.id;

        chatEl.innerHTML = `
            <div class="chat-title">${chat.name}</div>
            <div class="chat-last muted">Загрузка…</div>
        `;
        chatEl.addEventListener("click", () => {
            window.location.href = `/chat.html?chatId=${chat.id}`;
        });
        main.appendChild(chatEl);
    });
}

async function show_last_messages(chats) {
    await Promise.all(
        chats.map(async chat => {
            try {
                const messages = await fetch_last_message(chat.id);
                const last = messages.messages[0]?.text || "Новый чат";

                const chatEl = document.querySelector(
                    `.chat[data-chat-id="${chat.id}"] .chat-last`
                );

                if (chatEl) chatEl.textContent = last;
            } catch {

            }
        })
    );
}

(async () => {
    const chats = await fetch_chats();
    show_chats(chats);
    show_last_messages(chats);
})();