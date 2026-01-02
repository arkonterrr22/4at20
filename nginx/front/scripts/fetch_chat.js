async function fetch_chat() {
    try {
        const res = await window.api.chat_info(chatId);
        return res;
    } catch (e) {
        console.error("fetch_chat error:", e);
    }
}

async function show_chat() {
    res = await fetch_chat()
    const title = document.querySelector("#chat_title");
    title.innerText = res.chat.name;  
    };

show_chat()
