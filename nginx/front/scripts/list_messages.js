let displayedMessages = [];
const container = document.querySelector("#messages");
let page = 0;
let loading = false;
let buffer_size = 40;
async function fetch_messages(chatId, { before, after } = {}) {
    return window.api.chat_messages(chatId, {
        limit: 20,
        before,
        after,
    });
}

function render_messages() {
    container.innerHTML = "";
    displayedMessages.forEach(msg => {
        const div = document.createElement("div");
        div.className = msg.user_id === window.token.user_id ? "my_msg" : "foreign_msg";
        div.textContent = msg.text + ' ' + msg.id;
        container.appendChild(div);
    });
}

function scrollToBottom() {
    container.scrollTop = container.scrollHeight;
}

async function load_initial_messages() {
    const res = await fetch_messages(chatId);
    displayedMessages = res.messages
        .slice()
        .sort((a, b) => new Date(a.created_at) - new Date(b.created_at));

    render_messages();
    scrollToBottom();
}

function getOldest() {
    return displayedMessages[0];
}

function getNewest() {
    return displayedMessages[displayedMessages.length - 1];
}

function mergeMessages(buffer, incoming, direction, bufferSize = 40) {
    const map = new Map();

    for (const m of buffer) map.set(m.id, m);
    for (const m of incoming) map.set(m.id, m);

    const sorted = Array.from(map.values())
        .sort((a, b) => new Date(a.created_at) - new Date(b.created_at));

    // if (sorted.length <= bufferSize) return sorted;

    // if (direction === "older") {
    //     // грузили старые → режем новые
    //     return sorted.slice(0, bufferSize);
    // }

    // if (direction === "newer") {
    //     // грузили новые → режем старые
    //     return sorted.slice(-bufferSize);
    // }

    return sorted;
}


load_initial_messages();

container.addEventListener("scroll", async () => {
    if (loading) return;
    if (container.scrollTop <= 0) {
        const oldest = getOldest();
        if (!oldest) return;

        loading = true;
        const prevHeight = container.scrollHeight;

        const res = await fetch_messages(chatId, {
            before: oldest.created_at,
        });

        if (res.messages.length) {
            displayedMessages = mergeMessages(
                displayedMessages,
                res.messages,
                "older"
            );

            render_messages();
        }

        container.scrollTop = container.scrollHeight - prevHeight;
    }

    loading = false;
}
);

